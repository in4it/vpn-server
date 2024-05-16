//go:build linux
// +build linux

package processpeerconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"path"

	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	wireguardlinux "github.com/in4it/wireguard-server/pkg/wireguard/linux"
)

func SyncClients(storage storage.Iface, filename string) error {
	peerConfig, peerConfigFilename, err := getClientFile(storage, filename)
	if err != nil {
		return fmt.Errorf("getClientFile error: %s", err)
	}
	err = processPeerConfig(storage, peerConfig)
	if err != nil {
		return fmt.Errorf("could not process peerconfig (%s): %s", peerConfigFilename, err)
	}
	return nil
}

func SyncClientsAndCleanup(storage storage.Iface, filename string) {
	if err := SyncClients(storage, filename); err != nil {
		returnErrorInGoRoutine(err)
		return
	}
	if err := Cleanup(storage); err != nil {
		returnErrorInGoRoutine(err)
		return
	}
}

func DeleteClient(storage storage.Iface, filename string) {
	peerConfig, peerConfigFilename, err := getClientFile(storage, filename)
	if err != nil {
		returnErrorInGoRoutine(fmt.Errorf("getClientFile error: %s", err))
		return
	}
	err = processDeleteOfPeerConfig(peerConfig)
	if err != nil {
		returnErrorInGoRoutine(fmt.Errorf("could not process delete of peerconfig (%s): %s", peerConfigFilename, err))
		return
	}
}

func getClientFile(storage storage.Iface, filename string) (wireguard.PeerConfig, string, error) {
	var peerConfig wireguard.PeerConfig

	peerConfigFilename := storage.ConfigPath(path.Join(wireguard.VPN_CLIENTS_DIR, filename))
	peerConfigData, err := storage.ReadFile(peerConfigFilename)
	if err != nil {
		return peerConfig, peerConfigFilename, fmt.Errorf("could not read clients filename (%s): %s", peerConfigFilename, err)
	}
	err = json.Unmarshal(peerConfigData, &peerConfig)
	if err != nil {
		return peerConfig, peerConfigFilename, fmt.Errorf("could not read unmarshal peerconfig(%s): %s", peerConfigFilename, err)
	}
	return peerConfig, peerConfigFilename, nil
}

func processPeerConfig(storage storage.Iface, peerConfig wireguard.PeerConfig) error {
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

	found := false
	for _, peer := range device.Peers {
		if !peerConfig.Disabled && peer.PublicKey.String() == peerConfig.PublicKey {
			found = true
		}
	}

	if !found { // add peer
		if peerConfig.PublicKey == "" {
			log.Printf("Warning: corrupt peer config. Skipping peer (id: %s)", peerConfig.ID)
		} else {
			err = wgAddPeer(storage, peerConfig)
			if err != nil {
				return fmt.Errorf("wgAddPeer error: %s", err)
			}
		}
	}

	return nil
}

func processDeleteOfPeerConfig(peerConfig wireguard.PeerConfig) error {
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

	found := false
	for _, peer := range device.Peers {
		if peer.PublicKey.String() == peerConfig.PublicKey {
			found = true
		}
	}

	if found { // delete peer
		err = wgDeletePeer(peerConfig)
		if err != nil {
			return fmt.Errorf("wgDeletePeer error: %s", err)
		}
	}

	return nil
}

func returnErrorInGoRoutine(err error) {
	log.Println(err)
}
