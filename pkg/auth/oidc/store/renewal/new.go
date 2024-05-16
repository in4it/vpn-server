package oidcrenewal

import (
	"time"

	oidc "github.com/in4it/wireguard-server/pkg/auth/oidc"
	oidcstore "github.com/in4it/wireguard-server/pkg/auth/oidc/store"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/storage"
	"github.com/in4it/wireguard-server/pkg/users"
)

type Renewal struct {
	oidcStore     *oidcstore.Store
	enabled       bool
	oidcProviders []oidc.OIDCProvider
	userStore     *users.UserStore
	renewalTime   time.Duration
	storage       storage.Iface
}

func NewRenewal(storage storage.Iface, renewalTime int, contextLogLevel int, enabled bool, oidcstore *oidcstore.Store, oidcProviders []oidc.OIDCProvider, userStore *users.UserStore) (*Renewal, error) {
	r := &Renewal{
		enabled:       enabled,
		oidcStore:     oidcstore,
		oidcProviders: oidcProviders,
		userStore:     userStore,
		storage:       storage,
	}
	logging.Loglevel = contextLogLevel
	if renewalTime <= 5 {
		r.renewalTime = DEFAULT_RENEWAL_TIME_MINUTES * time.Minute
	} else {
		r.renewalTime = time.Duration(renewalTime) * time.Minute
	}
	go r.Worker()
	return r, nil
}

func (r *Renewal) SetEnabled(enabled bool) {
	r.enabled = enabled
}
