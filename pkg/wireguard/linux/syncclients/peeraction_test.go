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
