//go:build linux

package processpeerconfig

import (
	"errors"
	"fmt"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
	wireguardlinux "github.com/in4it/wireguard-server/pkg/wireguard/linux"
)

// Cleanup removes the peers from the interface that don't have a valid and
// enabled connection file. Connection files that can't be read or parsed are
// skipped, which means their peer is removed.
func Cleanup(storage storage.Iface) error {
	peerConfigs, err := wireguard.GetAllValidPeerConfigs(storage)
	if err != nil { // can't list the connections: don't remove anything
		return fmt.Errorf("cannot list files in users clients directory: %s", err)
	}

	c, available, err := wireguardlinux.New()
	if err != nil {
		return fmt.Errorf("cannot start wireguardlinux client: %s", err)
	}
	if !available {
		return fmt.Errorf("wireguard linux client not available")
	}
	defer c.Close() //nolint:errcheck
	device, err := c.Device(wireguard.VPN_INTERFACE_NAME)
	if err != nil {
		return fmt.Errorf("wireguard linux device 'vpn' not found: %s", err)
	}

	activePublicKeys := make([]string, len(device.Peers))
	for k, peer := range device.Peers {
		activePublicKeys[k] = peer.PublicKey.String()
	}

	errs := []error{}
	for _, publicKey := range peersToRemove(activePublicKeys, peerConfigs) {
		// keep going when a removal fails, so the other peers are still removed
		if err := wgDeletePeer(wireguard.PeerConfig{PublicKey: publicKey}); err != nil {
			errs = append(errs, fmt.Errorf("wgDeletePeer error: %s", err))
		}
	}

	return errors.Join(errs...)
}
