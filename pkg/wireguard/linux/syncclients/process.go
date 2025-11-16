//go:build linux

package processpeerconfig

import (
	"fmt"
	"log"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	wireguardlinux "github.com/in4it/wireguard-server/pkg/wireguard/linux"
)

func SyncClients(storage storage.Iface, peerConfig wireguard.PeerConfig) error {
	err := processPeerConfig(storage, peerConfig)
	if err != nil {
		return fmt.Errorf("could not process peerconfig (%s): %s", peerConfig.ID, err)
	}
	return nil
}

func SyncClientsAndCleanup(storage storage.Iface, peerConfig wireguard.PeerConfig) {
	if err := SyncClients(storage, peerConfig); err != nil {
		returnErrorInGoRoutine(err)
		return
	}
	if err := Cleanup(storage); err != nil {
		returnErrorInGoRoutine(err)
		return
	}
}

func DeleteClient(peerConfig wireguard.PeerConfig) {
	err := processDeleteOfPeerConfig(peerConfig)
	if err != nil {
		returnErrorInGoRoutine(fmt.Errorf("could not process delete of peerconfig (%s): %s", peerConfig.ID, err))
		return
	}
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
