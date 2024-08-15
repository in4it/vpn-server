package wireguard

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
)

func TestGetNextFreeIPFromList(t *testing.T) {
	ip := net.ParseIP("10.0.0.1")
	ipList := []string{"10.0.0.2", "10.0.0.3"}
	nextIP, err := getNextFreeIPFromList(ip, ipList)
	if err != nil {
		t.Errorf("next IP error: %s", err)
	}
	if nextIP.String() != "10.0.0.4" {
		t.Errorf("wrong ip outputted: %s", nextIP)
	}
}

func TestHasClientUserID(t *testing.T) {
	filename := "1-2-3-4-0.json"
	if !HasClientUserID(filename, "1-2-3-4") {
		t.Errorf("wrong expected return (got false, should be true)")
	}
	if HasClientUserID(filename, "1-2-3-5") {
		t.Errorf("wrong expected return (got false, should be true)")
	}
}

func TestGetConfigNumberFromConnectionFile(t *testing.T) {
	filename := "1-2-3-4-0.json"
	if res, err := getConfigNumberFromConnectionFile(filename); err != nil || res != 0 {
		t.Errorf("wrong result. Error: %v - res %v", err.Error(), res)
	}
	filename = "1.2.3.-3.json"
	if res, err := getConfigNumberFromConnectionFile(filename); err != nil || res != 3 {
		t.Errorf("wrong result. Error: %v - res %v", err.Error(), res)
	}
	filename = "1.2.3.-123456.json"
	if res, err := getConfigNumberFromConnectionFile(filename); err != nil || res != 123456 {
		t.Errorf("wrong result. Error: %v - res %v", err.Error(), res)
	}
}

func TestWriteConfig(t *testing.T) {
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
	prefix, err := netip.ParsePrefix(DEFAULT_VPN_PREFIX)
	if err != nil {
		t.Errorf("ParsePrefix error: %s", err)
	}
	if vpnconfig.AddressRange.String() != prefix.String() {
		t.Fatalf("wrong AddressRange: %s vs %s", vpnconfig.AddressRange.String(), prefix.String())
	}
	// generate the peerconfig
	peerConfig, err := NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}

	out, err := GenerateNewClientConfig(storage, peerConfig.ID, "2-2-2-2")
	if err != nil {
		t.Fatalf("GenerateNewClientConfig error: %s", err)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasSuffix(strings.TrimSpace(line), ",") {
			t.Fatalf("line ended with comma: '%s'", line)
		}
	}
}

func TestWriteConfigMultipleClients(t *testing.T) {
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
	prefix, err := netip.ParsePrefix(DEFAULT_VPN_PREFIX)
	if err != nil {
		t.Errorf("ParsePrefix error: %s", err)
	}
	if vpnconfig.AddressRange.String() != prefix.String() {
		t.Fatalf("wrong AddressRange: %s vs %s", vpnconfig.AddressRange.String(), prefix.String())
	}

	// generate the peerconfig
	peerConfig1, err := NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}

	peerConfig2, err := NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}
	if len(peerConfig1.ServerAllowedIPs) == 0 {
		t.Fatalf("server allowed IPs is empty")
	}
	if len(peerConfig2.ServerAllowedIPs) == 0 {
		t.Fatalf("server allowed IPs is empty")
	}
	if peerConfig1.ServerAllowedIPs[0] == peerConfig2.ServerAllowedIPs[0] {
		t.Fatalf("cant have the same IPs: %s", peerConfig1.ServerAllowedIPs[0])
	}

}

func TestCreateAndDeleteAllClientConfig(t *testing.T) {
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
	prefix, err := netip.ParsePrefix(DEFAULT_VPN_PREFIX)
	if err != nil {
		t.Errorf("ParsePrefix error: %s", err)
	}
	if vpnconfig.AddressRange.String() != prefix.String() {
		t.Fatalf("wrong AddressRange: %s vs %s", vpnconfig.AddressRange.String(), prefix.String())
	}
	// generate the peerconfig
	peerConfig, err := NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}

	if peerConfig.PublicKey != "" {
		t.Fatalf("public key already found in peerconfig")
	}

	_, err = GenerateNewClientConfig(storage, peerConfig.ID, "2-2-2-2")
	if err != nil {
		t.Fatalf("GenerateNewClientConfig error: %s", err)
	}

	writtenPeerconfig, ok := storage.Data["config/clients/2-2-2-2-1.json"]
	if !ok {
		t.Fatalf("couldn't find peer config file written in storage")
	}

	err = json.Unmarshal(writtenPeerconfig, &peerConfig)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if peerConfig.PublicKey == "" {
		t.Errorf("Public key not found in client config")
	}

	err = DeleteAllClientConfigs(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("DeleteAllClientConfigs error: %s", err)
	}
	_, ok = storage.Data["config/clients/2-2-2-2-1.json"]
	if ok {
		t.Fatalf("still can find config for client in storage")
	}
}
func TestCreateAndDeleteClientConfig(t *testing.T) {
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
	prefix, err := netip.ParsePrefix(DEFAULT_VPN_PREFIX)
	if err != nil {
		t.Errorf("ParsePrefix error: %s", err)
	}
	if vpnconfig.AddressRange.String() != prefix.String() {
		t.Fatalf("wrong AddressRange: %s vs %s", vpnconfig.AddressRange.String(), prefix.String())
	}
	// generate the peerconfig
	peerConfig, err := NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}

	if peerConfig.PublicKey != "" {
		t.Fatalf("public key already found in peerconfig")
	}

	_, err = GenerateNewClientConfig(storage, peerConfig.ID, "2-2-2-2")
	if err != nil {
		t.Fatalf("GenerateNewClientConfig error: %s", err)
	}

	writtenPeerconfig, ok := storage.Data["config/clients/2-2-2-2-1.json"]
	if !ok {
		t.Fatalf("couldn't find peer config file written in storage")
	}

	err = json.Unmarshal(writtenPeerconfig, &peerConfig)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if peerConfig.PublicKey == "" {
		t.Errorf("Public key not found in client config")
	}

	err = DeleteClientConfig(storage, "2-2-2-2-1", "2-2-2-2")
	if err != nil {
		t.Fatalf("DeleteAllClientConfigs error: %s", err)
	}
	_, ok = storage.Data["config/clients/2-2-2-2-1.json"]
	if ok {
		t.Fatalf("still can find config for client in storage")
	}
}

