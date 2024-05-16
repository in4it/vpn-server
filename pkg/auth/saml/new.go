package saml

import (
	"github.com/in4it/wireguard-server/pkg/storage"
	saml2 "github.com/russellhaering/gosaml2"
)

func New(providers *[]Provider, storage storage.Iface, protocol, hostname *string) Iface {
	s := &saml{
		Providers:       providers,
		sessions:        make(map[SessionKey]AuthenticatedUser),
		serviceProvider: make(map[string]*saml2.SAMLServiceProvider),
		protocol:        protocol,
		hostname:        hostname,
		storage:         storage,
	}
	for _, provider := range *providers {
		s.loadSP(provider)
	}
	return s
}
