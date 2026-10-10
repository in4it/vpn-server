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

// peersToRemove returns the public keys of the peers on the interface that
// don't belong to a valid, enabled peer config. peerConfigs must only contain
// the configs that could be read and parsed: a peer whose config file is
// unreadable or corrupt is not in the valid set, and is removed.
func peersToRemove(activePublicKeys []string, peerConfigs []wireguard.PeerConfig) []string {
	validPublicKeys := make(map[string]struct{}, len(peerConfigs))
	for _, peerConfig := range peerConfigs {
		if !peerConfig.Disabled && peerConfig.PublicKey != "" {
			validPublicKeys[peerConfig.PublicKey] = struct{}{}
		}
	}
	toRemove := []string{}
	for _, publicKey := range activePublicKeys {
		if _, ok := validPublicKeys[publicKey]; !ok {
			toRemove = append(toRemove, publicKey)
		}
	}
	return toRemove
}
