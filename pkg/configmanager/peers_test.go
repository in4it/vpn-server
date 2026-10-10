package configmanager

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/in4it/go-devops-platform/storage"
	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

// fakePeerSyncer records the calls and detects concurrent peer changes
type fakePeerSyncer struct {
	mu         sync.Mutex
	synced     []wireguard.PeerConfig
	cleanups   int
	updates    int
	inFlight   atomic.Int32
	overlapped atomic.Bool
}

func (f *fakePeerSyncer) enter() {
	if f.inFlight.Add(1) > 1 {
		f.overlapped.Store(true)
	}
	time.Sleep(2 * time.Millisecond)
}
func (f *fakePeerSyncer) exit() { f.inFlight.Add(-1) }

func (f *fakePeerSyncer) SyncClient(storage storage.Iface, peerConfig wireguard.PeerConfig) error {
	f.enter()
	defer f.exit()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.synced = append(f.synced, peerConfig)
	return nil
}
func (f *fakePeerSyncer) Cleanup(storage storage.Iface) error {
	f.enter()
	defer f.exit()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cleanups++
	return nil
}
func (f *fakePeerSyncer) UpdateServer(storage storage.Iface) error {
	f.enter()
	defer f.exit()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updates++
	return nil
}

func stubPeers(t *testing.T) (*fakePeerSyncer, *sync.WaitGroup) {
	t.Helper()
	fake := &fakePeerSyncer{}
	wg := &sync.WaitGroup{}
	origPeers, origRun := peers, runInBackground
	t.Cleanup(func() { peers, runInBackground = origPeers, origRun })
	peers = fake
	runInBackground = func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f()
		}()
	}
	return fake, wg
}

func writePeerConfig(t *testing.T, storage storage.Iface, peerConfig wireguard.PeerConfig) {
	t.Helper()
	out, err := json.Marshal(peerConfig)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	if err := storage.WriteFile(storage.ConfigPath("clients/"+peerConfig.ID+".json"), out); err != nil {
		t.Fatalf("write error: %s", err)
	}
}

// TestPeerChangesAreSerialised: add/delete/cleanup/refresh must never change the
// peers concurrently (a cleanup with a stale view could remove a new peer).
func TestPeerChangesAreSerialised(t *testing.T) {
	fake, wg := stubPeers(t)
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &wireguard.ClientCache{}
	for i := 1; i <= 5; i++ {
		writePeerConfig(t, storage, wireguard.PeerConfig{ID: fmt.Sprintf("user-%d", i), PublicKey: fmt.Sprintf("key-%d", i), Address: fmt.Sprintf("10.189.184.%d/32", i+1)})
	}

	var callers sync.WaitGroup
	for i := 1; i <= 5; i++ {
		callers.Add(3)
		filename := fmt.Sprintf("user-%d.json", i)
		go func() {
			defer callers.Done()
			if err := syncClient(storage, filename, clientCache); err != nil {
				t.Errorf("syncClient error: %s", err)
			}
		}()
		go func() {
			defer callers.Done()
			if err := deleteClient(storage, filename, clientCache); err != nil {
				t.Errorf("deleteClient error: %s", err)
			}
		}()
		go func() {
			defer callers.Done()
			if err := cleanupClients(storage, clientCache); err != nil {
				t.Errorf("cleanupClients error: %s", err)
			}
		}()
	}
	callers.Add(1)
	go func() {
		defer callers.Done()
		if err := refreshAllClientsAndServer(storage, clientCache); err != nil {
			t.Errorf("refreshAllClientsAndServer error: %s", err)
		}
	}()
	callers.Wait()
	wg.Wait()

	if fake.overlapped.Load() {
		t.Fatalf("peer changes ran concurrently")
	}
	if fake.cleanups < 10 {
		t.Fatalf("expected at least 10 cleanups, got %d", fake.cleanups)
	}
}

