package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	oidcstore "github.com/in4it/wireguard-server/pkg/auth/oidc/store"
	"github.com/in4it/wireguard-server/pkg/auth/saml"
	"github.com/in4it/wireguard-server/pkg/logging"
	"github.com/in4it/wireguard-server/pkg/rest/login"
)

func (c *Context) authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.returnError(w, fmt.Errorf("not a post request"), http.StatusBadRequest)
		return
	}

	if c.LocalAuthDisabled {
		c.returnError(w, fmt.Errorf("local auth is disabled in settings"), http.StatusForbidden)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var loginReq login.LoginRequest
	err := decoder.Decode(&loginReq)
	if err != nil {
		c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
		return
	}

	// check login attempts
	tooManyLogins := login.CheckTooManyLogins(c.LoginAttempts, loginReq.Login)
	if tooManyLogins {
		c.returnError(w, fmt.Errorf("too many login failures, try again later"), http.StatusTooManyRequests)
		return
	}

	loginResponse, user, err := login.Authenticate(loginReq, c.UserStore, c.JWTKeys.PrivateKey, c.JWTKeysKID)
	if err != nil {
		c.returnError(w, fmt.Errorf("authentication error: %s", err), http.StatusBadRequest)
		return
	}
	out, err := json.Marshal(loginResponse)
	if err != nil {
		c.returnError(w, fmt.Errorf("unable to marshal response: %s", err), http.StatusBadRequest)
		return
	}
	if loginResponse.MFARequired {
		c.write(w, out) // status ok, but unauthorized, because we need a second call with MFA code
		return
	} else if loginResponse.Authenticated {
		login.ClearAttemptsForLogin(c.LoginAttempts, loginReq.Login)
		user.LastLogin = time.Now()
		err = c.UserStore.UpdateUser(user)
		if err != nil {
			logging.ErrorLog(fmt.Errorf("last login update error: %s", err))
		}
		c.write(w, out)
	} else {
		// log login attempts
		login.RecordAttempt(c.LoginAttempts, loginReq.Login)
		// return Unauthorized
		c.writeWithStatus(w, out, http.StatusUnauthorized)
	}
}

