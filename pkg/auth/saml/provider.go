package saml

import "fmt"

func (s *saml) getProviderByID(id string) (Provider, error) {
	for k := range *s.Providers {
		if (*s.Providers)[k].ID == id {
			return (*s.Providers)[k], nil
		}
	}
	return Provider{}, fmt.Errorf("provider not found")
}
