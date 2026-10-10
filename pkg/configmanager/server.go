package configmanager

import (
	"fmt"
	"log"
	"net/http"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	"github.com/in4it/wireguard-server/pkg/wireguard/fsutils"
)

func StartServer(port int) {
	dir, err := executableDir()
	if err != nil {
		log.Fatalf("couldn't initialize storage: %s", err)
	}
	// set the immutable flag on the binaries (existing installs get it after an upgrade)
	ensureImmutableBinaries(dir, getBinaries())
	// the configmanager writes files in a directory owned by the vpn user:
	// use a storage that doesn't follow symbolic links
	localStorage, err := fsutils.NewNoFollowStorageWithConfigDirs(dir)
	if err != nil {
		log.Fatalf("couldn't initialize storage: %s", err)
	}
	c, err := initConfigManager(localStorage)
	if err != nil {
		log.Fatalf("Couldn't init config manager: %s", err)
	}

	// start server
	isRunning, err := wireguard.IsVPNRunning()
	if err != nil {
		log.Fatalf("couldn't check whether vpn is running or not: %s", err)
	}
	if !isRunning {
		log.Printf("VPN is not running. Starting...")
		err = startVPN(localStorage)
		if err != nil {
			// keep the config manager running, so the config can be fixed in the UI
			// and the vpn restarted. Exiting would make systemd give up restarting us.
			log.Printf("Error: couldn't start vpn: %s", err)
		} else {
			log.Printf("VPN Server started\n")
			isRunning = true
		}
	}

	// refresh all clients
	if isRunning {
		err = refreshAllClientsAndServer(localStorage, c.ClientCache)
		if err != nil {
			log.Printf("Error: could not refresh all clients: %s", err)
		}
	}

	// start goroutines
	startStats(localStorage)                                    // start gathering of wireguard stats
	startPacketLogger(localStorage, c.ClientCache, c.VPNConfig) // start packet logger (optional)

	log.Printf("Starting localhost http server at port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), c.getRouter()))
}

func initConfigManager(storage storage.Iface) (*ConfigManager, error) {
	c := &ConfigManager{
		Storage: storage,
		ClientCache: &wireguard.ClientCache{
			Addresses: []wireguard.ClientCacheAddresses{},
		},
	}

	vpnConfig, err := wireguard.GetVPNConfig(storage)
	if err != nil {
		return c, fmt.Errorf("failed to get vpn config: %s", err)
	}

	if vpnConfig.Endpoint == "" && vpnConfig.Port == 0 {
		vpnConfig, err = wireguard.CreateNewVPNConfig(storage)
		if err != nil {
			return c, fmt.Errorf("failed to create new vpn config: %s", err)
		}
		err = writeSetupCode(storage)
		if err != nil {
			return c, fmt.Errorf("failed to write setup-code.txt: %s", err)
		}
	}

	c.VPNConfig = &vpnConfig

	return c, nil
}
