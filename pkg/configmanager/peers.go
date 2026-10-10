package configmanager

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

// peerSyncer applies peer changes to the wireguard interface
type peerSyncer interface {
	// SyncClient adds or removes (when disabled) the peer
	SyncClient(storage storage.Iface, peerConfig wireguard.PeerConfig) error
	// Cleanup removes all peers that don't have a valid, enabled connection file
	Cleanup(storage storage.Iface) error
	// UpdateServer updates the server private key
	UpdateServer(storage storage.Iface) error
}

var (
	// peerMutex serialises all changes to the peers of the wireguard interface.
	// The connection files are (re)read while holding it, so an operation never
	// works with a stale view: e.g. a cleanup that read the connection files
	// before a new key was written can't remove the new peer anymore.
	peerMutex sync.Mutex
	// variables so tests can stub them
	peers           peerSyncer = newPeerSyncer()
	runInBackground            = func(f func()) { go f() }
)

// syncClient handles an add request for a connection file. The peer changes are
// done in the background (serialised), so the rest-server gets a quick response.
func syncClient(storage storage.Iface, filename string, clientCache *wireguard.ClientCache) error {
	if _, err := wireguard.GetPeerConfigByFilename(storage, filename); err != nil {
		return fmt.Errorf("getClientFile error: %s", err)
	}
	runInBackground(func() {
		peerMutex.Lock()
		defer peerMutex.Unlock()
		syncPeerFromFile(storage, filename, clientCache)
		cleanupPeers(storage, clientCache)
	})
	return nil
}

// deleteClient handles a delete request (the connection file is disabled)
func deleteClient(storage storage.Iface, filename string, clientCache *wireguard.ClientCache) error {
	if _, err := wireguard.GetPeerConfigByFilename(storage, filename); err != nil {
		return fmt.Errorf("getClientFile error: %s", err)
	}
	runInBackground(func() {
		peerMutex.Lock()
		defer peerMutex.Unlock()
		if !syncPeerFromFile(storage, filename, clientCache) {
			// the file is gone or corrupt: remove all peers without a valid config
			cleanupPeers(storage, clientCache)
		}
	})
	return nil
}

// cleanupClients handles a cleanup request (connection files were removed)
func cleanupClients(storage storage.Iface, clientCache *wireguard.ClientCache) error {
	runInBackground(func() {
		peerMutex.Lock()
		defer peerMutex.Unlock()
		cleanupPeers(storage, clientCache)
	})
	return nil
}

// syncPeerFromFile reads the connection file and adds or removes (if disabled)
// the peer. Must be called with peerMutex held. Returns false if the connection
// file couldn't be read.
func syncPeerFromFile(storage storage.Iface, filename string, clientCache *wireguard.ClientCache) bool {
	peerConfig, err := wireguard.GetPeerConfigByFilename(storage, filename)
	if err != nil {
		log.Printf("Warning: could not read connection file %s: %s", filename, err)
		return false
	}
	if err := wireguard.UpdateClientCache(peerConfig, clientCache); err != nil {
		log.Printf("Warning: update client cache error (%s): %s", filename, err)
	}
	if err := peers.SyncClient(storage, peerConfig); err != nil {
		log.Printf("Error: could not sync peer (%s): %s", filename, err)
	}
	return true
}

// cleanupPeers removes the peers without valid connection file and refreshes
// the client cache. Must be called with peerMutex held.
func cleanupPeers(storage storage.Iface, clientCache *wireguard.ClientCache) {
	if err := peers.Cleanup(storage); err != nil {
		log.Printf("Error: cleanup error: %s", err)
	}
	peerConfigs, err := wireguard.GetAllValidPeerConfigs(storage)
	if err != nil {
		log.Printf("Warning: could not refresh client cache: %s", err)
		return
	}
	if err := wireguard.ReplaceClientCache(peerConfigs, clientCache); err != nil {
		log.Printf("Warning: client cache error: %s", err)
	}
}

// refreshAllClientsAndServer syncs all peers with the connection files, removes
// peers without a valid connection file and updates the server key. Connection
// files that can't be parsed are skipped (their peer is removed by the cleanup).
func refreshAllClientsAndServer(storage storage.Iface, clientCache *wireguard.ClientCache) error {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	peerConfigs, err := wireguard.GetAllValidPeerConfigs(storage)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // directory doesn't exist, so no configs to be read
		}
		return fmt.Errorf("can not list clients: %s", err)
	}

	errs := []error{}
	if err := wireguard.ReplaceClientCache(peerConfigs, clientCache); err != nil {
		errs = append(errs, fmt.Errorf("update client cache error: %s", err))
	}
	for _, peerConfig := range peerConfigs {
		if err := peers.SyncClient(storage, peerConfig); err != nil {
			errs = append(errs, fmt.Errorf("SyncClients error (%s): %s", peerConfig.ID, err))
		}
	}
	if err := peers.Cleanup(storage); err != nil {
		errs = append(errs, fmt.Errorf("cleanup error: %s", err))
	}
	if err := peers.UpdateServer(storage); err != nil {
		errs = append(errs, fmt.Errorf("UpdateServer error: %s", err))
	}
	return errors.Join(errs...)
}
