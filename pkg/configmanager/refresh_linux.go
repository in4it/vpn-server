//go:build linux

package configmanager

import (
	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	syncclients "github.com/in4it/wireguard-server/pkg/wireguard/linux/syncclients"
)

type linuxPeerSyncer struct{}

func newPeerSyncer() peerSyncer {
	return linuxPeerSyncer{}
}

func (linuxPeerSyncer) SyncClient(storage storage.Iface, peerConfig wireguard.PeerConfig) error {
	return syncclients.SyncClients(storage, peerConfig)
}

func (linuxPeerSyncer) Cleanup(storage storage.Iface) error {
	return syncclients.Cleanup(storage)
}

func (linuxPeerSyncer) UpdateServer(storage storage.Iface) error {
	return syncclients.UpdateServer(storage)
}