// TestSyncClientRereadsFile: the background sync uses the connection file as
// it is when the sync runs, not when the request was received.
func TestSyncClientRereadsFile(t *testing.T) {
	fake, wg := stubPeers(t)
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &wireguard.ClientCache{}
	writePeerConfig(t, storage, wireguard.PeerConfig{ID: "user-1", PublicKey: "key-old", Address: "10.189.184.2/32"})

	peerMutex.Lock()
	if err := syncClient(storage, "user-1.json", clientCache); err != nil {
		peerMutex.Unlock()
		t.Fatalf("syncClient error: %s", err)
	}
	writePeerConfig(t, storage, wireguard.PeerConfig{ID: "user-1", PublicKey: "key-new", Address: "10.189.184.2/32"})
	peerMutex.Unlock()
	wg.Wait()

	if len(fake.synced) != 1 || fake.synced[0].PublicKey != "key-new" {
		t.Fatalf("expected the new key to be synced, got %+v", fake.synced)
	}
	if fake.cleanups != 1 {
		t.Fatalf("expected a cleanup after the add, got %d", fake.cleanups)
	}
}

// TestRefreshAllClientsSkipsCorruptFiles: one corrupt connection file must not
// stop the other peers from being added, nor the cleanup (which removes the
// peer of the corrupt file).
func TestRefreshAllClientsSkipsCorruptFiles(t *testing.T) {
	fake, _ := stubPeers(t)
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &wireguard.ClientCache{}
	writePeerConfig(t, storage, wireguard.PeerConfig{ID: "user-1", PublicKey: "key-1", Address: "10.189.184.2/32"})
	if err := storage.WriteFile(storage.ConfigPath("clients/user-2.json"), []byte("{corrupt")); err != nil {
		t.Fatalf("write error: %s", err)
	}
	writePeerConfig(t, storage, wireguard.PeerConfig{ID: "user-3", PublicKey: "key-3", Address: "10.189.184.4/32"})

	if err := refreshAllClientsAndServer(storage, clientCache); err != nil {
		t.Fatalf("refreshAllClientsAndServer error: %s", err)
	}
	if len(fake.synced) != 2 {
		t.Fatalf("expected 2 peers to be synced, got %+v", fake.synced)
	}
	if fake.cleanups != 1 || fake.updates != 1 {
		t.Fatalf("expected cleanup and server update, got %d cleanups, %d updates", fake.cleanups, fake.updates)
	}
	if len(clientCache.Addresses) != 2 {
		t.Fatalf("expected 2 client cache entries, got %+v", clientCache.Addresses)
	}
}

// TestDeleteClientRemovesFromCache: disabled connections are removed from the
// client cache, the cache doesn't grow on every action
func TestDeleteClientRemovesFromCache(t *testing.T) {
	_, wg := stubPeers(t)
	storage := &memorystorage.MockMemoryStorage{}
	clientCache := &wireguard.ClientCache{}
	writePeerConfig(t, storage, wireguard.PeerConfig{ID: "user-1", PublicKey: "key-1", Address: "10.189.184.2/32"})
	for i := 0; i < 3; i++ {
		if err := syncClient(storage, "user-1.json", clientCache); err != nil {
			t.Fatalf("syncClient error: %s", err)
		}
	}
	wg.Wait()
	if len(clientCache.Addresses) != 1 {
		t.Fatalf("expected 1 cache entry, got %d", len(clientCache.Addresses))
	}
	writePeerConfig(t, storage, wireguard.PeerConfig{ID: "user-1", PublicKey: "key-1", Address: "10.189.184.2/32", Disabled: true})
	if err := deleteClient(storage, "user-1.json", clientCache); err != nil {
		t.Fatalf("deleteClient error: %s", err)
	}
	wg.Wait()
	if len(clientCache.Addresses) != 0 {
		t.Fatalf("expected empty cache, got %+v", clientCache.Addresses)
	}
}
