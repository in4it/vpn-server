package configmanager

import (
	"fmt"
	"log"
	"net/http"

	"github.com/in4it/wireguard-server/pkg/storage"
	localstorage "github.com/in4it/wireguard-server/pkg/storage/local"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func StartServer(port int) {
	localStorage, err := localstorage.New()
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
			log.Fatalf("couldn't start vpn: %s", err)
		}
		log.Printf("VPN Server started\n")
	}

	// refresh all clients
	err = refreshAllClientsAndServer(localStorage)
	if err != nil {
		log.Fatalf("could not refresh all clients: %s", err)
	}

	// start goroutines
	startStats(localStorage)        // start gathering of wireguard stats
	startPacketLogger(localStorage) // start packet logger (optional)

	log.Printf("Starting localhost http server at port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), c.getRouter()))
}

func initConfigManager(storage storage.Iface) (*ConfigManager, error) {
	c := &ConfigManager{
		Storage: storage,
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

	return c, nil
}
