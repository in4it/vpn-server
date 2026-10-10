package vpn

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
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

func validVPNSetupRequest(t *testing.T, v *VPN) VPNSetupRequest {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/api/vpn/setup/vpn", nil)
	w := httptest.NewRecorder()
	v.vpnSetupHandler(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("get setup: expected 200, got %d", w.Result().StatusCode)
	}
	var setupRequest VPNSetupRequest
	if err := json.NewDecoder(w.Result().Body).Decode(&setupRequest); err != nil {
		t.Fatalf("decode error: %s", err)
	}
	return setupRequest
}

// TestVPNSetupRejectsInvalidInput: invalid values must be rejected before anything is written,
// otherwise every new connection (or the server config) breaks.
func TestVPNSetupRejectsInvalidInput(t *testing.T) {
	tests := map[string]func(req *VPNSetupRequest){
		"client address prefix without slash":    func(req *VPNSetupRequest) { req.ClientAddressPrefix = "32" },
		"client address prefix not a number":     func(req *VPNSetupRequest) { req.ClientAddressPrefix = "/abc" },
		"client address prefix too large":        func(req *VPNSetupRequest) { req.ClientAddressPrefix = "/33" },
		"client address prefix empty":            func(req *VPNSetupRequest) { req.ClientAddressPrefix = "" },
		"client address prefix equal to range":   func(req *VPNSetupRequest) { req.ClientAddressPrefix = "/21" },
		"client address prefix wider than range": func(req *VPNSetupRequest) { req.ClientAddressPrefix = "/16" },
		"client address prefix with plus sign":   func(req *VPNSetupRequest) { req.ClientAddressPrefix = "/+32" },
		"port zero":                              func(req *VPNSetupRequest) { req.Port = "0" },
		"port negative":                          func(req *VPNSetupRequest) { req.Port = "-1" },
		"port too large":                         func(req *VPNSetupRequest) { req.Port = "65536" },
		"port not a number":                      func(req *VPNSetupRequest) { req.Port = "abc" },
		"ipv6 address range":                     func(req *VPNSetupRequest) { req.AddressRange = "fd00::1/64" },
		"external interface shell injection":     func(req *VPNSetupRequest) { req.ExternalInterface = "eth0;reboot" },
		"external interface with spaces":         func(req *VPNSetupRequest) { req.ExternalInterface = "eth0 -j ACCEPT" },
		"external interface too long":            func(req *VPNSetupRequest) { req.ExternalInterface = "abcdefghijklmnop" },
	}
	for name, modify := range tests {
		t.Run(name, func(t *testing.T) {
			storage := &memorystorage.MockMemoryStorage{}
			if _, err := wireguard.CreateNewVPNConfig(storage); err != nil {
				t.Fatalf("create vpn config: %s", err)
			}
			v := &VPN{Storage: storage}
			before, err := storage.ReadFile(storage.ConfigPath(wireguard.VPN_CONFIG_NAME))
			if err != nil {
				t.Fatalf("read vpn config: %s", err)
			}
			req := validVPNSetupRequest(t, v)
			req.Routes = "10.0.0.0/8" // also a change, which must not be written either
			modify(&req)
			payload, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("marshal error: %s", err)
			}
			r := httptest.NewRequest(http.MethodPost, "http://example.com/api/vpn/setup/vpn", bytes.NewBuffer(payload))
			w := httptest.NewRecorder()
			v.vpnSetupHandler(w, r)
			if w.Result().StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Result().StatusCode)
			}
			after, err := storage.ReadFile(storage.ConfigPath(wireguard.VPN_CONFIG_NAME))
			if err != nil {
				t.Fatalf("read vpn config: %s", err)
			}
			if string(before) != string(after) {
				t.Fatalf("vpn config was changed")
			}
		})
	}
}

func TestVPNSetupValidation(t *testing.T) {
	addressRange := netip.MustParsePrefix("10.189.184.1/21")
	for _, prefix := range []string{"/22", "/30", "/32"} {
		if err := validateClientAddressPrefix(addressRange, prefix); err != nil {
			t.Fatalf("expected %s to be valid: %s", prefix, err)
		}
	}
	for _, port := range []string{"1", "51820", "65535"} {
		if _, err := validatePort(port); err != nil {
			t.Fatalf("expected port %s to be valid: %s", port, err)
		}
	}
	for _, iface := range []string{"", "eth0", "ens5", "wlan0.100", "br-1234_ab"} {
		if err := validateExternalInterface(iface); err != nil {
			t.Fatalf("expected interface %s to be valid: %s", iface, err)
		}
	}
	if err := validateAddressRange(addressRange); err != nil {
		t.Fatalf("expected address range to be valid: %s", err)
	}
}

// TestRestartVPNHandlerUpstreamError: an upstream (configmanager) failure must return 502 with the upstream response
func TestRestartVPNHandlerUpstreamError(t *testing.T) {
	l, err := net.Listen("tcp", wireguard.CONFIGMANAGER_URI)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.RequestURI == "/restart-vpn" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("wg-quick failed"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	ts.Listener.Close() //nolint:errcheck
	ts.Listener = l
	ts.Start()
	defer ts.Close() //nolint:errcheck

	v := &VPN{Storage: &memorystorage.MockMemoryStorage{}}
	r := httptest.NewRequest(http.MethodPost, "http://example.com/api/vpn/setup/restart-vpn", nil)
	w := httptest.NewRecorder()
	v.restartVPNHandler(w, r)
	if w.Result().StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Result().StatusCode)
	}
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("read error: %s", err)
	}
	if !strings.Contains(string(body), "wg-quick failed") {
		t.Fatalf("expected upstream response in error, got: %s", body)
	}
}
