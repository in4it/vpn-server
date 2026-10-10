package vpn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/in4it/go-devops-platform/rest"
	"github.com/in4it/go-devops-platform/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

// MAX_CONNECTIONS_PER_USER limits the amount of connections a single user can create, so a user
// can't exhaust the address pool. Applies to every user, including admins.
const MAX_CONNECTIONS_PER_USER = 25

var muClientDownload sync.Mutex

func (v *VPN) connectionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		user := r.Context().Value(rest.CustomValue("user")).(users.User)

		clients, err := v.Storage.ReadDir(v.Storage.ConfigPath(wireguard.VPN_CLIENTS_DIR))
		if err != nil {
			v.returnError(w, fmt.Errorf("cannot list connections for user: %s", err), http.StatusBadRequest)
			return
		}

		connectionList := []string{}
		for _, clientFilename := range clients {
			if wireguard.HasClientUserID(clientFilename, user.ID) {
				connectionList = append(connectionList, clientFilename)
			}
		}
		peerConfigs := make([]wireguard.PeerConfig, len(connectionList))
		for k, connection := range connectionList {
			var peerConfig wireguard.PeerConfig
			filename := v.Storage.ConfigPath(path.Join(wireguard.VPN_CLIENTS_DIR, connection))
			toDeleteFileContents, err := v.Storage.ReadFile(filename)
			if err != nil {
				v.returnError(w, fmt.Errorf("can't read file %s: %s", filename, err), http.StatusBadRequest)
				return
			}
			err = json.Unmarshal(toDeleteFileContents, &peerConfig)
			if err != nil {
				v.returnError(w, fmt.Errorf("can't unmarshal file %s: %s", filename, err), http.StatusBadRequest)
				return
			}
			peerConfigs[k] = peerConfig
		}
		connections := make([]Connection, len(peerConfigs))
		for k := range peerConfigs {
			connections[k] = Connection{
				ID:   peerConfigs[k].ID,
				Name: peerConfigs[k].Name,
			}
		}
		out, err := json.Marshal(connections)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not marshal list connection response: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	case http.MethodPost:
		muClientDownload.Lock()
		defer muClientDownload.Unlock()
		user := r.Context().Value(rest.CustomValue("user")).(users.User)
		existingConnections, err := wireguard.GetConfigNumbers(v.Storage, user.ID)
		if err != nil {
			v.returnError(w, fmt.Errorf("can't determine existing connections: %s", err), http.StatusBadRequest)
			return
		}
		if len(existingConnections) >= MAX_CONNECTIONS_PER_USER {
			v.returnError(w, fmt.Errorf("maximum number of connections reached (%d). Delete a connection before creating a new one", MAX_CONNECTIONS_PER_USER), http.StatusForbidden)
			return
		}
		peerConfig, err := wireguard.NewEmptyClientConfig(v.Storage, user.ID)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not generate client vpn config: %s", err), http.StatusBadRequest)
			return
		}
		newConnectionResponse := NewConnectionResponse{Name: peerConfig.Name}
		out, err := json.Marshal(newConnectionResponse)
		if err != nil {
			v.returnError(w, fmt.Errorf("could not marshal new connection response: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	default:
		v.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}
func (v *VPN) connectionsElementHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		user := r.Context().Value(rest.CustomValue("user")).(users.User)
		if err := validateConnectionID(r.PathValue("id"), user.ID); err != nil {
			v.returnError(w, err, http.StatusBadRequest)
			return
		}
		out, err := wireguard.GenerateNewClientConfig(v.Storage, r.PathValue("id"), user.ID)
		if err != nil {
			v.returnError(w, fmt.Errorf("GetClientConfig error: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, out)
	case http.MethodDelete:
		user := r.Context().Value(rest.CustomValue("user")).(users.User)
		if err := validateConnectionID(r.PathValue("id"), user.ID); err != nil {
			v.returnError(w, err, http.StatusBadRequest)
			return
		}
		err := wireguard.DeleteClientConfig(v.Storage, r.PathValue("id"), user.ID)
		if err != nil {
			v.returnError(w, fmt.Errorf("DeleteClientConfig error: %s", err), http.StatusBadRequest)
			return
		}
		v.write(w, []byte(`{"deleted": "`+r.PathValue("id")+`"}`))

	default:
		v.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

// validateConnectionID checks that the connection id has the format <userID>-<number>
func validateConnectionID(id, userID string) error {
	if strings.ContainsAny(id, `./\`) {
		return fmt.Errorf("connection id contains invalid characters")
	}
	if userID == "" || !strings.HasPrefix(id, userID+"-") {
		return fmt.Errorf("connection id is in invalid format (needs to contain user id)")
	}
	connectionNumber := strings.TrimPrefix(id, userID+"-")
	if connectionNumber == "" || strings.Trim(connectionNumber, "0123456789") != "" {
		return fmt.Errorf("connection id is in invalid format (needs to be user id followed by connection number)")
	}
	return nil
}

func (v *VPN) connectionLicenseHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(rest.CustomValue("user")).(users.User)
	licenseUserCount := r.Context().Value(rest.CustomValue("licenseUserCount")).(int)
	totalConnections, err := wireguard.GetConfigNumbers(v.Storage, user.ID)
	if err != nil {
		v.returnError(w, fmt.Errorf("can't determine total connections: %s", err), http.StatusBadRequest)
		return

	}
	out, err := json.Marshal(ConnectionLicenseResponse{LicenseUserCount: licenseUserCount, ConnectionCount: len(totalConnections)})
	if err != nil {
		v.returnError(w, fmt.Errorf("oidcProviders marshal error"), http.StatusBadRequest)
		return
	}
	v.write(w, out)
}
