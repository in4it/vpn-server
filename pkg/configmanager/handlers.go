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
	_, err = w.Write(out)
	if err != nil {
		returnError(w, fmt.Errorf("write error: %s", err), http.StatusBadRequest)
		return
	}
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
			err = cleanupClients(c.Storage, c.ClientCache)
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
				err = syncClient(c.Storage, filename, c.ClientCache)
				if err != nil {
					returnError(w, fmt.Errorf("syncClient error: %s", err), http.StatusBadRequest)
					return
				}
			case wireguard.ACTION_DELETE:
				err = deleteClient(c.Storage, filename, c.ClientCache)
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
func (c *ConfigManager) refreshServerConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		vpnConfig, err := wireguard.GetVPNConfig(c.Storage)
		if err != nil {
			returnError(w, fmt.Errorf("get vpn config error: %s", err), http.StatusBadRequest)
			return
		}
		// the packet logger reads these settings concurrently, so update them under its lock
		startPacketLogger := wireguard.SetPacketLogConfig(c.VPNConfig, vpnConfig.EnablePacketLogs, vpnConfig.PacketLogsTypes)
		if startPacketLogger {
			go runPacketLogger(c.Storage, c.ClientCache, c.VPNConfig)
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
		_, err = w.Write(out)
		if err != nil {
			returnError(w, fmt.Errorf("write error: %s", err), http.StatusBadRequest)
			return
		}
	case http.MethodPost:
		// only one upgrade at a time
		if !upgradeMutex.TryLock() {
			returnError(w, fmt.Errorf("upgrade already in progress"), http.StatusConflict)
			return
		}
		_, err := w.Write([]byte(`{"upgrade": "starting"}`))
		if err != nil {
			upgradeMutex.Unlock()
			fmt.Printf("upgrade: write error: %s\n", err)
			return
		}
		// the download can take longer than the rest-server's request timeout,
		// so run the upgrade in the background (the services are restarted at the end)
		go func() {
			defer upgradeMutex.Unlock()
			if err := runUpgrade(); err != nil {
				fmt.Printf("upgrade failed: %s\n", err)
			}
		}()
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
		_, err = w.Write(out)
		if err != nil {
			returnError(w, fmt.Errorf("write error: %s", err), http.StatusBadRequest)
			return
		}
	default:
		returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)
	}
}

func (c *ConfigManager) restartVpn(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// make sure the config renders before taking the vpn down, otherwise a
		// broken template leaves the vpn down
		err := wireguard.CheckWireGuardServerConfig(c.Storage)
		if err != nil {
			returnError(w, fmt.Errorf("vpn config error, vpn not restarted: %s", err), http.StatusBadRequest)
			return
		}
		err = vpnStop()
		if err != nil { // don't exit, as the VPN might be down already.
			fmt.Println("========= Warning =========")
			fmt.Printf("Warning: vpn stop error: %s\n", err)
			fmt.Println("=========================")
		}
		err = vpnStart(c.Storage)
		if err != nil {
			returnError(w, fmt.Errorf("vpn start error: %s", err), http.StatusBadRequest)
			return
		}
		err = vpnRefreshClients(c.Storage, c.ClientCache)
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
	_, err = w.Write([]byte(`{"error": "` + err.Error() + `"}`))
	if err != nil {
		fmt.Printf("write error: %s", err)
		return
	}
}