func (c *Context) oidcProviderHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		oidcProviders := make([]oidc.OIDCProvider, len(c.OIDCProviders))
		copy(oidcProviders, c.OIDCProviders)
		for k := range oidcProviders {
			oidcProviders[k].LoginURL = fmt.Sprintf("%s://%s%s", c.Protocol, c.Hostname, strings.Replace(oidcProviders[k].RedirectURI, "/callback/", "/login/", -1))
			oidcProviders[k].RedirectURI = fmt.Sprintf("%s://%s%s", c.Protocol, c.Hostname, oidcProviders[k].RedirectURI)
		}
		out, err := json.Marshal(oidcProviders)
		if err != nil {
			c.returnError(w, fmt.Errorf("oidcProviders marshal error"), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		var oidcProvider oidc.OIDCProvider
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&oidcProvider)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
			return
		}
		oidcProvider.ID = uuid.New().String()
		if oidcProvider.Name == "" {
			c.returnError(w, fmt.Errorf("name not set"), http.StatusBadRequest)
			return
		}
		if oidcProvider.ClientID == "" {
			c.returnError(w, fmt.Errorf("clientID not set"), http.StatusBadRequest)
			return
		}
		if oidcProvider.ClientSecret == "" {
			c.returnError(w, fmt.Errorf("clientSecret not set"), http.StatusBadRequest)
			return
		}
		if oidcProvider.Scope == "" {
			c.returnError(w, fmt.Errorf("scope not set"), http.StatusBadRequest)
			return
		}
		if oidcProvider.DiscoveryURI == "" {
			c.returnError(w, fmt.Errorf("discovery URL not set"), http.StatusBadRequest)
			return
		}
		oidcProvider.RedirectURI = "/callback/oidc/" + oidcProvider.ID
		c.OIDCProviders = append(c.OIDCProviders, oidcProvider)
		out, err := json.Marshal(oidcProvider)
		if err != nil {
			c.returnError(w, fmt.Errorf("oidcProvider marshal error: %s", err), http.StatusBadRequest)
			return
		}
		err = SaveConfig(c)
		if err != nil {
			c.returnError(w, fmt.Errorf("saveConfig error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
		return
	}
}

func (c *Context) oidcProviderElementHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		match := -1
		for k, oidcProvider := range c.OIDCProviders {
			if oidcProvider.ID == r.PathValue("id") {
				match = k
			}
		}
		if match == -1 {
			c.returnError(w, fmt.Errorf("oidc provider not found"), http.StatusBadRequest)
			return
		}
		c.OIDCProviders = append(c.OIDCProviders[:match], c.OIDCProviders[match+1:]...)
		// save config (changed providers)
		err := SaveConfig(c)
		if err != nil {
			c.returnError(w, fmt.Errorf("saveConfig error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, []byte(`{ "deleted": "`+r.PathValue("id")+`" }`))
	}
}

func (c *Context) authMethods(w http.ResponseWriter, r *http.Request) {
	response := AuthMethodsResponse{
		LocalAuthDisabled: c.LocalAuthDisabled,
		OIDCProviders:     make([]AuthMethodsProvider, len(c.OIDCProviders)),
	}
	for k, oidcProvider := range c.OIDCProviders {
		response.OIDCProviders[k] = AuthMethodsProvider{
			ID:   oidcProvider.ID,
			Name: oidcProvider.Name,
		}
	}

	out, err := json.Marshal(response)
	if err != nil {
		c.returnError(w, fmt.Errorf("response marshal error: %s", err), http.StatusBadRequest)
		return
	}
	c.write(w, out)
}
func (c *Context) authMethodsByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		switch r.PathValue("method") {
		case "saml":
			loginResponse := login.LoginResponse{}
			var samlCallback SAMLCallback
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&samlCallback)
			if err != nil {
				c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
				return
			}
			if samlCallback.Code == "" {
				c.returnError(w, fmt.Errorf("no code provided"), http.StatusBadRequest)
				return
			}
			var samlProvider saml.Provider
			for k := range *c.SAML.Providers {
				if r.PathValue("id") == (*c.SAML.Providers)[k].ID {
					samlProvider = (*c.SAML.Providers)[k]
				}
			}
			if samlProvider.ID == "" {
				c.returnError(w, fmt.Errorf("saml provider not found"), http.StatusBadRequest)
				return
			}

			samlSession, err := c.SAML.Client.GetAuthenticatedUser(samlProvider, samlCallback.Code)
			if err != nil {
				c.returnError(w, fmt.Errorf("saml session not found"), http.StatusBadRequest)
				return
			}

			// add user to the user database (or modify existing one)
			user, err := addOrModifyExternalUser(c.Storage.Client, c.UserStore, samlSession.Login, "saml", samlSession.ID)
			if err != nil {
				c.returnError(w, fmt.Errorf("couldn't add/modify user in database: %s", err), http.StatusBadRequest)
				return
			}

			if user.Suspended {
				loginResponse.Suspended = true
			}

			token, err := login.GetJWTTokenWithExpiration(user.Login, user.Role, c.JWTKeys.PrivateKey, c.JWTKeysKID, samlSession.ExpiresAt)
			if err != nil {
				c.returnError(w, fmt.Errorf("token generation failed: %s", err), http.StatusBadRequest)
				return
			}
			loginResponse.Authenticated = true
			loginResponse.Token = token

			out, err := json.Marshal(loginResponse)
			if err != nil {
				c.returnError(w, fmt.Errorf("loginResponse Marshal error: %s", err), http.StatusBadRequest)
				return
			}
			c.write(w, out)
			return
		default: // oidc is default
			loginResponse := login.LoginResponse{}
			var oidcCallback OIDCCallback
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&oidcCallback)
			if err != nil {
				c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
				return
			}
			for _, oidcProvider := range c.OIDCProviders {
				if r.PathValue("id") == oidcProvider.ID && oidcCallback.Code != "" { // we got the code back
					oidcstore.RetrieveTokenLock.Lock()
					defer oidcstore.RetrieveTokenLock.Unlock()
					oauth2data, err := oidc.RetrieveOAUth2DataUsingState(c.OIDCStore.OAuth2Data, oidcCallback.State) // get the oauth2 struct based on the state (key)
					if err != nil {
						c.returnError(w, fmt.Errorf("cannot find oauth2 data using state provided: %s", err), http.StatusBadRequest)
						return
					}
					if oauth2data.Token.AccessToken != "" {
						if oauth2data.Suspended {
							loginResponse.Suspended = true
						} else if c.LicenseUserCount >= c.UserStore.UserCount() {
							loginResponse.NoLicense = true
						} else {
							loginResponse.Authenticated = true
							loginResponse.Token = oauth2data.Token.AccessToken
						}
						out, err := json.Marshal(loginResponse)
						if err != nil {
							c.returnError(w, fmt.Errorf("loginResponse Marshal error: %s", err), http.StatusBadRequest)
							return
						}
						c.write(w, out)
						return
					}
					// no token, let's generate a new one
					discovery, err := c.OIDCStore.GetDiscoveryURI(oidcProvider.DiscoveryURI)
					if err != nil {
						c.returnError(w, fmt.Errorf("getDiscoveryURI error: %s", err), http.StatusBadRequest)
						return
					}
					jwks, err := c.OIDCStore.GetJwks(discovery.JwksURI)
					if err != nil {
						c.returnError(w, fmt.Errorf("get jwks error: %s", err), http.StatusBadRequest)
						return
					}
					updatedOauth2data, err := oidc.UpdateOAuth2DataWithToken(jwks, discovery, oidcProvider.ClientID, oidcProvider.ClientSecret, c.Protocol+"://"+c.Hostname+oidcCallback.RedirectURI, oidcCallback.Code, oidcCallback.State, oauth2data)
					if err != nil {
						c.returnError(w, fmt.Errorf("GetTokenFromCode error: %s", err), http.StatusBadRequest)
						return
					}
					// add user to the user database (or modify existing one)
					user, err := addOrModifyExternalUser(c.Storage.Client, c.UserStore, updatedOauth2data.UserInfo.Email, "oidc", updatedOauth2data.ID)
					if err != nil {
						c.returnError(w, fmt.Errorf("couldn't add/modify user in database: %s", err), http.StatusBadRequest)
						return
					}
					if user.Suspended {
						loginResponse.Suspended = true
						updatedOauth2data.Suspended = true
					} else {
						updatedOauth2data.Suspended = false
					}
					// save oauth data (only when we're sure it's not a suspended user)
					err = c.OIDCStore.SaveOAuth2Data(updatedOauth2data, oidcCallback.State)
					if err != nil {
						c.returnError(w, fmt.Errorf("oidc store save failed: %s", err), http.StatusBadRequest)
						return
					}
					// cleanup oauth2 data
					c.OIDCStore.CleanupOAuth2Data(updatedOauth2data)

					// save config (changed user info)
					err = SaveConfig(c)
					if err != nil {
						c.returnError(w, fmt.Errorf("saveConfig error: %s", err), http.StatusBadRequest)
						return
					}

					// set loginResponse
					if !loginResponse.Suspended {
						loginResponse.Authenticated = true
						loginResponse.Token = updatedOauth2data.Token.AccessToken
					}
					out, err := json.Marshal(loginResponse)
					if err != nil {
						c.returnError(w, fmt.Errorf("loginResponse Marshal error: %s", err), http.StatusBadRequest)
						return
					}
					c.write(w, out)
					return
				}
			}
		}
		c.returnError(w, fmt.Errorf("oidc provider not found"), http.StatusBadRequest)
	case http.MethodGet:
		switch r.PathValue("method") {
		case "saml":
			id := r.PathValue("id")
			samlProviderId := -1
			for k := range *c.SAML.Providers {
				if (*c.SAML.Providers)[k].ID == id {
					samlProviderId = k
				}
			}
			if samlProviderId == -1 {
				c.returnError(w, fmt.Errorf("cannot find saml provider"), http.StatusBadRequest)
				return
			}
			redirectURI, err := c.SAML.Client.GetAuthURL((*c.SAML.Providers)[samlProviderId])
			if err != nil {
				c.returnError(w, fmt.Errorf("cannot get auth url"), http.StatusBadRequest)
				return
			}
			response := AuthMethodsProvider{
				ID:          (*c.SAML.Providers)[samlProviderId].ID,
				Name:        (*c.SAML.Providers)[samlProviderId].Name,
				RedirectURI: redirectURI,
			}
			out, err := json.Marshal(response)
			if err != nil {
				c.returnError(w, fmt.Errorf("response marshal error: %s", err), http.StatusBadRequest)
				return
			}
			c.write(w, out)
			return
		default:
			id := r.PathValue("id")
			for _, oidcProvider := range c.OIDCProviders {
				if id == oidcProvider.ID {
					callback := fmt.Sprintf("%s://%s%s", c.Protocol, c.Hostname, oidcProvider.RedirectURI)
					discovery, err := c.OIDCStore.GetDiscoveryURI(oidcProvider.DiscoveryURI)
					if err != nil {
						c.returnError(w, fmt.Errorf("getDiscoveryURI error: %s", err), http.StatusBadRequest)
						return
					}
					redirectURI, state, err := oidc.GetRedirectURI(discovery, oidcProvider.ClientID, oidcProvider.Scope, callback, c.EnableOIDCTokenRenewal)
					if err != nil {
						c.returnError(w, fmt.Errorf("GetRedirectURI error: %s", err), http.StatusBadRequest)
						return
					}
					response := AuthMethodsProvider{
						ID:          oidcProvider.ID,
						Name:        oidcProvider.Name,
						RedirectURI: redirectURI,
					}
					out, err := json.Marshal(response)
					if err != nil {
						c.returnError(w, fmt.Errorf("response marshal error: %s", err), http.StatusBadRequest)
						return
					}
					newOAuthEntry := oidc.OAuthData{
						ID:             uuid.NewString(),
						OIDCProviderID: response.ID,
						CreatedAt:      time.Now(),
					}
					err = c.OIDCStore.SaveOAuth2Data(newOAuthEntry, state)
					if err != nil {
						c.returnError(w, fmt.Errorf("unable to save state to oidc store: %s", err), http.StatusBadRequest)
						return
					}
					c.write(w, out)
					return
				}
			}
			c.returnError(w, fmt.Errorf("element not found"), http.StatusBadRequest)
		}
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *Context) oidcRenewTokensHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.OIDCRenewal.RenewAllOIDCConnections()
		c.write(w, []byte(`{"status": "done"}`))
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}
