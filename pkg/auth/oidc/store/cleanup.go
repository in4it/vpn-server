package oidcstore

import (
	"slices"
	"time"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
)

func (store *Store) CleanupOAuth2DataForAllEntries() int {
	deleted := 0
	for _, oauthData := range store.OAuth2Data {
		deleted += store.CleanupOAuth2Data(oauthData)
	}
	return deleted
}

func (store *Store) CleanupOAuth2Data(oauthData oidc.OAuthData) int {
	keysToDelete1 := []string{}
	keysToDelete2 := []string{}
	keysToDelete3 := []string{}
	store.Mu.Lock()
	defer store.Mu.Unlock()
	for k := range store.OAuth2Data {
		if oauthData.CreatedAt.After(store.OAuth2Data[k].CreatedAt) {
			// cleanup old oauth2 data that might be duplicates (same subject & oidc provider, but older tokens)
			if store.OAuth2Data[k].ID != oauthData.ID && store.OAuth2Data[k].OIDCProviderID == oauthData.OIDCProviderID && store.OAuth2Data[k].Subject == oauthData.Subject {
				keysToDelete1 = append(keysToDelete1, k)
			}
			// cleanup oauthdata with the same email address
			if store.OAuth2Data[k].ID != oauthData.ID && store.OAuth2Data[k].UserInfo.Email == oauthData.UserInfo.Email {
				keysToDelete3 = append(keysToDelete3, k)
			}
		}
		// cleanup old oauth2 data that doesn't have a token and is stale
		if store.OAuth2Data[k].Token.AccessToken == "" && store.OAuth2Data[k].CreatedAt.Add(10*time.Minute).After(time.Now()) {
			keysToDelete2 = append(keysToDelete2, k)
		}
	}
	keysToDelete := []string{}
	keysToDelete = append(keysToDelete, keysToDelete1...)
	keysToDelete = append(keysToDelete, keysToDelete2...)
	keysToDelete = append(keysToDelete, keysToDelete3...)
	slices.Sort(keysToDelete)

	for _, key := range slices.Compact(keysToDelete) {
		delete(store.OAuth2Data, key)
	}
	return len(keysToDelete)
}
