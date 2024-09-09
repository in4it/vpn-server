package oidcstore

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestGetJwks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/jwks.json" {
			jwksKeys := oidc.Jwks{
				Keys: []oidc.JwksKey{
					{
						Kid: "1-2-3-4",
						Kty: "kty",
					},
				},
			}
			out, err := json.Marshal(jwksKeys)
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
	uri := ts.URL + "/jwks.json"
	jwks, err := store.GetJwks(uri)
	if err != nil {
		t.Fatalf("get jwks error: %s", err)
	}
	if len(jwks.Keys) == 0 {
		t.Fatalf("jwks is empty")
	}
	if jwks.Keys[0].Kid != "1-2-3-4" {
		t.Fatalf("wrong kid: %s", jwks.Keys[0].Kid)
	}
	// cached response

	jwks, err = store.GetJwks(uri)
	if err != nil {
		t.Fatalf("get jwks error: %s", err)
	}
	if len(jwks.Keys) == 0 {
		t.Fatalf("jwks is empty")
	}
	if jwks.Keys[0].Kid != "1-2-3-4" {
		t.Fatalf("wrong kid: %s", jwks.Keys[0].Kid)
	}
	if _, ok := store.JwksCache[uri]; !ok {
		t.Fatalf("jwks not in cache")
	}

	// get all jwks
	allJwks, err := store.GetAllJwks([]oidc.Discovery{{JwksURI: uri}})
	if err != nil {
		t.Fatalf("get all jwks error: %s", err)
	}
	if len(allJwks) == 0 {
		t.Fatalf("all jwks is zero")
	}
	if allJwks[0].Keys[0].Kid != "1-2-3-4" {
		t.Fatalf("wrong kid for allJwks: %s", jwks.Keys[0].Kid)
	}
}
