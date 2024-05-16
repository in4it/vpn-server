//go:build linux
// +build linux

package configmanager

import (
	"errors"
	"fmt"
	"os"

	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	syncclients "github.com/in4it/wireguard-server/pkg/wireguard/linux/syncclients"
)

func syncClient(storage storage.Iface, filename string) error {
	go syncclients.SyncClientsAndCleanup(storage, filename)
	return nil
}
func deleteClient(storage storage.Iface, filename string) error {
	go syncclients.DeleteClient(storage, filename)
	return nil
}
func cleanupClients(storage storage.Iface) error {
	go syncclients.Cleanup(storage)
	return nil
}

func refreshAllClientsAndServer(storage storage.Iface) error {
	peerConfigPath := storage.ConfigPath(wireguard.VPN_CLIENTS_DIR)

	if _, err := os.Stat(peerConfigPath); errors.Is(err, os.ErrNotExist) {
		return nil // directory doesn't exist, so no configs to be read
	}

	entries, err := storage.ReadDir(peerConfigPath)
	if err != nil {
		return fmt.Errorf("can not list clients from dir %s: %s", peerConfigPath, err)
	}

	for _, e := range entries {
		err = syncclients.SyncClients(storage, e)
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
