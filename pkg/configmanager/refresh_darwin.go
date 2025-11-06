//go:build darwin

package configmanager

import (
	"errors"
	"fmt"
	"os"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func refreshAllClientsAndServer(storage storage.Iface, clientCache *wireguard.ClientCache) error {
	peerConfigPath := storage.ConfigPath(wireguard.VPN_CLIENTS_DIR)

	if _, err := os.Stat(peerConfigPath); errors.Is(err, os.ErrNotExist) {
		return nil // directory doesn't exist, so no configs to be read
	}

	entries, err := storage.ReadDir(peerConfigPath)
	if err != nil {
		return fmt.Errorf("can not list clients from dir %s: %s", peerConfigPath, err)
	}

	for _, filename := range entries {
		peerConfig, err := wireguard.GetPeerConfigByFilename(storage, filename)
		if err != nil {
			return fmt.Errorf("getClientFile error: %s", err)
		}
		err = wireguard.UpdateClientCache(peerConfig, clientCache)
		if err != nil {
			return fmt.Errorf("update client cache error: %s", err)
		}
	}
	fmt.Printf("Warning: not refreshAllClients supported on darwin\n")
	return nil
}

func syncClient(storage storage.Iface, filename string, clientCache *wireguard.ClientCache) error {
	peerConfig, err := wireguard.GetPeerConfigByFilename(storage, filename)
	if err != nil {
		return fmt.Errorf("getClientFile error: %s", err)
	}
	err = wireguard.UpdateClientCache(peerConfig, clientCache)
	if err != nil {
		return fmt.Errorf("update client cache error: %s", err)
	}
	fmt.Printf("Warning: syncClient not supported on darwin. Cannot sync: %s\n", filename)
	return nil
}

func cleanupClients(storage storage.Iface) error {
	fmt.Printf("Warning: cleanupClients() not supported on darwin.")
	return nil
}

func deleteClient(storage storage.Iface, filename string, clientCache *wireguard.ClientCache) error {
	peerConfig, err := wireguard.GetPeerConfigByFilename(storage, filename)
	if err != nil {
		return fmt.Errorf("getClientFile error: %s", err)
	}
	err = wireguard.UpdateClientCache(peerConfig, clientCache)
	if err != nil {
		return fmt.Errorf("update client cache error: %s", err)
	}
	fmt.Printf("Warning: deleteClient not supported on darwin. Cannot delete: %s\n", filename)
	return nil
}
