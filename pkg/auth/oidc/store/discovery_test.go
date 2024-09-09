package oidcstore

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestGetDiscovery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/discovery.json" {
			discovery := oidc.Discovery{
				Issuer: "test-issuer",
			}
			out, err := json.Marshal(discovery)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}
			w.Write(out)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	store, err := NewStore(&memorystorage.MockMemoryStorage{})
	if err != nil {
		t.Fatalf("new store error: %s", err)
	}
	uri := ts.URL + "/discovery.json"
	discovery, err := store.GetDiscoveryURI(uri)
	if err != nil {
		t.Fatalf("get discovery error: %s", err)
	}
	if discovery.Issuer != "test-issuer" {
		t.Fatalf("wrong issuer")
	}

	// cached response
	discovery, err = store.GetDiscoveryURI(uri)
	if err != nil {
		t.Fatalf("get discovery error: %s", err)
	}
	if discovery.Issuer != "test-issuer" {
		t.Fatalf("wrong issuer")
	}
	if _, ok := store.DiscoveryCache[uri]; !ok {
		t.Fatalf("discovery not in cache")
	}
}
