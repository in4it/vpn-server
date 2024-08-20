package rest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	"github.com/in4it/wireguard-server/pkg/auth/saml"
	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *Context) contextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)
		var contextReq ContextRequest
		err := decoder.Decode(&contextReq)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
			return
		}
		if !c.Storage.Client.FileExists(SETUP_CODE_FILE) {
			c.SetupCompleted = true
		}
		if !c.SetupCompleted {
			localSecret, err := c.Storage.Client.ReadFile(SETUP_CODE_FILE)
			if err != nil {
				c.returnError(w, fmt.Errorf("secret file read error: %s", err), http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(string(localSecret)) != contextReq.Secret {
				c.returnError(w, fmt.Errorf("wrong secret provided"), http.StatusUnauthorized)
				return
			}
			if contextReq.AdminPassword != "" {
				adminUser := users.User{
					Login:    "admin",
					Password: contextReq.AdminPassword,
					Role:     "admin",
				}
				if c.UserStore.LoginExists("admin") {
					err = c.UserStore.UpdateUser(adminUser)
					if err != nil {
						c.returnError(w, fmt.Errorf("could not update user: %s", err), http.StatusBadRequest)
						return
					}
				} else {
					_, err = c.UserStore.AddUser(adminUser)
					if err != nil {
						c.returnError(w, fmt.Errorf("could not add user: %s", err), http.StatusBadRequest)
						return
					}
				}

				c.SetupCompleted = true
				c.Hostname = contextReq.Hostname
				protocol := contextReq.Protocol
				protocol = strings.Replace(protocol, "http:", "http", -1)
				protocol = strings.Replace(protocol, "https:", "https", -1)
				c.Protocol = protocol

				err = SaveConfig(c)
				if err != nil {
					c.SetupCompleted = false
					c.returnError(w, fmt.Errorf("unable to save file: %s", err), http.StatusBadRequest)
					return
				}

				// update hostname in vpn config
				vpnconfig, err := wireguard.GetVPNConfig(c.Storage.Client)
				if err != nil {
					c.SetupCompleted = false
					c.returnError(w, fmt.Errorf("unable to get vpn-config: %s", err), http.StatusBadRequest)
					return
				}
				vpnconfig.Endpoint = c.Hostname
				err = wireguard.WriteVPNConfig(c.Storage.Client, vpnconfig)
				if err != nil {
					c.SetupCompleted = false
					c.returnError(w, fmt.Errorf("unable to write vpn-config: %s", err), http.StatusBadRequest)
					return
				}
			}
		}
	}
	cOut := Context{
		SetupCompleted: c.SetupCompleted,
	}
	out, err := json.Marshal(cOut)
	if err != nil {
		c.returnError(w, err, http.StatusBadRequest)
		return
	}
	c.write(w, out)
}

