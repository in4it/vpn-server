package processpeerconfig

import (
	"testing"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func TestPeerAction(t *testing.T) {
	active := []string{"key-active"}
	tests := []struct {
		name       string
		peerConfig wireguard.PeerConfig
		want       PeerAction
	}{
		{"new peer is added", wireguard.PeerConfig{ID: "1", PublicKey: "key-new"}, PEER_ACTION_ADD},
		{"active peer is left alone", wireguard.PeerConfig{ID: "2", PublicKey: "key-active"}, PEER_ACTION_NONE},
		{"disabled peer is never added", wireguard.PeerConfig{ID: "3", PublicKey: "key-new", Disabled: true}, PEER_ACTION_NONE},
		{"disabled peer that is active is removed", wireguard.PeerConfig{ID: "4", PublicKey: "key-active", Disabled: true}, PEER_ACTION_REMOVE},
		{"peer without key is skipped", wireguard.PeerConfig{ID: "5"}, PEER_ACTION_SKIP_CORRUPT},
		{"disabled peer without key is ignored", wireguard.PeerConfig{ID: "6", Disabled: true}, PEER_ACTION_NONE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := peerAction(active, tt.peerConfig); got != tt.want {
				t.Fatalf("peerAction() = %d, want %d", got, tt.want)
			}
		})
	}
	if got := peerAction(nil, wireguard.PeerConfig{ID: "7", PublicKey: "key", Disabled: true}); got != PEER_ACTION_NONE {
		t.Fatalf("disabled peer on empty interface: got %d, want none", got)
	}
}

func TestPeersToRemove(t *testing.T) {
	// the peer config of key-corrupt couldn't be parsed, so it's not in the valid set
	peerConfigs := []wireguard.PeerConfig{
		{ID: "1", PublicKey: "key-valid"},
		{ID: "2", PublicKey: "key-disabled", Disabled: true},
		{ID: "3"}, // no key yet
	}
	active := []string{"key-valid", "key-disabled", "key-deleted", "key-corrupt"}
	got := peersToRemove(active, peerConfigs)
	want := []string{"key-disabled", "key-deleted", "key-corrupt"}
	if len(got) != len(want) {
		t.Fatalf("peersToRemove() = %v, want %v", got, want)
	}
	for k := range want {
		if got[k] != want[k] {
			t.Fatalf("peersToRemove() = %v, want %v", got, want)
		}
	}
	if got := peersToRemove(active, nil); len(got) != len(active) {
		t.Fatalf("without valid peer configs all peers must be removed, got %v", got)
	}
	if got := peersToRemove(nil, peerConfigs); len(got) != 0 {
		t.Fatalf("expected nothing to remove on an empty interface, got %v", got)
	}
}
