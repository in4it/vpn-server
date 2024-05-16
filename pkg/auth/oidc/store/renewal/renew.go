package oidcrenewal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	oidcstore "github.com/in4it/wireguard-server/pkg/auth/oidc/store"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (r *Renewal) RenewAllOIDCConnections() {
	// force renewal of all tokens, even if they're not expired (unless they're empty)
	for key, oauth2Data := range r.oidcStore.OAuth2Data {
		if oidcProvider, err := getOIDCProvider(oauth2Data.OIDCProviderID, r.oidcProviders); err == nil {
			if discovery, err := r.oidcStore.GetDiscoveryURI(oidcProvider.DiscoveryURI); err == nil {
				if oauth2Data.RenewalFailed || oauth2Data.Token.AccessToken == "" {
					logging.DebugLog(fmt.Errorf("skipping %s (renewal already failed or access token is empty. RenewalFailed: %v, AccessToken is empty: %v)", oauth2Data.ID, oauth2Data.RenewalFailed, oauth2Data.Token.AccessToken == ""))
				} else {
					logging.DebugLog(fmt.Errorf("trying to renew %s", oauth2Data.ID))
					r.renew(discovery, key, oauth2Data, oidcProvider)
				}
			} else {
				logging.DebugLog(fmt.Errorf("could not get discovery url for %s: %s", oauth2Data.ID, err))
			}
		} else {
			logging.DebugLog(fmt.Errorf("could not get oidcprovider for %s: %s", oauth2Data.ID, err))
		}
	}
}
func (r *Renewal) renew(discovery oidc.Discovery, key string, oauth2Data oidc.OAuthData, oidcProvider oidc.OIDCProvider) {
	newToken, newTokenTimestamp, err := refreshToken(discovery, oauth2Data.Token.RefreshToken, oidcProvider.ClientID, oidcProvider.ClientSecret)
	if err != nil {
		oauth2Data.RenewalRetries++
		logging.ErrorLog(fmt.Errorf("renewal Worker: could not refresh token for %s (attemp %d/%d): %s", oauth2Data.ID, oauth2Data.RenewalRetries, RENEWAL_RETRIES, err))
		if oauth2Data.RenewalRetries >= RENEWAL_RETRIES {
			oauth2Data.RenewalFailed = true
		}
		err = r.oidcStore.StoreEntry(key, oauth2Data)
		if err != nil {
			logging.ErrorLog(fmt.Errorf("renewal Worker: [error] StoreEntry: %s", err))
		}
		err = r.oidcStore.SaveOIDCStore()
		if err != nil {
			logging.ErrorLog(fmt.Errorf("renewal Worker: [error] SaveOIDCStore: %s", err))
		}
		// suspend connections
		if oauth2Data.RenewalFailed {
			err = disableUser(r.storage, oauth2Data, r.userStore)
			if err != nil {
				logging.ErrorLog(fmt.Errorf("renewal Worker: [error] disableUser: %s", err))
			}
		}
		return
	}
	logging.DebugLog(fmt.Errorf("new token issued at %v: %+v", newToken, newTokenTimestamp))
	oauth2Data.LastTokenRenewal = newTokenTimestamp
	oauth2Data.Token.AccessToken = newToken.AccessToken
	oauth2Data.Token.ExpiresIn = newToken.ExpiresIn
	oauth2Data.Token.RefreshToken = newToken.RefreshToken
	if newToken.IDToken != "" {
		oauth2Data.Token.IDToken = newToken.IDToken
	}
	err = r.oidcStore.StoreEntry(key, oauth2Data)
	if err != nil {
		logging.ErrorLog(fmt.Errorf("renewal Worker: [error] StoreEntry: %s", err))
	}
	err = r.oidcStore.SaveOIDCStore()
	if err != nil {
		logging.ErrorLog(fmt.Errorf("renewal Worker: [error] SaveOIDCStore: %s", err))
	}
}

