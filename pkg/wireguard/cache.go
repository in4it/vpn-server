package wireguard

import (
	"fmt"
	"net"
)

// UpdateClientCache adds or updates the cache entry for a connection. The cache
// is keyed by connection ID (peerConfig.ID). Disabled connections are removed
// from the cache, as they can't send traffic anymore.
func UpdateClientCache(peerConfig PeerConfig, clientCache *ClientCache) error {
	if peerConfig.Disabled {
		RemoveFromClientCache(peerConfig.ID, clientCache)
		return nil
	}
	entry, err := newClientCacheEntry(peerConfig)
	if err != nil {
		return err
	}

	clientCache.mu.Lock()
	defer clientCache.mu.Unlock()

	for k := range clientCache.Addresses {
		if clientCache.Addresses[k].ConnectionID == peerConfig.ID {
			clientCache.Addresses[k] = entry
			return nil
		}
	}
	clientCache.Addresses = append(clientCache.Addresses, entry)
	return nil
}

// RemoveFromClientCache removes the cache entry of a connection
func RemoveFromClientCache(connectionID string, clientCache *ClientCache) {
	clientCache.mu.Lock()
	defer clientCache.mu.Unlock()
	addresses := clientCache.Addresses[:0]
	for _, address := range clientCache.Addresses {
		if address.ConnectionID != connectionID {
			addresses = append(addresses, address)
		}
	}
	// clear the tail, so no stale entries are kept in the backing array
	for k := len(addresses); k < len(clientCache.Addresses); k++ {
		clientCache.Addresses[k] = ClientCacheAddresses{}
	}
	clientCache.Addresses = addresses
}

// ReplaceClientCache replaces the full cache with the (enabled) peer configs
func ReplaceClientCache(peerConfigs []PeerConfig, clientCache *ClientCache) error {
	addresses := make([]ClientCacheAddresses, 0, len(peerConfigs))
	var lastErr error
	for _, peerConfig := range peerConfigs {
		if peerConfig.Disabled {
			continue
		}
		entry, err := newClientCacheEntry(peerConfig)
		if err != nil {
			lastErr = err
			continue
		}
		addresses = append(addresses, entry)
	}
	clientCache.mu.Lock()
	clientCache.Addresses = addresses
	clientCache.mu.Unlock()
	return lastErr
}

// getClientIDByIP returns the client (user) ID that has the ip address assigned
func (clientCache *ClientCache) getClientIDByIP(ip net.IP) string {
	clientCache.mu.RLock()
	defer clientCache.mu.RUnlock()
	clientID := ""
	for _, address := range clientCache.Addresses {
		if address.Address.Contains(ip) {
			clientID = address.ClientID
		}
	}
	return clientID
}

func newClientCacheEntry(peerConfig PeerConfig) (ClientCacheAddresses, error) {
	_, peerConfigAddressParsed, err := net.ParseCIDR(peerConfig.Address)
	if err != nil {
		return ClientCacheAddresses{}, fmt.Errorf("cannot parse peerConfig's address: %s", err)
	}
	clientID, _, err := getClientIDAndConfigID(peerConfig.ID)
	if err != nil {
		return ClientCacheAddresses{}, fmt.Errorf("can't parse peer config ID (%s): %s", peerConfig.ID, err)
	}
	return ClientCacheAddresses{
		Address:      *peerConfigAddressParsed,
		ClientID:     clientID,
		ConnectionID: peerConfig.ID,
	}, nil
}
