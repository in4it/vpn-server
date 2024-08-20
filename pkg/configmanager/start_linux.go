//go:build linux
// +build linux

package configmanager

import (
	"log"

	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func startVPN(storage storage.Iface) error {
	err := wireguard.WriteWireGuardServerConfig(storage)
	if err != nil {
		log.Fatalf("WriteWireGuardServerConfig error: %s", err)
	}
	return wireguard.StartVPN()
}

func stopVPN(storage storage.Iface) error {
	return wireguard.StopVPN()
}
