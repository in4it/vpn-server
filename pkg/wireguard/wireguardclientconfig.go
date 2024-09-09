package wireguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/in4it/wireguard-server/pkg/storage"
)

var clientConfigMutex sync.Mutex

func GetConfigNumbers(storage storage.Iface, userID string) ([]int, error) {
	configNumbers := []int{}
	// determine config number
	clients, err := storage.ReadDir(storage.ConfigPath(VPN_CLIENTS_DIR))
	if err != nil {
		return configNumbers, fmt.Errorf("cannot list connections for user: %s", err)
	}

	for _, clientFilename := range clients {
		if HasClientUserID(clientFilename, userID) {
			configNumber, err := getConfigNumberFromConnectionFile(clientFilename)
			if err != nil {
				return configNumbers, fmt.Errorf("cannot find config number from file (%s): %s", clientFilename, err)
			}
			configNumbers = append(configNumbers, configNumber)
		}
	}
	return configNumbers, nil
}

func NewEmptyClientConfig(storage storage.Iface, userID string) (PeerConfig, error) {
	clientConfigMutex.Lock()
	defer clientConfigMutex.Unlock()

	vpnConfig, err := GetVPNConfig(storage)
	if err != nil {
		return PeerConfig{}, fmt.Errorf("failed to get vpn config: %s", err)
	}

	// get next IP address, write in client file
	nextFreeIP, err := getNextFreeIP(storage, vpnConfig.AddressRange, vpnConfig.ClientAddressPrefix)
	if err != nil {
		return PeerConfig{}, fmt.Errorf("getNextFreeIP error: %s", err)
	}

	// determine config number
	configNumbers, err := GetConfigNumbers(storage, userID)
	if err != nil {
		return PeerConfig{}, fmt.Errorf("GetConfigNumbers error: %s", err)
	}
	newConfigNumber := 1
	if len(configNumbers) > 0 {
		newConfigNumber = slices.Max(configNumbers) + 1
	}

	clientAllowedIPs, err := getClientAllowedIPs(vpnConfig.AddressRange.Addr().String()+vpnConfig.ClientAddressPrefix, vpnConfig.ClientRoutes, vpnConfig.Nameservers)
	if err != nil {
		return PeerConfig{}, fmt.Errorf("getClientAllowedIPs error: %s", err)
	}

	// validate address
	address := nextFreeIP.String() + vpnConfig.ClientAddressPrefix
	_, _, err = net.ParseCIDR(address)
	if err != nil {
		return PeerConfig{}, fmt.Errorf("cannot parse client address: %s", err)
	}

	peerConfig := PeerConfig{
		ID:               fmt.Sprintf("%s-%d", userID, newConfigNumber),
		DNS:              strings.Join(vpnConfig.Nameservers, ", "),
		Name:             fmt.Sprintf("connection-%d", newConfigNumber),
		Address:          address,
		ServerAllowedIPs: []string{address},
		ClientAllowedIPs: clientAllowedIPs,
	}

	// write peerconfig

	peerConfigOut, err := json.Marshal(peerConfig)
	if err != nil {
		return peerConfig, fmt.Errorf("peerConfig marshal error: %s", err)
	}
	userConfigFilename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, fmt.Sprintf("%s.json", peerConfig.ID)))
	err = storage.WriteFile(userConfigFilename, peerConfigOut)
	if err != nil {
		return peerConfig, fmt.Errorf("could not save vpn client info to file: %s", err)
	}

	return peerConfig, nil
}

