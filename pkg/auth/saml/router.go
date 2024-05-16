package saml

import (
	"net/http"
)

func (s *saml) GetRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/saml/acs/{id}", http.HandlerFunc(s.samlHandler))

	return mux
}
