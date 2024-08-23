//go:build linux
// +build linux

package stats

import (
	"fmt"
	"time"

	wireguardlinux "github.com/in4it/wireguard-server/pkg/wireguard/linux"
)

func GetStats() ([]PeerStat, error) {
	c, available, err := wireguardlinux.New()
	if err != nil {
		return []PeerStat{}, fmt.Errorf("cannot start wireguardlinux client: %s", err)
	}
	if !available {
		return []PeerStat{}, fmt.Errorf("wireguard linux client not available")
	}
	device, err := c.Device(wireguardlinux.VPN_INTERFACE_NAME)
	if err != nil {
		return []PeerStat{}, fmt.Errorf("wireguard linux device 'vpn' not found: %s", err)
	}

	peerStats := make([]PeerStat, len(device.Peers))

	for k, peer := range device.Peers {
		peerStats[k] = PeerStat{
			Timestamp:         time.Now(),
			PublicKey:         peer.PublicKey.String(),
			LastHandshakeTime: peer.LastHandshakeTime,
			ReceiveBytes:      peer.ReceiveBytes,
			TransmitBytes:     peer.TransmitBytes,
		}
	}
	return peerStats, nil
}
