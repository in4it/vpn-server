package saml

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (s *saml) samlHandler(w http.ResponseWriter, r *http.Request) {
	providerID := r.PathValue("id")

	if providerID == "" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("saml error: no provider specified\n"))
		return
	}

	provider, err := s.getProviderByID(providerID)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("saml error: can't find provider with specified id: %s", err)))
		return
	}

	err = s.ensureSPLoaded(provider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("saml error: invalid saml configuration: %s\n", err)))
		return
	}
	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("saml error: not a POST request\n"))
		return
	}

	if r.FormValue("SAMLResponse") == "" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("saml error: empty SAMLResponse\n"))
		return
	}

	if _, ok := s.serviceProvider[providerID]; !ok {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("saml error: can't find provider with specified id\n"))
		return
	}

	assertionInfo, err := s.serviceProvider[providerID].RetrieveAssertionInfo(r.FormValue("SAMLResponse"))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("saml error: %s\n", err)))
		return
	}

	if assertionInfo.WarningInfo.InvalidTime {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("saml error: invalid time\n"))
		return
	}

	if assertionInfo.WarningInfo.NotInAudience {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("saml error: incorrect audience\n"))
		return
	}

	login := assertionInfo.NameID
	notAfter := *assertionInfo.SessionNotOnOrAfter

	randomString, err := getRandomString(128)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("saml error: could not create session: %s\n", err)))
		return
	}
	sessionKey := SessionKey{
		ProviderID: providerID,
		SessionID:  randomString,
	}
	s.CreateSession(sessionKey, AuthenticatedUser{
		ID:        uuid.New().String(),
		Login:     login,
		ExpiresAt: notAfter,
	})
	w.Header().Add("Location", fmt.Sprintf("/callback/saml/%s?code=%s", providerID, sessionKey.SessionID))
	w.WriteHeader(http.StatusFound)
}
