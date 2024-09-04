package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func TestCreateUserConnectionDeleteUserFlow(t *testing.T) {
	l, err := net.Listen("tcp", wireguard.CONFIGMANAGER_URI)
	if err != nil {
		t.Fatal(err)
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

	// first create a new user
	storage := &testingmocks.MockMemoryStorage{}

	c, err := newContext(storage, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context")
	}

	err = c.UserStore.Empty()
	if err != nil {
		t.Fatalf("Cannot create context")
	}

	// create a user
	user := users.User{
		Login:    "john",
		Role:     "user",
		Password: "xyz",
	}
	payload, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Cannot create payload: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/users", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	c.usersHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	// generate VPN config
	_, err = wireguard.CreateNewVPNConfig(c.Storage.Client)
	if err != nil {
		t.Fatalf("Cannot create vpn config: %s", err)
	}

	req = httptest.NewRequest("POST", "http://example.com/connections", nil)
	w = httptest.NewRecorder()
	c.connectionsHandler(w, req.WithContext(context.WithValue(context.Background(), CustomValue("user"), user)))

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	connectionID := fmt.Sprintf("%s-1", user.ID)

	userConfigFilename := storage.ConfigPath(path.Join(wireguard.VPN_CLIENTS_DIR, connectionID+".json"))
	configBytes, err := storage.ReadFile(userConfigFilename)
	if err != nil {
		t.Fatalf("could not read user config file")
	}

	var config wireguard.PeerConfig
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		t.Fatalf("could not parse config: %s", err)
	}
	if config.Disabled {
		t.Fatalf("VPN connection is disabled. Expected not disabled")
	}

	req = httptest.NewRequest("GET", "http://example.com/connection/"+connectionID, nil)
	req.SetPathValue("id", connectionID)
	w = httptest.NewRecorder()
	c.connectionsElementHandler(w, req.WithContext(context.WithValue(context.Background(), CustomValue("user"), user)))

	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("readall error: %s", err)
	}
	if !strings.Contains(string(body), "[Interface]") {
		t.Fatalf("output doesn't look like a wireguard client config: %s", body)
	}

	req = httptest.NewRequest("DELETE", "http://example.com/user/"+user.ID, nil)
	req.SetPathValue("id", user.ID)
	w = httptest.NewRecorder()
	c.userHandler(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	_, err = storage.ReadFile(userConfigFilename)
	if err == nil {
		t.Fatalf("could read user config file, expected not to")
	}
}

func TestCreateUser(t *testing.T) {
	// first create a new user
	storage := &testingmocks.MockMemoryStorage{}

	c, err := newContext(storage, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context")
	}

	err = c.UserStore.Empty()
	if err != nil {
		t.Fatalf("Cannot create context")
	}

	// create a user
	payload := []byte(`{"id": "", "login": "testuser", "password": "tttt213", "role": "user", "oidcID": "", "samlID": "", "lastLogin": "", "provisioned": false, "role":"user","samlID":"","suspended":false}`)
	req := httptest.NewRequest("POST", "http://example.com/users", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()
	c.usersHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var user users.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

}
