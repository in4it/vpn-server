package wireguard

import (
	"fmt"
	"net"
)

func IsVPNRunning() (bool, error) {
	addrs, err := net.Interfaces()
	if err != nil {
		return false, fmt.Errorf("net.Interfaces error: %s", err)
	}
	for _, addr := range addrs {
		if addr.Name == VPN_INTERFACE_NAME {
			return true, nil
		}
	}
	return false, nil
}
