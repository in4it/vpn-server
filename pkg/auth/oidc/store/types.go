package oidcstore

import (
	"sync"

	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	"github.com/in4it/wireguard-server/pkg/storage"
)

type Store struct {
	Mu             sync.Mutex
	OAuth2Data     map[string]oidc.OAuthData      `json:"oauth2Data"`
	DiscoveryCache map[string]oidc.DiscoveryCache `json:"discoveryCache"`
	JwksCache      map[string]oidc.JwksCache      `json:"jwksCache"`
	storage        storage.Iface
}
