//go:build linux
// +build linux

package processpeerconfig

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	wireguardlinux "github.com/in4it/wireguard-server/pkg/wireguard/linux"
)

func Cleanup(storage storage.Iface) error {
	clients, err := storage.ReadDir(storage.ConfigPath(wireguard.VPN_CLIENTS_DIR))
	if err != nil {
		return fmt.Errorf("cannot list files in users clients directory: %s", err)
	}

	pubKeys := []string{}
	for _, clientFilename := range clients {
		var peerConfig wireguard.PeerConfig
		clientFilenameBytes, err := storage.ReadFile(storage.ConfigPath(path.Join(wireguard.VPN_CLIENTS_DIR, clientFilename)))
		if err != nil {
			return fmt.Errorf("cannot read %s: %s", clientFilename, err)
		}
		err = json.Unmarshal(clientFilenameBytes, &peerConfig)
		if err != nil {
			return fmt.Errorf("cannot unmarshal %s: %s", clientFilename, err)
		}
		if !peerConfig.Disabled {
			pubKeys = append(pubKeys, peerConfig.PublicKey)
		}
	}

	c, available, err := wireguardlinux.New()
	if err != nil {
		return fmt.Errorf("cannot start wireguardlinux client: %s", err)
	}
	if !available {
		return fmt.Errorf("wireguard linux client not available")
	}
	device, err := c.Device(wireguard.VPN_INTERFACE_NAME)
	if err != nil {
		return fmt.Errorf("wireguard linux device 'vpn' not found: %s", err)
	}

	for _, peer := range device.Peers {
		found := false
		for _, pubKey := range pubKeys {
			if peer.PublicKey.String() == pubKey {
				found = true
			}
		}
		if !found {
			err = wgDeletePeer(wireguard.PeerConfig{PublicKey: peer.PublicKey.String()})
			if err != nil {
				return fmt.Errorf("wgDeletePeer error: %s", err)
			}
		}
	}

	return nil
}
