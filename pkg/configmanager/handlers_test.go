package configmanager

import (
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/in4it/go-devops-platform/storage"
	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

type vpnCalls struct {
	stopped, started, refreshed int
}

func stubVPN(t *testing.T) *vpnCalls {
	calls := &vpnCalls{}
	origStop, origStart, origRefresh := vpnStop, vpnStart, vpnRefreshClients
	t.Cleanup(func() { vpnStop, vpnStart, vpnRefreshClients = origStop, origStart, origRefresh })
	vpnStop = func() error { calls.stopped++; return nil }
	vpnStart = func(storage.Iface) error { calls.started++; return nil }
	vpnRefreshClients = func(storage.Iface, *wireguard.ClientCache) error { calls.refreshed++; return nil }
	return calls
}

func restartVPN(c *ConfigManager) int {
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/restart-vpn", nil)
	w := httptest.NewRecorder()
	c.restartVpn(w, req)
	return w.Result().StatusCode
}

func newTestConfigManager(t *testing.T) *ConfigManager {
	storage := &memorystorage.MockMemoryStorage{}
	if _, err := wireguard.CreateNewVPNConfig(storage); err != nil {
		t.Fatalf("CreateNewVPNConfig error: %s", err)
	}
	return &ConfigManager{Storage: storage, ClientCache: &wireguard.ClientCache{}}
}

// TestRestartVPNBrokenServerTemplate: a server template that doesn't render
// must not take the vpn down.
func TestRestartVPNBrokenServerTemplate(t *testing.T) {
	calls := stubVPN(t)
	c := newTestConfigManager(t)
	templateFile := c.Storage.ConfigPath(path.Join(wireguard.WIREGUARD_TEMPLATE_DIR, wireguard.WIREGUARD_TEMPLATE_SERVER))
	if err := c.Storage.WriteFile(templateFile, []byte("{{ .Foo }}")); err != nil {
		t.Fatalf("write template error: %s", err)
	}
	if status := restartVPN(c); status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
	if calls.stopped != 0 || calls.started != 0 {
		t.Fatalf("vpn should not be stopped or started with a broken template: %+v", calls)
	}
}

func TestRestartVPN(t *testing.T) {
	calls := stubVPN(t)
	c := newTestConfigManager(t)
	if status := restartVPN(c); status != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", status)
	}
	if calls.stopped != 1 || calls.started != 1 || calls.refreshed != 1 {
		t.Fatalf("expected stop, start and refresh once: %+v", calls)
	}
}
