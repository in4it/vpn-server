//go:build darwin
// +build darwin

package configmanager

import (
	"fmt"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func startVPN(storage storage.Iface) error {
	fmt.Printf("Warning: startVPN is not implemented in darwin\n")
	return nil
}

func stopVPN() error {
	fmt.Printf("Warning: startVPN is not implemented in darwin\n")
	return nil
}

func startStats(storage storage.Iface) {
	fmt.Printf("Warning: startStats is not implemented in darwin\n")
}

func startPacketLogger(storage storage.Iface, clientCache *wireguard.ClientCache, vpnConfig *wireguard.VPNConfig) {
	go wireguard.RunPacketLogger(storage, clientCache, vpnConfig)
	// run cleanup
	go wireguard.PacketLoggerLogRotation(storage)
}
