package wireguard

import (
	"fmt"
	"net"
	"sync"
	"testing"
)

func TestUpdateClientCacheDedupesByConnectionID(t *testing.T) {
	clientCache := &ClientCache{}
	peerConfig := PeerConfig{ID: "1-2-3-4-1", Address: "10.189.184.2/32"}
	for i := 0; i < 3; i++ {
		if err := UpdateClientCache(peerConfig, clientCache); err != nil {
			t.Fatalf("UpdateClientCache error: %s", err)
		}
	}
	if len(clientCache.Addresses) != 1 {
		t.Fatalf("expected 1 cache entry, got %d", len(clientCache.Addresses))
	}
	if clientCache.Addresses[0].ClientID != "1-2-3-4" || clientCache.Addresses[0].ConnectionID != "1-2-3-4-1" {
		t.Fatalf("unexpected cache entry: %+v", clientCache.Addresses[0])
	}

	// address change updates the entry
	peerConfig.Address = "10.189.184.9/32"
	if err := UpdateClientCache(peerConfig, clientCache); err != nil {
		t.Fatalf("UpdateClientCache error: %s", err)
	}
	if len(clientCache.Addresses) != 1 || clientCache.getClientIDByIP(net.ParseIP("10.189.184.9")) != "1-2-3-4" {
		t.Fatalf("expected the address to be updated: %+v", clientCache.Addresses)
	}
	if clientCache.getClientIDByIP(net.ParseIP("10.189.184.2")) != "" {
		t.Fatalf("old address should not be in the cache anymore")
	}

	// second connection of the same user
	if err := UpdateClientCache(PeerConfig{ID: "1-2-3-4-2", Address: "10.189.184.3/32"}, clientCache); err != nil {
		t.Fatalf("UpdateClientCache error: %s", err)
	}
	if len(clientCache.Addresses) != 2 {
		t.Fatalf("expected 2 cache entries, got %d", len(clientCache.Addresses))
	}

	// disabled connection is removed
	if err := UpdateClientCache(PeerConfig{ID: "1-2-3-4-2", Address: "10.189.184.3/32", Disabled: true}, clientCache); err != nil {
		t.Fatalf("UpdateClientCache error: %s", err)
	}
	if len(clientCache.Addresses) != 1 || clientCache.getClientIDByIP(net.ParseIP("10.189.184.3")) != "" {
		t.Fatalf("disabled connection should be removed from the cache: %+v", clientCache.Addresses)
	}

	RemoveFromClientCache("1-2-3-4-1", clientCache)
	if len(clientCache.Addresses) != 0 {
		t.Fatalf("expected empty cache, got %+v", clientCache.Addresses)
	}
}

func TestReplaceClientCache(t *testing.T) {
	clientCache := &ClientCache{}
	if err := UpdateClientCache(PeerConfig{ID: "deleted-1", Address: "10.189.184.5/32"}, clientCache); err != nil {
		t.Fatalf("UpdateClientCache error: %s", err)
	}
	err := ReplaceClientCache([]PeerConfig{
		{ID: "user-1", Address: "10.189.184.2/32"},
		{ID: "user-2", Address: "10.189.184.3/32", Disabled: true},
	}, clientCache)
	if err != nil {
		t.Fatalf("ReplaceClientCache error: %s", err)
	}
	if len(clientCache.Addresses) != 1 || clientCache.Addresses[0].ConnectionID != "user-1" {
		t.Fatalf("unexpected cache: %+v", clientCache.Addresses)
	}
}

// TestClientCacheConcurrentAccess must be run with -race: the packet logger
// reads the cache and the packet log settings while the configmanager handlers
// update them.
func TestClientCacheConcurrentAccess(t *testing.T) {
	clientCache := &ClientCache{}
	vpnConfig := &VPNConfig{PacketLogsTypes: map[string]bool{}}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				id := fmt.Sprintf("user-%d", j%5)
				_ = UpdateClientCache(PeerConfig{ID: id, Address: fmt.Sprintf("10.189.184.%d/32", j%5+2)}, clientCache)
				if j%7 == 0 {
					RemoveFromClientCache(id, clientCache)
				}
				SetPacketLogConfig(vpnConfig, j%2 == 0, map[string]bool{"tcp": j%3 == 0})
			}
		}(i)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				clientCache.getClientIDByIP(net.ParseIP("10.189.184.3"))
				enabled, types := getPacketLogConfig(vpnConfig)
				_ = enabled
				_ = types["tcp"]
			}
		}()
	}
	wg.Wait()
}

func TestSetPacketLogConfig(t *testing.T) {
	vpnConfig := &VPNConfig{}
	types := map[string]bool{"dns": true}
	if !SetPacketLogConfig(vpnConfig, true, types) {
		t.Fatalf("expected start when enabling packet logs")
	}
	if SetPacketLogConfig(vpnConfig, true, types) {
		t.Fatalf("expected no start when packet logs are already enabled")
	}
	types["tcp"] = true // must not change the shared config
	if _, got := getPacketLogConfig(vpnConfig); got["tcp"] {
		t.Fatalf("packet log types should be copied")
	}
	if SetPacketLogConfig(vpnConfig, false, nil) {
		t.Fatalf("expected no start when disabling packet logs")
	}
	if isPacketLogEnabled(vpnConfig) {
		t.Fatalf("expected packet logs to be disabled")
	}
}
