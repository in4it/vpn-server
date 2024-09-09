package oidcstore

import (
	"testing"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestSave(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	store, err := NewStore(storage)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	err = store.setDiscoveryCache("test", oidc.DiscoveryCache{Discovery: oidc.Discovery{Issuer: "testissuer"}})
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	err = store.SaveOIDCStore()
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	_, err = storage.ReadFile(storage.ConfigPath("oidcstore.json"))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	store2, err := NewStore(storage)
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	discovery, ok := store2.getDiscoveryCache("test")
	if !ok {
		t.Fatalf("can't find the cache")
	}
	if discovery.Discovery.Issuer != "testissuer" {
		t.Fatalf("expected testissuer. Got: %s", discovery.Discovery.Issuer)
	}
}
