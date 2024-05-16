package oidcstore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
)

func (store *Store) GetJwks(jwksURI string) (oidc.Jwks, error) {
	var jwksKeys oidc.Jwks

	if cachedJwks, ok := store.JwksCache[jwksURI]; ok {
		if cachedJwks.Expiration.Before(time.Now()) {
			return cachedJwks.Jwks, nil // cache hit, we can return
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(jwksURI)
	if err != nil {
		return jwksKeys, fmt.Errorf("discoveryURL Get error: %s", err)
	}
	if resp.StatusCode != 200 {
		return jwksKeys, fmt.Errorf("DiscoveryURI Request unsuccesful (status code returned: %d)", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&jwksKeys)
	if err != nil {
		return jwksKeys, fmt.Errorf("discoveryURL decode error: %s", err)
	}

	store.setJwksCache(jwksURI, oidc.JwksCache{
		Expiration: time.Now().Add(20 * time.Minute), // 20 minute cache standard
		Jwks:       jwksKeys,
	})

	return jwksKeys, nil
}

func (store *Store) GetAllJwks(discoveryProviders []oidc.Discovery) ([]oidc.Jwks, error) {
	allJwks := make([]oidc.Jwks, len(discoveryProviders))
	for k, discovery := range discoveryProviders {
		var err error
		allJwks[k], err = store.GetJwks(discovery.JwksURI)
		if err != nil {
			return []oidc.Jwks{}, fmt.Errorf("get jwks error: %s", err)
		}
	}
	return allJwks, nil
}
