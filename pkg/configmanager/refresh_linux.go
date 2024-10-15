//go:build linux
// +build linux

package configmanager

import (
	"errors"
	"fmt"
	"os"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	syncclients "github.com/in4it/wireguard-server/pkg/wireguard/linux/syncclients"
)

func syncClient(storage storage.Iface, filename string, clientCache *wireguard.ClientCache) error {
	peerConfig, err := wireguard.GetPeerConfigByFilename(storage, filename)
	if err != nil {
		return fmt.Errorf("getClientFile error: %s", err)
	}
	err = wireguard.UpdateClientCache(peerConfig, clientCache)
	if err != nil {
		return fmt.Errorf("update client cache error: %s", err)
	}
	go syncclients.SyncClientsAndCleanup(storage, peerConfig)
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
	go syncclients.DeleteClient(peerConfig)
	return nil
}
func cleanupClients(storage storage.Iface) error {
	go syncclients.Cleanup(storage)
	return nil
}

func refreshAllClientsAndServer(storage storage.Iface, clientCache *wireguard.ClientCache) error {
	peerConfigPath := storage.ConfigPath(wireguard.VPN_CLIENTS_DIR)

	if _, err := os.Stat(peerConfigPath); errors.Is(err, os.ErrNotExist) {
		return nil // directory doesn't exist, so no configs to be read
	}

	entries, err := storage.ReadDir(peerConfigPath)
	if err != nil {
		return fmt.Errorf("can not list clients from dir %s: %s", peerConfigPath, err)
	}

	for _, e := range entries {
		peerConfig, err := wireguard.GetPeerConfigByFilename(storage, e)
		if err != nil {
			return fmt.Errorf("getClientFile error: %s", err)
		}
		err = wireguard.UpdateClientCache(peerConfig, clientCache)
		if err != nil {
			return fmt.Errorf("update client cache error: %s", err)
		}
		err = syncclients.SyncClients(storage, peerConfig)
		if err != nil {
			return fmt.Errorf("SyncClients error: %s", err)
		}
	}
	err = syncclients.Cleanup(storage)
	if err != nil {
		return fmt.Errorf("cleanup error: %s", err)
	}
	err = syncclients.UpdateServer(storage)
	if err != nil {
		return fmt.Errorf("UpdateServer error: %s", err)
	}
	return nil
}
