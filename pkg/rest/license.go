package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/in4it/wireguard-server/pkg/license"
	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *Context) licenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("action") == "get-more" {
		c.LicenseUserCount = license.RefreshLicense(c.Storage.Client, c.CloudType, c.LicenseUserCount)
	}

	currentUserCount := c.UserStore.UserCount()
	licenseResponse := LicenseResponse{LicenseUserCount: c.LicenseUserCount, CurrentUserCount: currentUserCount, CloudType: c.CloudType}

	if r.PathValue("action") == "get-more" {
		licenseResponse.Key = license.GetLicenseKey(c.Storage.Client, c.CloudType)
	}

	out, err := json.Marshal(licenseResponse)
	if err != nil {
		c.returnError(w, fmt.Errorf("oidcProviders marshal error"), http.StatusBadRequest)
		return
	}
	c.write(w, out)
}
func (c *Context) connectionLicenseHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(CustomValue("user")).(users.User)
	totalConnections, err := wireguard.GetConfigNumbers(c.Storage.Client, user.ID)
	if err != nil {
		c.returnError(w, fmt.Errorf("can't determine total connections: %s", err), http.StatusBadRequest)
		return

	}
	out, err := json.Marshal(ConnectionLicenseResponse{LicenseUserCount: c.LicenseUserCount, ConnectionCount: len(totalConnections)})
	if err != nil {
		c.returnError(w, fmt.Errorf("oidcProviders marshal error"), http.StatusBadRequest)
		return
	}
	c.write(w, out)
}
