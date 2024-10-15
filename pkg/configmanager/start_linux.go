//go:build linux
// +build linux

package configmanager

import (
	"log"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func startVPN(storage storage.Iface) error {
	err := wireguard.WriteWireGuardServerConfig(storage)
	if err != nil {
		log.Fatalf("WriteWireGuardServerConfig error: %s", err)
	}

	return wireguard.StartVPN()
}

func stopVPN() error {
	return wireguard.StopVPN()
}

func startStats(storage storage.Iface) {
	// run statistics go routine
	go wireguard.RunStats(storage)
}

func startPacketLogger(storage storage.Iface, clientCache *wireguard.ClientCache, vpnConfig *wireguard.VPNConfig) {
	// run statistics go routine
	go wireguard.RunPacketLogger(storage, clientCache, vpnConfig)
	// run cleanup
	go wireguard.PacketLoggerLogRotation(storage)
}
