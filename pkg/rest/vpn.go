package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *Context) connectionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		user := r.Context().Value(CustomValue("user")).(users.User)

		clients, err := c.Storage.Client.ReadDir(c.Storage.Client.ConfigPath(wireguard.VPN_CLIENTS_DIR))
		if err != nil {
			c.returnError(w, fmt.Errorf("cannot list connections for user: %s", err), http.StatusBadRequest)
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
			filename := c.Storage.Client.ConfigPath(path.Join(wireguard.VPN_CLIENTS_DIR, connection))
			toDeleteFileContents, err := c.Storage.Client.ReadFile(filename)
			if err != nil {
				c.returnError(w, fmt.Errorf("can't read file %s: %s", filename, err), http.StatusBadRequest)
				return
			}
			err = json.Unmarshal(toDeleteFileContents, &peerConfig)
			if err != nil {
				c.returnError(w, fmt.Errorf("can't unmarshal file %s: %s", filename, err), http.StatusBadRequest)
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
			c.returnError(w, fmt.Errorf("could not marshal list connection response: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		muClientDownload.Lock()
		defer muClientDownload.Unlock()
		user := r.Context().Value(CustomValue("user")).(users.User)
		peerConfig, err := wireguard.NewEmptyClientConfig(c.Storage.Client, user.ID)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not generate client vpn config: %s", err), http.StatusBadRequest)
			return
		}
		newConnectionResponse := NewConnectionResponse{Name: peerConfig.Name}
		out, err := json.Marshal(newConnectionResponse)
		if err != nil {
			c.returnError(w, fmt.Errorf("could not marshal new connection response: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}
func (c *Context) connectionsElementHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		user := r.Context().Value(CustomValue("user")).(users.User)
		if !strings.HasPrefix(r.PathValue("id"), user.ID) {
			c.returnError(w, fmt.Errorf("connection id is in invalid format (needs to contain user id)"), http.StatusBadRequest)
			return
		}
		if strings.Contains(r.PathValue("id"), ".") || strings.Contains(r.PathValue("id"), "/") {
			c.returnError(w, fmt.Errorf("connection id contains invalid characters"), http.StatusBadRequest)
			return
		}
		out, err := wireguard.GenerateNewClientConfig(c.Storage.Client, r.PathValue("id"), user.ID)
		if err != nil {
			c.returnError(w, fmt.Errorf("GetClientConfig error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodDelete:
		user := r.Context().Value(CustomValue("user")).(users.User)
		if !strings.HasPrefix(r.PathValue("id"), user.ID) {
			c.returnError(w, fmt.Errorf("connection id is in invalid format (needs to contain user id)"), http.StatusBadRequest)
			return
		}
		if strings.Contains(r.PathValue("id"), ".") || strings.Contains(r.PathValue("id"), "/") {
			c.returnError(w, fmt.Errorf("connection id contains invalid characters"), http.StatusBadRequest)
			return
		}
		err := wireguard.DeleteClientConfig(c.Storage.Client, r.PathValue("id"), user.ID)
		if err != nil {
			c.returnError(w, fmt.Errorf("DeleteClientConfig error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, []byte(`{"deleted": "`+r.PathValue("id")+`"}`))

	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}
