package wireguard

import (
	"fmt"
	"net"
)

func getClientAllowedIPs(addressRange string, clientRoutes, nameservers []string) ([]string, error) {
	clientAllowedIPs := []string{}

	_, clientAddressRange, err := net.ParseCIDR(addressRange)
	if err != nil {
		return clientAllowedIPs, fmt.Errorf("could not parse client address range (%s): %s", addressRange, err)
	}

	clientAddressRangeIntersects := false
	if len(clientRoutes) > 0 {
		for _, network := range clientRoutes {
			//networkIntersects
			_, ipnet, err := net.ParseCIDR(network)
			if err == nil {
				if networkIntersects(ipnet, clientAddressRange) {
					clientAddressRangeIntersects = true
				}
			}
		}
		if clientAddressRangeIntersects {
			clientAllowedIPs = clientRoutes
		} else {
			clientAllowedIPs = append(clientAllowedIPs, clientRoutes...)
			clientAllowedIPs = append(clientAllowedIPs, clientAddressRange.String())
		}
	} else {
		clientAllowedIPs = []string{clientAddressRange.String()}
	}
	// add nameserver
	for _, nameserver := range nameservers {
		interSects := false
		_, nameserverIPNet, err := net.ParseCIDR(nameserver + "/32")
		if err == nil {
			for _, clientAllowedIPString := range clientAllowedIPs {
				_, clientAllowedIP, err2 := net.ParseCIDR(clientAllowedIPString)
				if err2 == nil {
					if networkIntersects(nameserverIPNet, clientAllowedIP) {
						interSects = true
					}
				}
			}
		}
		if !interSects {
			clientAllowedIPs = append(clientAllowedIPs, nameserver+"/32")
		}
	}
	return clientAllowedIPs, nil
}
