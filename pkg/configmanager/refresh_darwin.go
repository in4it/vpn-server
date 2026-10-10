//go:build darwin

package configmanager

import (
	"fmt"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

type darwinPeerSyncer struct{}

func newPeerSyncer() peerSyncer {
	return darwinPeerSyncer{}
}

func (darwinPeerSyncer) SyncClient(storage storage.Iface, peerConfig wireguard.PeerConfig) error {
	fmt.Printf("Warning: syncClient not supported on darwin. Cannot sync: %s\n", peerConfig.ID)
	return nil
}

func (darwinPeerSyncer) Cleanup(storage storage.Iface) error {
	fmt.Printf("Warning: cleanupClients() not supported on darwin.\n")
	return nil
}

func (darwinPeerSyncer) UpdateServer(storage storage.Iface) error {
	fmt.Printf("Warning: refreshAllClients not supported on darwin\n")
	return nil
}
