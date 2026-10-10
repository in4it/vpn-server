package wireguard

import (
	"errors"
	"strings"
	"testing"

	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
)

func stubDefaultGw(t *testing.T, iface string, err error) {
	t.Helper()
	orig := getInterfaceDefaultGw
	t.Cleanup(func() { getInterfaceDefaultGw = orig })
	getInterfaceDefaultGw = func() (string, error) { return iface, err }
}

func TestGetExternalInterface(t *testing.T) {
	tests := []struct {
		name      string
		vpnConfig VPNConfig
		detected  string
		detectErr error
		want      string
		wantErr   bool
	}{
		{name: "configured", vpnConfig: VPNConfig{ExternalInterface: "ens4"}, want: "ens4"},
		{name: "empty is detected again", vpnConfig: VPNConfig{}, detected: "ens5", want: "ens5"},
		{name: "empty and not detected", vpnConfig: VPNConfig{}, detectErr: errors.New("no default route"), wantErr: true},
		{name: "empty detected", vpnConfig: VPNConfig{}, detected: "", wantErr: true},
		{name: "nat disabled", vpnConfig: VPNConfig{DisableNAT: true}, detectErr: errors.New("no default route"), want: ""},
		{name: "shell injection", vpnConfig: VPNConfig{ExternalInterface: "eth0; touch /tmp/x"}, wantErr: true},
		{name: "too long", vpnConfig: VPNConfig{ExternalInterface: "abcdefghijklmnop"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubDefaultGw(t, tt.detected, tt.detectErr)
			got, err := getExternalInterface(tt.vpnConfig)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGenerateServerConfigEmptyExternalInterface: the server config must never
// contain "-o  -j MASQUERADE"
func TestGenerateServerConfigEmptyExternalInterface(t *testing.T) {
	stubConfigManager(t)
	storage := &memorystorage.MockMemoryStorage{}
	vpnConfig, err := CreateNewVPNConfig(storage)
	if err != nil {
		t.Fatalf("CreateNewVPNConfig error: %s", err)
	}
	vpnConfig.ExternalInterface = ""
	if err := WriteVPNConfig(storage, vpnConfig); err != nil {
		t.Fatalf("WriteVPNConfig error: %s", err)
	}

	stubDefaultGw(t, "", errors.New("no default route"))
	if out, err := generateWireGuardServerConfig(storage); err == nil {
		t.Fatalf("expected error, got config:\n%s", out)
	}

	stubDefaultGw(t, "ens4", nil)
	out, err := generateWireGuardServerConfig(storage)
	if err != nil {
		t.Fatalf("generate error: %s", err)
	}
	if !strings.Contains(string(out), "-o ens4 -j MASQUERADE") {
		t.Fatalf("expected detected interface in config:\n%s", out)
	}
}
