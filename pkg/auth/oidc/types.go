package oidc

import (
	"time"
)

type Discovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
}

type OIDCProvider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret,omitempty"`
	Scope        string `json:"scope"`
	DiscoveryURI string `json:"discoveryURI"`
	RedirectURI  string `json:"redirectURI"`
	LoginURL     string `json:"loginURL,omitempty"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

// jwks
type Jwks struct {
	Keys []JwksKey `json:"keys"`
}
type JwksKey struct {
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
}

type DiscoveryCache struct {
	Expiration time.Time `json:"expiration"`
	Discovery  Discovery `json:"discovery"`
}
type JwksCache struct {
	Expiration time.Time `json:"expiration"`
	Jwks       Jwks      `json:"jwks"`
}
type OAuthData struct {
	ID               string    `json:"id"`
	OIDCProviderID   string    `json:"oidcProviderID"`
	CreatedAt        time.Time `json:"createdAt"`
	Subject          string    `json:"subject"`
	Issuer           string    `json:"issuer"`
	UserInfo         UserInfo  `json:"userInfo"`
	Token            Token     `json:"token"`
	AuthFailed       bool      `json:"authFailed"`
	Suspended        bool      `json:"suspended"`
	LastTokenRenewal time.Time `json:"lastTokenRenewal"`
	RenewalFailed    bool      `json:"renewalFailed"`
	RenewalRetries   int       `json:"renewalRetries"`
}

type UserInfo struct {
	Email string `json:"email"`
}
