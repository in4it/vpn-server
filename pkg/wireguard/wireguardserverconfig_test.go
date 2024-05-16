package wireguard

import (
	"strings"
	"testing"

	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
)

func TestWriteWireGuardServerConfig(t *testing.T) {
	storage := &testingmocks.MockMemoryStorage{}

	// first create a new vpn config
	vpnconfig, err := CreateNewVPNConfig(storage)
	if err != nil {
		t.Fatalf("CreateNewVPNConfig error: %s", err)
	}

	vpnconfigFile, err := generateWireGuardServerConfig(storage)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if !strings.Contains(string(vpnconfigFile), vpnconfig.AddressRange.String()) {
		t.Fatalf("couldn't find address range in vpn config file: %s", vpnconfig.AddressRange.String())
	}
}