func UpdateClientsConfig(storage storage.Iface) error {
	clientConfigMutex.Lock()
	defer clientConfigMutex.Unlock()
	vpnConfig, err := GetVPNConfig(storage)
	if err != nil {
		return fmt.Errorf("failed to get vpn config: %s", err)
	}
	clients, err := storage.ReadDir(storage.ConfigPath(VPN_CLIENTS_DIR))
	if err != nil {
		return fmt.Errorf("cannot list client connections: %s", err)
	}
	for _, clientFilename := range clients {
		peerConfig, err := getPeerConfig(storage, strings.TrimSuffix(clientFilename, ".json"))
		if err != nil {
			return fmt.Errorf("cannot get peer config (%s): %s", clientFilename, err)
		}
		// update attributes
		clientAllowedIPs, err := getClientAllowedIPs(vpnConfig.AddressRange.Addr().String()+vpnConfig.ClientAddressPrefix, vpnConfig.ClientRoutes, vpnConfig.Nameservers)
		if err != nil {
			return fmt.Errorf("getClientAllowedIPs error: %s", err)
		}

		rewriteFile := false
		if !slices.Equal(clientAllowedIPs, peerConfig.ClientAllowedIPs) {
			rewriteFile = true
			peerConfig.ClientAllowedIPs = clientAllowedIPs
		}
		if peerConfig.DNS != strings.Join(vpnConfig.Nameservers, ", ") {
			rewriteFile = true
			peerConfig.DNS = strings.Join(vpnConfig.Nameservers, ", ")
		}

		addressParsed, err := netip.ParsePrefix(peerConfig.Address)
		if err != nil {
			return fmt.Errorf("couldn't parse existing address of vpn config %s", clientFilename)
		}
		if !vpnConfig.AddressRange.Contains(addressParsed.Addr()) { // client IP address is not in address range (address range might have changed)
			nextFreeIP, err := getNextFreeIP(storage, vpnConfig.AddressRange, vpnConfig.ClientAddressPrefix)
			if err != nil {
				return fmt.Errorf("getNextFreeIP error: %s", err)
			}
			peerConfig.Address = nextFreeIP.String() + vpnConfig.ClientAddressPrefix
			peerConfig.ServerAllowedIPs = []string{nextFreeIP.String() + "/32"}
			rewriteFile = true
		}

		if !strings.HasSuffix(peerConfig.Address, vpnConfig.ClientAddressPrefix) {
			rewriteFile = true
			peerConfig.Address = addressParsed.Addr().String() + vpnConfig.ClientAddressPrefix
		}

		if rewriteFile {
			peerConfigOut, err := json.Marshal(peerConfig)
			if err != nil {
				return fmt.Errorf("peerConfig marshal error: %s", err)
			}
			userConfigFilename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, clientFilename))
			err = storage.WriteFile(userConfigFilename, peerConfigOut)
			if err != nil {
				return fmt.Errorf("could not save vpn client info to file (%s): %s", clientFilename, err)
			}
		}
	}
	return nil
}

func getPeerConfig(storage storage.Iface, connectionID string) (PeerConfig, error) {
	return GetPeerConfigByFilename(storage, fmt.Sprintf("%s.json", connectionID))
}

func GetPeerConfigByFilename(storage storage.Iface, filename string) (PeerConfig, error) {
	var peerConfig PeerConfig
	peerConfigFilename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, filename))
	peerConfigBytes, err := storage.ReadFile(peerConfigFilename)
	if err != nil {
		return peerConfig, fmt.Errorf("cannot read connection config: %s", err)
	}
	err = json.Unmarshal(peerConfigBytes, &peerConfig)
	if err != nil {
		return peerConfig, fmt.Errorf("cannot unmarshal peer config: %s", err)
	}
	return peerConfig, nil
}

func GetAllPeerConfigs(storage storage.Iface) ([]PeerConfig, error) {
	peerConfigPath := storage.ConfigPath(VPN_CLIENTS_DIR)

	entries, err := storage.ReadDir(peerConfigPath)
	if err != nil {
		return []PeerConfig{}, fmt.Errorf("can not list clients from dir %s: %s", peerConfigPath, err)
	}
	peerConfigs := make([]PeerConfig, len(entries))
	for k, entry := range entries {
		peerConfig, err := GetPeerConfigByFilename(storage, entry)
		if err != nil {
			return peerConfigs, fmt.Errorf("cnanot get peer config (%s): %s", entry, err)
		}
		peerConfigs[k] = peerConfig
	}
	return peerConfigs, nil
}

func GetClientTemplate(storage storage.Iface) ([]byte, error) {
	filename := storage.ConfigPath("templates/client.tmpl")
	err := storage.EnsurePath(storage.ConfigPath("templates"))
	if err != nil {
		return nil, fmt.Errorf("cannot ensure path templates for client.tmpl: %s", err)
	}
	if !storage.FileExists(filename) {
		err := storage.WriteFile(filename, []byte(DEFAULT_CLIENT_TEMPLATE))
		if err != nil {
			return nil, fmt.Errorf("could not create initial client template: %s", err)
		}
	}
	data, err := storage.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read client template: %s", err)
	}
	return data, err
}

func WriteClientTemplate(storage storage.Iface, body []byte) error {
	filename := storage.ConfigPath("templates/client.tmpl")
	err := storage.WriteFile(filename, body)
	if err != nil {
		return fmt.Errorf("could not write client template: %s", err)
	}
	return nil
}