func (c *Context) setupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		setupRequest := GeneralSetupRequest{
			Hostname:               c.Hostname,
			EnableTLS:              c.EnableTLS,
			RedirectToHttps:        c.RedirectToHttps,
			DisableLocalAuth:       c.LocalAuthDisabled,
			EnableOIDCTokenRenewal: c.EnableOIDCTokenRenewal,
		}
		out, err := json.Marshal(setupRequest)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		var setupRequest GeneralSetupRequest
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&setupRequest)
		if c.Hostname != setupRequest.Hostname {
			c.Hostname = setupRequest.Hostname
		}
		if c.RedirectToHttps != setupRequest.RedirectToHttps {
			c.RedirectToHttps = setupRequest.RedirectToHttps
		}
		if c.EnableTLS != setupRequest.EnableTLS {
			if !c.EnableTLS && setupRequest.EnableTLS && !TLSWaiterCompleted && canEnableTLS(c.Hostname) {
				enableTLSWaiter <- true
			}
			c.EnableTLS = setupRequest.EnableTLS
		}
		if c.LocalAuthDisabled != setupRequest.DisableLocalAuth {
			c.LocalAuthDisabled = setupRequest.DisableLocalAuth
		}
		if c.EnableOIDCTokenRenewal != setupRequest.EnableOIDCTokenRenewal {
			c.EnableOIDCTokenRenewal = setupRequest.EnableOIDCTokenRenewal
			c.OIDCRenewal.SetEnabled(c.EnableOIDCTokenRenewal)
		}
		err := SaveConfig(c)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not save config to disk: %s", err), http.StatusBadRequest)
			return
		}
		out, err := json.Marshal(setupRequest)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *Context) vpnSetupHandler(w http.ResponseWriter, r *http.Request) {
	vpnConfig, err := wireguard.GetVPNConfig(c.Storage.Client)
	if err != nil {
		c.returnError(w, fmt.Errorf("could not get vpn config: %s", err), http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		setupRequest := VPNSetupRequest{
			Routes:              strings.Join(vpnConfig.ClientRoutes, ", "),
			VPNEndpoint:         vpnConfig.Endpoint,
			AddressRange:        vpnConfig.AddressRange.String(),
			ClientAddressPrefix: vpnConfig.ClientAddressPrefix,
			Port:                vpnConfig.Port,
			ExternalInterface:   vpnConfig.ExternalInterface,
			Nameservers:         strings.Join(vpnConfig.Nameservers, ","),
			DisableNAT:          vpnConfig.DisableNAT,
		}
		out, err := json.Marshal(setupRequest)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		var setupRequest VPNSetupRequest
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&setupRequest)
		if strings.Join(vpnConfig.ClientRoutes, ", ") != setupRequest.Routes {
			networks := strings.Split(setupRequest.Routes, ",")
			validatedNetworks := []string{}
			for _, network := range networks {
				if strings.TrimSpace(network) == "::/0" {
					validatedNetworks = append(validatedNetworks, "::/0")
				} else {
					_, ipnet, err := net.ParseCIDR(network)
					if err == nil {
						validatedNetworks = append(validatedNetworks, ipnet.String())
					}
				}
			}
			vpnConfig.ClientRoutes = validatedNetworks
			err = wireguard.WriteVPNConfig(c.Storage.Client, vpnConfig)
			if err != nil {
				c.returnError(w, fmt.Errorf("could write vpn config: %s", err), http.StatusBadRequest)
				return
			}
			// rewrite client configs
			err = wireguard.UpdateClientsConfig(c.Storage.Client)
			if err != nil {
				c.returnError(w, fmt.Errorf("could not update client vpn configs: %s", err), http.StatusBadRequest)
				return
			}
		}
		if vpnConfig.Endpoint != setupRequest.VPNEndpoint {
			vpnConfig.Endpoint = setupRequest.VPNEndpoint
			err = wireguard.WriteVPNConfig(c.Storage.Client, vpnConfig)
			if err != nil {
				c.SetupCompleted = false
				c.returnError(w, fmt.Errorf("unable to write vpn-config: %s", err), http.StatusBadRequest)
				return
			}
			// rewrite client configs
			err = wireguard.UpdateClientsConfig(c.Storage.Client)
			if err != nil {
				c.returnError(w, fmt.Errorf("could not update client vpn configs: %s", err), http.StatusBadRequest)
				return
			}
		}
		err := SaveConfig(c)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not save config to disk: %s", err), http.StatusBadRequest)
			return
		}
		out, err := json.Marshal(setupRequest)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal SetupRequest: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *Context) scimSetupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		scimSetup := SCIMSetup{
			Enabled: c.SCIM.EnableSCIM,
		}
		if c.SCIM.EnableSCIM {
			scimSetup.Token = c.SCIM.Token
			scimSetup.BaseURL = fmt.Sprintf("%s://%s/%s", c.Protocol, c.Hostname, "api/scim/v2/")
		}
		out, err := json.Marshal(scimSetup)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal scim setup: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		saveConfig := false
		var scimSetupRequest SCIMSetup
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&scimSetupRequest)
		if scimSetupRequest.Enabled && !c.SCIM.EnableSCIM {
			c.SCIM.EnableSCIM = true
			saveConfig = true
		}
		if !scimSetupRequest.Enabled && c.SCIM.EnableSCIM {
			c.SCIM.EnableSCIM = false
			saveConfig = true
		}
		if scimSetupRequest.RegenerateToken || (scimSetupRequest.Enabled && c.SCIM.Token == "") {
			// Generate new token
			randomString, err := oidc.GetRandomString(64)
			if err != nil {
				c.returnError(w, fmt.Errorf("could not enable scim: %s", err), http.StatusBadRequest)
				return
			}
			token := base64.StdEncoding.EncodeToString([]byte(randomString))
			scimSetupRequest.Token = token
			c.SCIM.Token = token
			c.SCIM.Client.UpdateToken(token)
			saveConfig = true
		}
		if saveConfig {
			// save config
			err := SaveConfig(c)
			if err != nil {
				c.returnError(w, fmt.Errorf("could not save config to disk: %s", err), http.StatusBadRequest)
				return
			}
		}
		out, err := json.Marshal(scimSetupRequest)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal scim setup: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *Context) samlSetupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		samlProviders := make([]saml.Provider, len(*c.SAML.Providers))
		copy(samlProviders, *c.SAML.Providers)
		for k := range samlProviders {
			samlProviders[k].Issuer = fmt.Sprintf("%s://%s/%s/%s", c.Protocol, c.Hostname, saml.ISSUER_URL, samlProviders[k].ID)
			samlProviders[k].Audience = fmt.Sprintf("%s://%s/%s/%s", c.Protocol, c.Hostname, saml.AUDIENCE_URL, samlProviders[k].ID)
			samlProviders[k].Acs = fmt.Sprintf("%s://%s/%s/%s", c.Protocol, c.Hostname, saml.ACS_URL, samlProviders[k].ID)
		}
		out, err := json.Marshal(samlProviders)
		if err != nil {
			c.returnError(w, fmt.Errorf("oidcProviders marshal error"), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		var samlProvider saml.Provider
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&samlProvider)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
			return
		}
		samlProvider.ID = uuid.New().String()
		if samlProvider.Name == "" {
			c.returnError(w, fmt.Errorf("name not set"), http.StatusBadRequest)
			return
		}
		if samlProvider.MetadataURL == "" {
			c.returnError(w, fmt.Errorf("metadata URL not set"), http.StatusBadRequest)
			return
		}
		_, err = c.SAML.Client.HasValidMetadataURL(samlProvider.MetadataURL)
		if err != nil {
			c.returnError(w, fmt.Errorf("metadata error: %s", err), http.StatusBadRequest)
			return
		}

		*c.SAML.Providers = append(*c.SAML.Providers, samlProvider)
		out, err := json.Marshal(samlProvider)
		if err != nil {
			c.returnError(w, fmt.Errorf("samlProvider marshal error: %s", err), http.StatusBadRequest)
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
	}
}

