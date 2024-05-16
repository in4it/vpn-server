package oidcstore

import (
	"encoding/json"
	"fmt"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
)

func (store *Store) SaveOIDCStore() error {
	store.Mu.Lock()
	defer store.Mu.Unlock()
	out, err := json.Marshal(store)
	if err != nil {
		return fmt.Errorf("oidc store marshal error: %s", err)
	}
	filename := store.storage.ConfigPath("oidcstore.json")
	err = store.storage.WriteFile(filename, out)
	if err != nil {
		return fmt.Errorf("oidcstore write error: %s", err)
	}
	return nil
}

func (store *Store) SaveOAuth2Data(oauth2Data oidc.OAuthData, key string) error {
	store.Mu.Lock()
	store.OAuth2Data[key] = oauth2Data
	store.Mu.Unlock()
	return store.SaveOIDCStore()
}
