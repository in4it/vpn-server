package wireguard

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
)

func TestWriteWireGuardServerConfig(t *testing.T) {
	var (
		l   net.Listener
		err error
	)
	for {
		l, err = net.Listen("tcp", CONFIGMANAGER_URI)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "address already in use") {
				t.Fatal(err)
			}
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if r.RequestURI == "/refresh-clients" {
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("OK"))
				return
			}
			if r.RequestURI == "/refresh-server-config" {
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("OK"))
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))

	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()
	defer l.Close()

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
