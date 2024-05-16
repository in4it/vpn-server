package saml

import (
	"fmt"
	"time"
)

func (s *saml) GetAuthenticatedUser(provider Provider, sessionID string) (AuthenticatedUser, error) {
	sessionKey := SessionKey{
		ProviderID: provider.ID,
		SessionID:  sessionID,
	}
	if authenticatedUser, ok := s.sessions[sessionKey]; ok {
		if authenticatedUser.ExpiresAt.Before(time.Now()) {
			return authenticatedUser, fmt.Errorf("session is expired")
		}
		return authenticatedUser, nil
	}
	return AuthenticatedUser{}, fmt.Errorf("session not found")
}

func (s *saml) CreateSession(key SessionKey, value AuthenticatedUser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[key] = value
}
