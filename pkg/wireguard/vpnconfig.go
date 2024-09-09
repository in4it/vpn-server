package wireguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os/user"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard/network"
)

var vpnConfigMutex sync.Mutex

func GetVPNConfig(storage storage.Iface) (VPNConfig, error) {
	var vpnConfig VPNConfig

	filename := storage.ConfigPath(VPN_CONFIG_NAME)

	// check if config exists
	if !storage.FileExists(filename) {
		vpnConfig, err := getEmptyVPNConfig()
		if err != nil {
			return vpnConfig, fmt.Errorf("could not generate empty VPN Config: %s", err)
		}
		return vpnConfig, nil
	}

	data, err := storage.ReadFile(filename)
	if err != nil {
		return vpnConfig, fmt.Errorf("config read error: %s", err)
	}
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	err = decoder.Decode(&vpnConfig)
	if err != nil {
		return vpnConfig, fmt.Errorf("decode input error: %s", err)
	}

	if vpnConfig.PacketLogsTypes == nil {
		vpnConfig.PacketLogsTypes = make(map[string]bool)
	}

	return vpnConfig, nil
}

func getEmptyVPNConfig() (VPNConfig, error) {
	vpnConfig := VPNConfig{
		PacketLogsTypes: make(map[string]bool),
	}
	return vpnConfig, nil
}

func CreateNewVPNConfig(storage storage.Iface) (VPNConfig, error) {
	vpnConfig, err := getEmptyVPNConfig()
	if err != nil {
		return vpnConfig, fmt.Errorf("get empty vpn config error: %s", err)
	}
	prefix, err := netip.ParsePrefix(DEFAULT_VPN_PREFIX)
	if err != nil {
		return vpnConfig, fmt.Errorf("ParsePrefix error: %s", err)
	}
	vpnConfig.AddressRange = prefix

	privateKey, publicKey, err := GenerateKeys()
	if err != nil {
		return vpnConfig, fmt.Errorf("GenerateKeys error: %s", err)
	}
	vpnConfig.PublicKey = publicKey

	vpnConfig.ClientRoutes = []string{"0.0.0.0/0", "::/0"}

	err = storage.EnsurePath(VPN_SERVER_SECRETS_PATH)
	if err != nil {
		return vpnConfig, fmt.Errorf("could not ensure path exists %s: %s", VPN_SERVER_SECRETS_PATH, err)
	}
	err = storage.WriteFile(path.Join(VPN_SERVER_SECRETS_PATH, VPN_PRIVATE_KEY_FILENAME), []byte(privateKey))
	if err != nil {
		return vpnConfig, fmt.Errorf("could not write private key to %s: %s", path.Join(VPN_SERVER_SECRETS_PATH, VPN_PRIVATE_KEY_FILENAME), err)
	}
	privateKey = ""

	presharedKey, err := GeneratePresharedKey()
	if err != nil {
		return vpnConfig, fmt.Errorf("GeneratePresharedKey error: %s", err)
	}

	// write preshared key
	err = storage.WriteFile(path.Join(VPN_SERVER_SECRETS_PATH, PRESHARED_KEY_FILENAME), []byte(presharedKey))
	if err != nil {
		return vpnConfig, fmt.Errorf("could not write presharedkey key to %s: %s", path.Join(VPN_SERVER_SECRETS_PATH, VPN_PRIVATE_KEY_FILENAME), err)
	}

	vpnConfig.PresharedKey = presharedKey
	vpnConfig.Port = 51820
	vpnConfig.Endpoint = guessHostname()
	vpnConfig.ClientAddressPrefix = "/32"

	vpnConfig.ExternalInterface, err = network.GetInterfaceDefaultGw()
	if err != nil {
		log.Printf("Warning: unable to get network interface with default gateway: %s", err)
	}

	vpnConfig.Nameservers, err = network.GetNameservers()
	if err != nil {
		log.Printf("Warning: unable to get nameservers: %s", err)
	}

	err = WriteVPNConfig(storage, vpnConfig)
	if err != nil {
		return vpnConfig, fmt.Errorf("WriteVPNConfig error: %s", err)
	}

	return vpnConfig, nil
}

func WriteVPNConfig(storage storage.Iface, vpnConfig VPNConfig) error {
	vpnConfigMutex.Lock()
	defer vpnConfigMutex.Unlock()
	vpnConfigCopy := vpnConfig
	out, err := json.Marshal(vpnConfigCopy)
	if err != nil {
		return fmt.Errorf("vpn config marshal error: %s", err)
	}
	err = storage.WriteFile(storage.ConfigPath(VPN_CONFIG_NAME), out)
	if err != nil {
		return fmt.Errorf("vpn config write error: %s", err)
	}
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %s", err)
	}
	if currentUser.Username != "vpn" {
		err = storage.EnsureOwnership(storage.ConfigPath(VPN_CONFIG_NAME), "vpn")
		if err != nil {
			return fmt.Errorf("config write error: %s", err)
		}
	}

	// notify configmanager
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post("http://"+CONFIGMANAGER_URI+"/refresh-server-config", "application/json", nil)
	if err != nil {
		return fmt.Errorf("configmanager post error: %s", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("configmanager post error: received status code %d", resp.StatusCode)
	}

	return nil
}

func guessHostname() string {
	// try to get hostname from local loopback
	hostname := ""
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err == nil {
		defer conn.Close()
		hostname = conn.LocalAddr().(*net.UDPAddr).String()
	}

	hostnameSplit := strings.Split(hostname, ":")
	hostname = hostnameSplit[0]

	// try to get assigned external IP on EC2
	tokenEndpoint := "http://169.254.169.254/latest/api/token"
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("PUT", tokenEndpoint, nil)
	if err != nil {
		return hostname
	}

	req.Header.Add("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	resp, err := client.Do(req)
	if err != nil {
		return hostname
	}

	token := ""

	if resp.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		token = string(bodyBytes)
	}

	endpoint := "http://169.254.169.254/latest/meta-data/public-ipv4"
	req, err = http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return hostname
	}
	if token != "" {
		req.Header.Add("X-aws-ec2-metadata-token", token)
	}

	resp, err = client.Do(req)
	if err != nil {
		return hostname
	}
	if resp.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return string(bodyBytes)
	}
	return hostname
}
