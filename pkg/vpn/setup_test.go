package vpn

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func postTemplates(t *testing.T, v *VPN, req TemplateSetupRequest) int {
	t.Helper()
	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	r := httptest.NewRequest(http.MethodPost, "http://example.com/api/vpn/setup/templates", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	v.templateSetupHandler(w, r)
	return w.Result().StatusCode
}

// TestTemplateSetupRejectsInvalidTemplates: a broken template must not be
// saved, as it would break every config download (client) or stop the vpn
// from starting (server).
func TestTemplateSetupRejectsInvalidTemplates(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	v := &VPN{Storage: storage}

	clientTemplate, err := wireguard.GetClientTemplate(storage)
	if err != nil {
		t.Fatalf("get client template: %s", err)
	}
	serverTemplate, err := wireguard.GetServerTemplate(storage)
	if err != nil {
		t.Fatalf("get server template: %s", err)
	}

	tests := map[string]TemplateSetupRequest{
		"broken client template":                          {ClientTemplate: "{{ .Adress }}", ServerTemplate: string(serverTemplate)},
		"broken server template":                          {ClientTemplate: string(clientTemplate), ServerTemplate: "{{ .Foo }}"},
		"valid client change with broken server template": {ClientTemplate: string(clientTemplate) + "\n# changed", ServerTemplate: "{{ .Address"},
	}
	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			if status := postTemplates(t, v, req); status != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", status)
			}
			gotClient, _ := wireguard.GetClientTemplate(storage)
			gotServer, _ := wireguard.GetServerTemplate(storage)
			if string(gotClient) != string(clientTemplate) {
				t.Fatalf("client template was changed")
			}
			if string(gotServer) != string(serverTemplate) {
				t.Fatalf("server template was changed")
			}
		})
	}
}

func TestTemplateSetupSavesValidServerTemplate(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	v := &VPN{Storage: storage}
	clientTemplate, _ := wireguard.GetClientTemplate(storage)
	serverTemplate, _ := wireguard.GetServerTemplate(storage)
	newServerTemplate := string(serverTemplate) + "\n# custom comment\n"
	if status := postTemplates(t, v, TemplateSetupRequest{ClientTemplate: string(clientTemplate), ServerTemplate: newServerTemplate}); status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	gotServer, _ := wireguard.GetServerTemplate(storage)
	if string(gotServer) != newServerTemplate {
		t.Fatalf("server template was not saved")
	}
}
