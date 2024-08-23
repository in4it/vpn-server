package configmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func (c *ConfigManager) getPubKey(w http.ResponseWriter, r *http.Request) {
	var pubKeyExchange wireguard.PubKeyExchange
	pubKeyExchange.PubKey = c.PublicKey
	out, err := json.Marshal(pubKeyExchange)
	if err != nil {
		returnError(w, fmt.Errorf("pub exchange marshal error: %s", err), http.StatusBadRequest)
		return
	}
	w.Write(out)
}
func (c *ConfigManager) refreshClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var payload wireguard.RefreshClientRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&payload)
		if err != nil {
			returnError(w, fmt.Errorf("wrong payload (expected refresh client request)"), http.StatusBadRequest)
			return
		}
		if payload.Action == wireguard.ACTION_CLEANUP {
			err = cleanupClients(c.Storage)
			if err != nil {
				returnError(w, fmt.Errorf("cleanup clients error: %s", err), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			return // no further actions needed
		}

		// action is add / delete
		if len(payload.Filenames) == 0 {
			returnError(w, fmt.Errorf("wrong payload (no filenames supplied)"), http.StatusBadRequest)
			return
		}
		if payload.Action != wireguard.ACTION_ADD && payload.Action != wireguard.ACTION_DELETE {
			returnError(w, fmt.Errorf("wrong action supplied"), http.StatusBadRequest)
			return
		}
		for _, filename := range payload.Filenames {
			if filename == "" {
				returnError(w, fmt.Errorf("wrong payload (expected not empty filename in refresh client request)"), http.StatusBadRequest)
				return
			}
			if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
				returnError(w, fmt.Errorf("filename in wrong format"), http.StatusBadRequest)
				return
			}
			switch payload.Action {
			case wireguard.ACTION_ADD:
				err = syncClient(c.Storage, filename)
				if err != nil {
					returnError(w, fmt.Errorf("syncClient error: %s", err), http.StatusBadRequest)
					return
				}
			case wireguard.ACTION_DELETE:
				err = deleteClient(c.Storage, filename)
				if err != nil {
					returnError(w, fmt.Errorf("deleteClient error: %s", err), http.StatusBadRequest)
					return
				}
			}
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *ConfigManager) upgrade(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		newVersionAvailable, version, err := newVersionAvailable()
		if err != nil {
			returnError(w, fmt.Errorf("new version available error: %s", err), http.StatusBadRequest)
			return
		}
		out, err := json.Marshal(UpgradeResponse{NewVersionAvailable: newVersionAvailable, NewVersion: version, CurrentVersion: getVersion()})
		if err != nil {
			returnError(w, fmt.Errorf("upgrade response marshal error: %s", err), http.StatusBadRequest)
			return
		}
		w.Write(out)
	case http.MethodPost:
		w.Write([]byte(`{"upgrade": "starting"}`))
		err := upgrade()
		if err != nil {
			fmt.Printf("upgrade failed: %s\n", err)
		}
	default:
		returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *ConfigManager) version(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		out, err := json.Marshal(map[string]string{"version": getVersion()})
		if err != nil {
			returnError(w, fmt.Errorf("version marshal error: %s", err), http.StatusBadRequest)
			return
		}
		w.Write(out)
	default:
		returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *ConfigManager) restartVpn(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		err := stopVPN()
		if err != nil { // don't exit, as the VPN might be down already.
			fmt.Println("========= Warning =========")
			fmt.Printf("Warning: vpn stop error: %s\n", err)
			fmt.Println("=========================")
		}
		err = startVPN(c.Storage)
		if err != nil {
			returnError(w, fmt.Errorf("vpn start error: %s", err), http.StatusBadRequest)
			return
		}
		err = refreshAllClientsAndServer(c.Storage)
		if err != nil {
			returnError(w, fmt.Errorf("could not refresh all clients: %s", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func returnError(w http.ResponseWriter, err error, statusCode int) {
	fmt.Println("========= ERROR =========")
	fmt.Printf("Error: %s\n", err)
	fmt.Println("=========================")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + err.Error() + `"}`))
}
