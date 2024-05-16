package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/user"
	"sync"

	"github.com/in4it/wireguard-server/pkg/storage"
)

var mu sync.Mutex

func SaveConfig(c *Context) error {
	mu.Lock()
	defer mu.Unlock()
	cCopy := *c
	cCopy.SCIM = &SCIM{ // we don't save the client, but we want the token and enabled
		EnableSCIM: c.SCIM.EnableSCIM,
		Token:      c.SCIM.Token,
	}
	cCopy.SAML = &SAML{ // we don't save the client, but we want the config
		Providers: c.SAML.Providers,
	}
	cCopy.JWTKeys = nil       // we retrieve JWTKeys from pem files at startup
	cCopy.OIDCStore = nil     // we save this separately
	cCopy.UserStore = nil     // we save this separately
	cCopy.OIDCRenewal = nil   // we don't save this
	cCopy.LoginAttempts = nil // no need to save this
	cCopy.Observability = nil // no need to save the client
	cCopy.Storage = nil       // no need to save storage
	out, err := json.Marshal(cCopy)
	if err != nil {
		return fmt.Errorf("context marshal error: %s", err)
	}
	err = c.Storage.Client.WriteFile(c.Storage.Client.ConfigPath("config.json"), out)
	if err != nil {
		return fmt.Errorf("config write error: %s", err)
	}
	// fix permissions
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %s", err)
	}
	if currentUser.Username != "vpn" {
		err = c.Storage.Client.EnsureOwnership(c.Storage.Client.ConfigPath("config.json"), "vpn")
		if err != nil {
			return fmt.Errorf("config write error: %s", err)
		}
	}

	return nil
}

func GetConfig(storage storage.Iface) (*Context, error) {
	var c *Context

	appDir := storage.GetPath()

	// check if config exists
	if !storage.FileExists(storage.ConfigPath("config.json")) {
		return getEmptyContext(appDir)
	}

	data, err := storage.ReadFile(storage.ConfigPath("config.json"))
	if err != nil {
		return c, fmt.Errorf("config read error: %s", err)
	}
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	err = decoder.Decode(&c)
	if err != nil {
		return c, fmt.Errorf("decode input error: %s", err)
	}

	c.AppDir = appDir

	return c, nil
}