func disableUser(storage storage.Iface, oauth2Data oidc.OAuthData, userStore *users.UserStore) error {
	logging.DebugLog(fmt.Errorf("disable user with oidc id %s", oauth2Data.ID))
	user, err := userStore.GetUserByOIDCIDs([]string{oauth2Data.ID})
	if err != nil {
		return fmt.Errorf("no user found with oidc id %s", oauth2Data.ID)
	}
	err = wireguard.DisableAllClientConfigs(storage, user.ID)
	if err != nil {
		return fmt.Errorf("DisableAllClientConfigs error for userID %s: %s", user.ID, err)
	}
	user.ConnectionsDisabledOnAuthFailure = true
	err = userStore.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("could not update connectionsDisabledOnAuthFailure user with userID %s: %s", user.ID, err)
	}
	return nil
}

func getOIDCProvider(id string, oidcProviders []oidc.OIDCProvider) (oidc.OIDCProvider, error) {
	for _, oidcProvider := range oidcProviders {
		if oidcProvider.ID == id {
			return oidcProvider, nil
		}
	}
	return oidc.OIDCProvider{}, fmt.Errorf("oidc provider not found")

}

func getExpirationDate(token string) (time.Time, error) {
	jwtSplit := strings.Split(token, ".")
	if len(jwtSplit) < 2 {
		return time.Time{}, fmt.Errorf("token split < 2")
	}
	data, err := base64.RawURLEncoding.DecodeString(jwtSplit[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("could not base64 decode data part of jwt")
	}
	var jwt jwtExp
	err = json.Unmarshal(data, &jwt)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not unmarshal jwt data")
	}
	if jwt.Expiration == 0 {
		return time.Time{}, fmt.Errorf("exp not found in jwt data")
	}
	return time.Unix(jwt.Expiration, 0), nil
}

func canRenew(renewalTime time.Duration, oauth2Data oidc.OAuthData, store *oidcstore.Store, oidcProviders []oidc.OIDCProvider) (bool, oidc.OIDCProvider, oidc.Discovery, error) {
	if oauth2Data.RenewalFailed {
		return false, oidc.OIDCProvider{}, oidc.Discovery{}, nil
	}
	if oauth2Data.Token.AccessToken == "" {
		logging.DebugLog(fmt.Errorf("access token empty of oidc id %s", oauth2Data.ID))
		return false, oidc.OIDCProvider{}, oidc.Discovery{}, nil
	}
	expirationDate, err := getExpirationDate(oauth2Data.Token.AccessToken)
	if err != nil {
		return false, oidc.OIDCProvider{}, oidc.Discovery{}, fmt.Errorf("can't get expiration date of refresh_token (id:%s). error: %s", oauth2Data.ID, err)
	}

	if time.Since(oauth2Data.LastTokenRenewal) > renewalTime || time.Now().After(expirationDate) {
		logging.DebugLog(fmt.Errorf("going to renew token for %s", oauth2Data.ID))
		oidcProvider, err := getOIDCProvider(oauth2Data.OIDCProviderID, oidcProviders)
		if err != nil {
			return false, oidc.OIDCProvider{}, oidc.Discovery{}, fmt.Errorf("could not get oidcprovider for %s: %s", oauth2Data.ID, err)
		}
		discovery, err := store.GetDiscoveryURI(oidcProvider.DiscoveryURI)
		if err != nil {
			return false, oidc.OIDCProvider{}, oidc.Discovery{}, fmt.Errorf("could not get discovery url for %s: %s", oauth2Data.ID, err)
		}
		return true, oidcProvider, discovery, nil
	} else {
		logging.DebugLog(fmt.Errorf("not renewing oidc id %s. time since last token renewal: %d", oauth2Data.ID, time.Since(oauth2Data.LastTokenRenewal)))
	}
	return false, oidc.OIDCProvider{}, oidc.Discovery{}, nil
}
