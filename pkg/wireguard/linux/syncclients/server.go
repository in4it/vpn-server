//go:build linux
// +build linux

package processpeerconfig

import (
	"fmt"
	"path"
	"strings"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	wireguardlinux "github.com/in4it/wireguard-server/pkg/wireguard/linux"
)

func UpdateServer(storage storage.Iface) error {
	c, available, err := wireguardlinux.New()
	if err != nil {
		return fmt.Errorf("cannot start wireguardlinux client: %s", err)
	}
	if !available {
		return fmt.Errorf("wireguard linux client not available")
	}
	device, err := c.Device(wireguard.VPN_INTERFACE_NAME)
	if err != nil {
		return fmt.Errorf("wireguard linux device 'vpn' not found: %s", err)
	}
	privateKeyBytes, err := storage.ReadFile(path.Join(wireguard.VPN_SERVER_SECRETS_PATH, wireguard.VPN_PRIVATE_KEY_FILENAME))
	if err != nil {
		return fmt.Errorf("failed to read private key: %s", err)
	}

	privateKey := strings.TrimSpace(string(privateKeyBytes))

	if device.PrivateKey.String() != privateKey {
		err := wgSetServerPrivateKey(path.Join(storage.GetPath(), wireguard.VPN_SERVER_SECRETS_PATH, wireguard.VPN_PRIVATE_KEY_FILENAME))
		if err != nil {
			return fmt.Errorf("failed to set server private key: %s", err)
		}
	}
	return nil
}