func (c *Context) samlSetupElementHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		match := -1
		for k, samlProvider := range *c.SAML.Providers {
			if samlProvider.ID == r.PathValue("id") {
				match = k
			}
		}
		if match == -1 {
			c.returnError(w, fmt.Errorf("saml provider not found"), http.StatusBadRequest)
			return
		}
		*c.SAML.Providers = append((*c.SAML.Providers)[:match], (*c.SAML.Providers)[match+1:]...)
		// save config (changed providers)
		err := SaveConfig(c)
		if err != nil {
			c.returnError(w, fmt.Errorf("saveConfig error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, []byte(`{ "deleted": "`+r.PathValue("id")+`" }`))
	case http.MethodPut:
		var samlProvider saml.Provider
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&samlProvider)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
			return
		}
		samlProviderID := -1
		for k := range *c.SAML.Providers {
			if (*c.SAML.Providers)[k].ID == r.PathValue("id") {
				samlProviderID = k
			}
		}
		if samlProviderID == -1 {
			c.returnError(w, fmt.Errorf("cannot find saml provider: %s", err), http.StatusBadRequest)
			return
		}
		saveConfig := false
		if (*c.SAML.Providers)[samlProviderID].AllowMissingAttributes != samlProvider.AllowMissingAttributes {
			(*c.SAML.Providers)[samlProviderID].AllowMissingAttributes = samlProvider.AllowMissingAttributes
			saveConfig = true
		}
		if (*c.SAML.Providers)[samlProviderID].MetadataURL != samlProvider.MetadataURL {
			_, err := c.SAML.Client.HasValidMetadataURL(samlProvider.MetadataURL)
			if err != nil {
				c.returnError(w, fmt.Errorf("metadata error: %s", err), http.StatusBadRequest)
				return
			}
			(*c.SAML.Providers)[samlProviderID].MetadataURL = samlProvider.MetadataURL
			saveConfig = true
		}
		out, err := json.Marshal(samlProvider)
		if err != nil {
			c.returnError(w, fmt.Errorf("samlProvider marshal error: %s", err), http.StatusBadRequest)
			return
		}
		if saveConfig {
			err = SaveConfig(c)
			if err != nil {
				c.returnError(w, fmt.Errorf("saveConfig error: %s", err), http.StatusBadRequest)
				return
			}
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}