func TestCreateAndDisableAllClientConfig(t *testing.T) {
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
	prefix, err := netip.ParsePrefix(DEFAULT_VPN_PREFIX)
	if err != nil {
		t.Errorf("ParsePrefix error: %s", err)
	}
	if vpnconfig.AddressRange.String() != prefix.String() {
		t.Fatalf("wrong AddressRange: %s vs %s", vpnconfig.AddressRange.String(), prefix.String())
	}
	// generate the peerconfig
	peerConfig, err := NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}

	if peerConfig.PublicKey != "" {
		t.Fatalf("public key already found in peerconfig")
	}

	_, err = GenerateNewClientConfig(storage, peerConfig.ID, "2-2-2-2")
	if err != nil {
		t.Fatalf("GenerateNewClientConfig error: %s", err)
	}

	writtenPeerconfig, ok := storage.Data["config/clients/2-2-2-2-1.json"]
	if !ok {
		t.Fatalf("couldn't find peer config file written in storage")
	}

	err = json.Unmarshal(writtenPeerconfig, &peerConfig)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if peerConfig.Disabled {
		t.Errorf("Peer config is disabled")
	}

	err = DisableAllClientConfigs(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("DisableAllClientConfigs error: %s", err)
	}

	writtenPeerconfig, ok = storage.Data["config/clients/2-2-2-2-1.json"]
	if !ok {
		t.Fatalf("couldn't find peer config file written in storage")
	}

	if !ok {
		t.Fatalf("couldn't find peer config file written in storage")
	}

	err = json.Unmarshal(writtenPeerconfig, &peerConfig)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if !peerConfig.Disabled {
		t.Errorf("peer config not disabled")
	}

	err = ReactivateAllClientConfigs(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("DisableAllClientConfigs error: %s", err)
	}
	writtenPeerconfig, ok = storage.Data["config/clients/2-2-2-2-1.json"]
	if !ok {
		t.Fatalf("couldn't find peer config file written in storage")
	}

	if !ok {
		t.Fatalf("couldn't find peer config file written in storage")
	}

	err = json.Unmarshal(writtenPeerconfig, &peerConfig)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if peerConfig.Disabled {
		t.Errorf("peer config still disabled")
	}

}

func TestUpdateClientConfig(t *testing.T) {
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
	prefix, err := netip.ParsePrefix(DEFAULT_VPN_PREFIX)
	if err != nil {
		t.Errorf("ParsePrefix error: %s", err)
	}
	if vpnconfig.AddressRange.String() != prefix.String() {
		t.Fatalf("wrong AddressRange: %s vs %s", vpnconfig.AddressRange.String(), prefix.String())
	}
	// generate the peerconfig
	peerConfig, err := NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}

	if peerConfig.ClientAllowedIPs[0] != "0.0.0.0/0" {
		t.Fatalf("wrong client allowed ips")
	}

	newClientRoutes := []string{"1.2.3.4/32"}
	vpnconfig.ClientRoutes = newClientRoutes
	err = WriteVPNConfig(storage, vpnconfig)
	if err != nil {
		t.Fatalf("WriteVPNConfig error: %s", err)
	}
	err = UpdateClientsConfig(storage)
	if err != nil {
		t.Fatalf("UpdateClientsConfig error: %s", err)
	}

	peerConfig, err = NewEmptyClientConfig(storage, "2-2-2-2")
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)
	}

	if peerConfig.ClientAllowedIPs[0] != "1.2.3.4/32" {
		t.Fatalf("wrong client allowed ips")
	}

}
