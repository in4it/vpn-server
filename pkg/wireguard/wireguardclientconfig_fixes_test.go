package wireguard

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
)

type trackingBody struct {
	io.Reader
	closed *atomic.Int32
}

func (b trackingBody) Close() error { b.closed.Add(1); return nil }

type fakeConfigManager struct {
	requests atomic.Int32
	closed   atomic.Int32
}

func (f *fakeConfigManager) RoundTrip(r *http.Request) (*http.Response, error) {
	f.requests.Add(1)
	return &http.Response{
		StatusCode: http.StatusAccepted,
		Body:       trackingBody{Reader: strings.NewReader("OK"), closed: &f.closed},
		Header:     http.Header{},
		Request:    r,
	}, nil
}

func stubConfigManager(t *testing.T) *fakeConfigManager {
	t.Helper()
	fake := &fakeConfigManager{}
	orig := configManagerTransport
	t.Cleanup(func() { configManagerTransport = orig })
	configManagerTransport = fake
	return fake
}

func TestNotifyConfigManagerClosesBody(t *testing.T) {
	fake := stubConfigManager(t)
	storage := &memorystorage.MockMemoryStorage{}
	if err := storage.WriteFile(storage.ConfigPath("clients/user-1.json"), []byte(`{"id":"user-1"}`)); err != nil {
		t.Fatalf("write error: %s", err)
	}
	if err := DeleteClientConfig(storage, "user-1", "user"); err != nil {
		t.Fatalf("DeleteClientConfig error: %s", err)
	}
	if err := ReloadVPNServerConfig(); err != nil {
		t.Fatalf("ReloadVPNServerConfig error: %s", err)
	}
	if fake.requests.Load() != 2 || fake.closed.Load() != 2 {
		t.Fatalf("expected 2 requests with closed bodies, got %d requests and %d closed bodies", fake.requests.Load(), fake.closed.Load())
	}
}

// TestDeleteClientConfigTakesMutex: a delete must not interleave with
// GenerateNewClientConfig, which would write the deleted connection again.
func TestDeleteClientConfigTakesMutex(t *testing.T) {
	stubConfigManager(t)
	storage := &memorystorage.MockMemoryStorage{}
	filename := storage.ConfigPath("clients/user-1.json")
	if err := storage.WriteFile(filename, []byte(`{"id":"user-1"}`)); err != nil {
		t.Fatalf("write error: %s", err)
	}

	clientConfigMutex.Lock()
	done := make(chan error, 1)
	go func() {
		done <- DeleteClientConfig(storage, "user-1", "user")
	}()
	select {
	case err := <-done:
		clientConfigMutex.Unlock()
		t.Fatalf("DeleteClientConfig didn't wait for the client config mutex (err: %v)", err)
	case <-time.After(100 * time.Millisecond):
	}
	if !storage.FileExists(filename) {
		clientConfigMutex.Unlock()
		t.Fatalf("file deleted while the client config mutex was held")
	}
	clientConfigMutex.Unlock()
	if err := <-done; err != nil {
		t.Fatalf("DeleteClientConfig error: %s", err)
	}
	if storage.FileExists(filename) {
		t.Fatalf("file not deleted")
	}
}

func TestGetAllValidPeerConfigsSkipsCorruptFiles(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	valid, _ := json.Marshal(PeerConfig{ID: "user-1", PublicKey: "key-1", Address: "10.189.184.2/32"})
	files := map[string][]byte{
		"clients/user-1.json": valid,
		"clients/user-2.json": []byte(`{corrupt`),
		"clients/user-3.json": []byte(``),
	}
	for name, contents := range files {
		if err := storage.WriteFile(storage.ConfigPath(name), contents); err != nil {
			t.Fatalf("write error: %s", err)
		}
	}
	peerConfigs, err := GetAllValidPeerConfigs(storage)
	if err != nil {
		t.Fatalf("GetAllValidPeerConfigs error: %s", err)
	}
	if len(peerConfigs) != 1 || peerConfigs[0].ID != "user-1" {
		t.Fatalf("expected only the valid peer config, got %+v", peerConfigs)
	}
}
