//go:build darwin
// +build darwin

package configmanager

import (
	"fmt"

	"github.com/in4it/wireguard-server/pkg/storage"
)

func refreshAllClientsAndServer(storage storage.Iface) error {
	fmt.Printf("Warning: not refreshAllClients supported on darwin\n")
	return nil
}

func syncClient(storage storage.Iface, filename string) error {
	fmt.Printf("Warning: syncClient not supported on darwin. Cannot sync: %s\n", filename)
	return nil
}

func cleanupClients(storage storage.Iface) error {
	fmt.Printf("Warning: cleanupClients() not supported on darwin.")
	return nil
}

func deleteClient(storage storage.Iface, filename string) error {
	fmt.Printf("Warning: deleteClient not supported on darwin. Cannot delete: %s\n", filename)
	return nil
}
