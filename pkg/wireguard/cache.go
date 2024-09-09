package wireguard

import (
	"fmt"
	"net"
)

func UpdateClientCache(peerConfig PeerConfig, clientCache *ClientCache) error {
	_, peerConfigAddressParsed, err := net.ParseCIDR(peerConfig.Address)
	if err != nil {
		return fmt.Errorf("cannot parse peerConfig's address: %s", err)
	}
	found := false
	for k, addressesItem := range clientCache.Addresses {
		if addressesItem.ClientID == peerConfig.ID {
			found = true
			if addressesItem.Address.String() != peerConfig.Address {
				clientCache.Addresses[k].Address = *peerConfigAddressParsed
				return nil
			}
		}
	}

	if !found {
		clientCache.Addresses = append(clientCache.Addresses, ClientCacheAddresses{
			Address:  *peerConfigAddressParsed,
			ClientID: peerConfig.ID,
		})
	}

	return nil
}