func GenerateNewClientConfig(storage storage.Iface, connectionID, userID string) ([]byte, error) {
	clientConfigMutex.Lock()
	defer clientConfigMutex.Unlock()

	// parse template
	privateKey, publicKey, err := GenerateKeys()
	if err != nil {
		return nil, fmt.Errorf("GenerateKeys error: %s", err)
	}
	vpnConfig, err := GetVPNConfig(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to get vpn config: %s", err)
	}

	peerConfig, err := getPeerConfig(storage, connectionID)
	if err != nil {
		return nil, fmt.Errorf("could not get peer config: %s", err)
	}

	// set public key
	peerConfig.PublicKey = publicKey

	peerConfigOut, err := json.Marshal(peerConfig)
	if err != nil {
		return nil, fmt.Errorf("peerConfig marshal error: %s", err)
	}

	// config template
	vpnClientData := VPNClientData{
		Address:         peerConfig.Address,
		DNS:             peerConfig.DNS,
		PrivateKey:      privateKey,
		ServerPublicKey: vpnConfig.PublicKey,
		PresharedKey:    vpnConfig.PresharedKey,
		Endpoint:        vpnConfig.Endpoint + ":" + fmt.Sprintf("%d", vpnConfig.Port),
		AllowedIPs:      peerConfig.ClientAllowedIPs,
	}

	templatefileContents, err := GetClientTemplate(storage)
	if err != nil {
		return nil, fmt.Errorf("could not get client template: %s", err)
	}

	tmpl, err := template.New("client.tmpl").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(string(templatefileContents))
	if err != nil {
		return nil, fmt.Errorf("could not parse client template: %s", err)
	}
	out := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(out, vpnClientData)
	if err != nil {
		return nil, fmt.Errorf("could not parse client template (execute parsing): %s", err)
	}

	// save new public key for configmanager server to pick up
	userConfigFilename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, fmt.Sprintf("%s.json", peerConfig.ID)))
	err = storage.WriteFile(userConfigFilename, peerConfigOut)
	if err != nil {
		return nil, fmt.Errorf("could not save vpn client info to file: %s", err)
	}

	// notify configmanager
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	refreshClientRequest := RefreshClientRequest{
		Action:    ACTION_ADD,
		Filenames: []string{path.Base(userConfigFilename)},
	}
	payload, err := json.Marshal(refreshClientRequest)
	if err != nil {
		return nil, fmt.Errorf("could not marshal refresh client request: %s", err)
	}
	resp, err := client.Post("http://"+CONFIGMANAGER_URI+"/refresh-clients", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("configmanager post error: %s", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("configmanager post error: received status code %d", resp.StatusCode)
	}

	return out.Bytes(), nil
}
func DeleteAllClientConfigs(storage storage.Iface, userID string) error {
	clients, err := storage.ReadDir(storage.ConfigPath(VPN_CLIENTS_DIR))
	if err != nil {
		return fmt.Errorf("cannot list files in users clients directory: %s", err)
	}

	for _, clientFilename := range clients {
		if HasClientUserID(clientFilename, userID) {
			filename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, clientFilename))
			err = storage.Remove(filename)
			if err != nil {
				return fmt.Errorf("removal of file %s failed: %s", filename, err)
			}
		}
	}
	// notify configmanager
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	refreshClientRequest := RefreshClientRequest{
		Action:    ACTION_CLEANUP,
		Filenames: []string{},
	}
	payload, err := json.Marshal(refreshClientRequest)
	if err != nil {
		return fmt.Errorf("could not marshal refresh client request: %s", err)
	}
	resp, err := client.Post("http://"+CONFIGMANAGER_URI+"/refresh-clients", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("configmanager post error: %s", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("configmanager post error: received status code %d", resp.StatusCode)
	}
	return nil
}
func DeleteClientConfig(storage storage.Iface, connectionID, userID string) error {
	toDeleteFilename := fmt.Sprintf("%s.json", connectionID)
	filename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, toDeleteFilename))
	err := storage.Remove(filename)
	if err != nil {
		return fmt.Errorf("removal of file %s failed: %s", filename, err)
	}
	// notify configmanager
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	refreshClientRequest := RefreshClientRequest{
		Action:    ACTION_CLEANUP,
		Filenames: []string{},
	}
	payload, err := json.Marshal(refreshClientRequest)
	if err != nil {
		return fmt.Errorf("could not marshal refresh client request: %s", err)
	}
	resp, err := client.Post("http://"+CONFIGMANAGER_URI+"/refresh-clients", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("configmanager post error: %s", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("configmanager post error: received status code %d", resp.StatusCode)
	}
	return nil
}
func DisableAllClientConfigs(storage storage.Iface, userID string) error {
	clientConfigMutex.Lock()
	defer clientConfigMutex.Unlock()
	clients, err := storage.ReadDir(storage.ConfigPath(VPN_CLIENTS_DIR))
	if err != nil {
		return fmt.Errorf("cannot list files in users clients directory: %s", err)
	}

	toDelete := []string{}
	for _, clientFilename := range clients {
		if HasClientUserID(clientFilename, userID) {
			toDelete = append(toDelete, clientFilename)
		}
	}

	// set the disabled flag on each file
	for _, toDeleteFilename := range toDelete {
		var peerConfig PeerConfig
		filename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, toDeleteFilename))
		toDeleteFileContents, err := storage.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("can't read file %s: %s", filename, err)
		}
		err = json.Unmarshal(toDeleteFileContents, &peerConfig)
		if err != nil {
			return fmt.Errorf("can't unmarshal file %s: %s", filename, err)
		}
		peerConfig.Disabled = true
		toDeleteToWrite, err := json.Marshal(peerConfig)
		if err != nil {
			return fmt.Errorf("can't marshal peer config file %s: %s", filename, err)
		}
		err = storage.WriteFile(filename, toDeleteToWrite)
		if err != nil {
			return fmt.Errorf("can't write peer config file %s: %s", filename, err)
		}
	}

	// notify configmanager
	if len(toDelete) > 0 {
		client := http.Client{
			Timeout: 10 * time.Second,
		}
		refreshClientRequest := RefreshClientRequest{
			Action:    ACTION_DELETE,
			Filenames: toDelete,
		}
		payload, err := json.Marshal(refreshClientRequest)
		if err != nil {
			return fmt.Errorf("could not marshal refresh client request: %s", err)
		}
		resp, err := client.Post("http://"+CONFIGMANAGER_URI+"/refresh-clients", "application/json", bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("configmanager post error: %s", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("configmanager post error: received status code %d", resp.StatusCode)
		}
	}
	return nil
}
func ReactivateAllClientConfigs(storage storage.Iface, userID string) error {
	clientConfigMutex.Lock()
	defer clientConfigMutex.Unlock()
	clients, err := storage.ReadDir(storage.ConfigPath(VPN_CLIENTS_DIR))
	if err != nil {
		return fmt.Errorf("cannot list files in users clients directory: %s", err)
	}

	toAdd := []string{}
	for _, clientFilename := range clients {
		if HasClientUserID(clientFilename, userID) {
			toAdd = append(toAdd, clientFilename)
		}
	}

	// set the disabled flag on each file
	for _, toAddFilename := range toAdd {
		var peerConfig PeerConfig
		filename := storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, toAddFilename))
		toAddFileContents, err := storage.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("can't read file %s: %s", filename, err)
		}
		err = json.Unmarshal(toAddFileContents, &peerConfig)
		if err != nil {
			return fmt.Errorf("can't unmarshal file %s: %s", filename, err)
		}
		peerConfig.Disabled = false
		toAddToWrite, err := json.Marshal(peerConfig)
		if err != nil {
			return fmt.Errorf("can't marshal peer config file %s: %s", filename, err)
		}
		err = storage.WriteFile(filename, toAddToWrite)
		if err != nil {
			return fmt.Errorf("can't write peer config file %s: %s", filename, err)
		}
	}

	// notify configmanager
	if len(toAdd) > 0 {
		client := http.Client{
			Timeout: 10 * time.Second,
		}
		refreshClientRequest := RefreshClientRequest{
			Action:    ACTION_ADD,
			Filenames: toAdd,
		}
		payload, err := json.Marshal(refreshClientRequest)
		if err != nil {
			return fmt.Errorf("could not marshal refresh client request: %s", err)
		}
		resp, err := client.Post("http://"+CONFIGMANAGER_URI+"/refresh-clients", "application/json", bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("configmanager post error: %s", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("configmanager post error: received status code %d", resp.StatusCode)
		}
	}
	return nil
}

func HasClientUserID(filename string, userID string) bool {
	clientID, _, _ := getClientIDAndConfigID(strings.TrimSuffix(filename, ".json"))
	return clientID == userID
}

func getConfigNumberFromConnectionFile(filename string) (int, error) {
	_, configNumber, err := getClientIDAndConfigID(strings.TrimSuffix(filename, ".json"))
	return configNumber, err
}
func getClientIDAndConfigID(name string) (string, int, error) {
	nameSplit := strings.Split(name, "-")
	if len(nameSplit) < 2 {
		return "", -1, fmt.Errorf("invalid connection name")
	}
	i, err := strconv.Atoi(nameSplit[len(nameSplit)-1])
	if err != nil {
		return "", -1, fmt.Errorf("could not convert string to int: %s", err)
	}
	clientID := strings.Join(nameSplit[:len(nameSplit)-1], "-")
	return clientID, i, nil
}

func networkIntersects(network1, network2 *net.IPNet) bool {
	return network2.Contains(network1.IP) || network1.Contains(network2.IP)
}
