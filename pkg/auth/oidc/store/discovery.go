package oidcstore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
)

func (store *Store) GetDiscoveryURI(discoveryURI string) (oidc.Discovery, error) {
	var discovery oidc.Discovery
	if cachedDiscovery, ok := store.getDiscoveryCache(discoveryURI); ok {
		if cachedDiscovery.Expiration.After(time.Now()) {
			return cachedDiscovery.Discovery, nil // cache hit, we can return
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(discoveryURI)
	if err != nil {
		return discovery, fmt.Errorf("discoveryURL Get error: %s", err)
	}
	if resp.StatusCode != 200 {
		return discovery, fmt.Errorf("DiscoveryURI Request unsuccesful (status code returned: %d)", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&discovery)
	if err != nil {
		return discovery, fmt.Errorf("discoveryURL decode error: %s", err)
	}

	store.setDiscoveryCache(discoveryURI, oidc.DiscoveryCache{
		Expiration: time.Now().Add(12 * time.Hour), // standard 12h cache
		Discovery:  discovery,
	})

	return discovery, nil
}
