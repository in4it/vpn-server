package configmanager

import (
	"errors"
	"log"
	"os"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

// serverConfigPath is the wireguard config written by writeServerConfig (variable for tests)
var serverConfigPath = wireguard.WIREGUARD_CONFIG

// writeServerConfig writes the wireguard server config. When the server template
// contains settings that aren't supported, the existing config (written earlier
// by the config manager) is kept, so the vpn keeps working until the template is
// fixed. Saving the template or restarting the vpn from the UI reports the error.
func writeServerConfig(storage storage.Iface, write func(storage.Iface) error) error {
	err := write(storage)
	if err == nil {
		return nil
	}
	if errors.Is(err, wireguard.ErrServerConfigNotAllowed) {
		if info, statErr := os.Lstat(serverConfigPath); statErr == nil && info.Mode().IsRegular() {
			log.Printf("Error: %s. Using the existing %s until the server template is updated", err, serverConfigPath)
			return nil
		}
	}
	return err
}
