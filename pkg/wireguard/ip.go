package wireguard

import (
	"encoding/json"
	"fmt"
	"net"
	"path"

	"github.com/in4it/wireguard-server/pkg/storage"
)

func getNextFreeIP(storage storage.Iface, startIP net.IP) (net.IP, error) {
	ipList := []string{}

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
		peerConfigAddress, _, err := net.ParseCIDR(peerConfig.Address)
		if err != nil {
			return nil, fmt.Errorf("could not parse peer config address %s: %s", peerConfig.Address, err)
		}
		ipList = append(ipList, peerConfigAddress.String())
	}

	newIP, err := getNextFreeIPFromList(startIP, ipList)
	if err != nil {
		return nil, fmt.Errorf("getNextFreeIPFromList error: %s", err)
	}

	return newIP, nil
}
func getNextFreeIPFromList(startIP net.IP, ipList []string) (net.IP, error) {
	nextIPAddress := startIP
	for i := 0; i < 100000; i++ {
		nextIPAddress = nextIP(nextIPAddress, 1)
		ipExists := false
		for _, ip := range ipList {
			if nextIPAddress.String() == ip {
				ipExists = true
			}
		}
		if !ipExists {
			return nextIPAddress, nil
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
