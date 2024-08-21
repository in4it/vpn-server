package wireguard

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"path"
	"strings"

	"github.com/in4it/wireguard-server/pkg/storage"
)

func getNextFreeIP(storage storage.Iface, addressRange netip.Prefix, addressPrefix string) (net.IP, error) {
	ipList := []string{}
	startIP, addressRangeParsed, err := net.ParseCIDR(addressRange.String())
	if err != nil {
		return nil, fmt.Errorf("cannot parse address range: %s: %s", addressRange, err)
	}

	clients, err := storage.ReadDir(storage.ConfigPath(VPN_CLIENTS_DIR))
	if err != nil {
		return nil, fmt.Errorf("cannot list files in users clients directory: %s", err)
	}
	for _, clientFilename := range clients {
		var peerConfig PeerConfig
		clientFilenameBytes, err := storage.ReadFile(storage.ConfigPath(path.Join(VPN_CLIENTS_DIR, clientFilename)))
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %s", clientFilename, err)
		}
		err = json.Unmarshal(clientFilenameBytes, &peerConfig)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal %s: %s", clientFilename, err)
		}
		ipList = append(ipList, peerConfig.Address)
	}

	newIP, err := getNextFreeIPFromList(startIP, addressRangeParsed, ipList, addressPrefix)
	if err != nil {
		return nil, fmt.Errorf("getNextFreeIPFromList error: %s", err)
	}

	return newIP, nil
}
func getNextFreeIPFromList(startIP net.IP, addressRange *net.IPNet, ipList []string, addressPrefix string) (net.IP, error) {
	nextIPAddress := startIP
	for i := 0; i < 100000; i++ {
		nextIPAddress = nextIP(nextIPAddress, 1)
		ipExists := false
		for _, ip := range ipList {
			ipRange := ip
			if !strings.Contains(ip, "/") {
				ipRange += addressPrefix
			}
			_, ipRangeParsed, err := net.ParseCIDR(ipRange)
			if err != nil {
				return nil, fmt.Errorf("cannot parse IP address: %s (ip range %s)", ip, ipRange)
			}
			if ipRangeParsed.Contains(nextIPAddress) {
				ipExists = true
			}
		}
		if !ipExists {
			if !addressRange.Contains(nextIPAddress) {
				return nil, fmt.Errorf("next IP (%s) is not within address range (%s). Address Range might be too small", nextIPAddress.String(), addressRange.String())
			}
			_, ipRangeParsed, err := net.ParseCIDR(nextIPAddress.String() + addressPrefix)
			if err != nil {
				return nil, fmt.Errorf("cannot parse new IP address range: %s: %s", nextIPAddress.String()+addressPrefix, err)
			}
			if !ipRangeParsed.Contains(startIP) { // don't pick a range where the start ip is in the range
				return nextIPAddress, nil
			}
		}
	}

	return nil, fmt.Errorf("couldn't determine next ip address")
}

func nextIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3)
}
