package processpeerconfig

import (
	"slices"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

type PeerAction int

const (
	PEER_ACTION_NONE PeerAction = iota
	PEER_ACTION_ADD
	PEER_ACTION_REMOVE
	PEER_ACTION_SKIP_CORRUPT
)

// peerAction decides what to do with a peer config, given the public keys of
// the peers that are currently on the interface. Disabled peers (suspended or
// offboarded users) must never be added.
func peerAction(activePublicKeys []string, peerConfig wireguard.PeerConfig) PeerAction {
	if peerConfig.PublicKey == "" {
		if peerConfig.Disabled {
			return PEER_ACTION_NONE
		}
		return PEER_ACTION_SKIP_CORRUPT
	}
	active := slices.Contains(activePublicKeys, peerConfig.PublicKey)
	if peerConfig.Disabled {
		if active {
			return PEER_ACTION_REMOVE
		}
		return PEER_ACTION_NONE
	}
	if active {
		return PEER_ACTION_NONE
	}
	return PEER_ACTION_ADD
}
