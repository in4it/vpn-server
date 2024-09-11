package rest

import (
	"fmt"
	"sync"
	"time"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	oidcstore "github.com/in4it/wireguard-server/pkg/auth/oidc/store"
	oidcrenewal "github.com/in4it/wireguard-server/pkg/auth/oidc/store/renewal"
	"github.com/in4it/wireguard-server/pkg/auth/provisioning/scim"
	"github.com/in4it/wireguard-server/pkg/auth/saml"
	"github.com/in4it/wireguard-server/pkg/license"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/observability"
	"github.com/in4it/wireguard-server/pkg/rest/login"
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/users"
)

var muClientDownload sync.Mutex

func newContext(storage storage.Iface, serverType string) (*Context, error) {
	c, err := GetConfig(storage)
	if err != nil {
		return c, fmt.Errorf("getConfig error: %s", err)
	}
	c.ServerType = serverType

	c.Storage = &Storage{
		Client: storage,
	}

	c.JWTKeys, err = getJWTKeys(storage)
	if err != nil {
		return c, fmt.Errorf("getJWTKeys error: %s", err)
	}
	c.OIDCStore, err = oidcstore.NewStore(storage)
	if err != nil {
		return c, fmt.Errorf("getOIDCStore error: %s", err)
	}
	if c.OIDCProviders == nil {
		c.OIDCProviders = []oidc.OIDCProvider{}
	}

	c.LicenseUserCount, c.CloudType = license.GetMaxUsers(c.Storage.Client)
	go func() { // run license refresh
		logging.DebugLog(fmt.Errorf("starting license refresh in background (current licenses: %d, cloud type: %s)", c.LicenseUserCount, c.CloudType))
		for {
			time.Sleep(time.Hour * 24)
			newLicenseCount := license.RefreshLicense(storage, c.CloudType, c.LicenseUserCount)
			if newLicenseCount != c.LicenseUserCount {
				logging.InfoLog(fmt.Sprintf("License changed from %d users to %d users", c.LicenseUserCount, newLicenseCount))
				c.LicenseUserCount = newLicenseCount
			}
		}
	}()
	c.UserStore, err = users.NewUserStore(c.Storage.Client, c.LicenseUserCount)
	if err != nil {
		return c, fmt.Errorf("userstore initialization error: %s", err)
	}

	c.OIDCRenewal, err = oidcrenewal.NewRenewal(storage, c.TokenRenewalTimeMinutes, c.LogLevel, c.EnableOIDCTokenRenewal, c.OIDCStore, c.OIDCProviders, c.UserStore)
	if err != nil {
		return c, fmt.Errorf("oidcrenewal init error: %s", err)
	}

	if c.LoginAttempts == nil {
		c.LoginAttempts = make(login.Attempts)
	}

	if c.SCIM == nil {
		c.SCIM = &SCIM{
			Client:     scim.New(storage, c.UserStore, ""),
			Token:      "",
			EnableSCIM: false,
		}
	} else {
		c.SCIM.Client = scim.New(storage, c.UserStore, c.SCIM.Token)
	}
	if c.SAML == nil {
		providers := []saml.Provider{}
		c.SAML = &SAML{
			Client:    saml.New(&providers, storage, &c.Protocol, &c.Hostname),
			Providers: &providers,
		}
	} else {
		c.SAML.Client = saml.New(c.SAML.Providers, storage, &c.Protocol, &c.Hostname)
	}

	if c.Observability == nil {
		c.Observability = &Observability{
			Client: observability.New(),
		}
	} else {
		c.Observability.Client = observability.New()
	}

	return c, nil
}

func getEmptyContext(appDir string) (*Context, error) {
	randomString, err := oidc.GetRandomString(64)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate random string for local kid")
	}
	c := &Context{
		AppDir:                  appDir,
		JWTKeysKID:              randomString,
		TokenRenewalTimeMinutes: oidcrenewal.DEFAULT_RENEWAL_TIME_MINUTES,
		LogLevel:                logging.LOG_ERROR,
		SCIM:                    &SCIM{EnableSCIM: false},
		SAML:                    &SAML{Providers: &[]saml.Provider{}},
	}
	return c, nil
}
