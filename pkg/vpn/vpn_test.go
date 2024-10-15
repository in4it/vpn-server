package vpn

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

	"github.com/in4it/go-devops-platform/auth/provisioning/scim"
	"github.com/in4it/go-devops-platform/rest"
	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/go-devops-platform/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

const USERSTORE_MAX_USERS = 1000

func TestSCIMCreateUserConnectionDeleteUserFlow(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	userStore, err := users.NewUserStore(storage, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user store: %s", err)
	}
	userStore.Empty()
	if err != nil {
		t.Fatalf("cannot empty user store")
	}
	s := scim.New(storage, userStore, "token", wireguard.DisableAllClientConfigs, wireguard.ReactivateAllClientConfigs)

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

	// create a user
	payload := scim.PostUserRequest{
		UserName: "john@domain.inv",
		Name: scim.Name{
			GivenName:  "John",
			FamilyName: "Doe",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("cannot marshal payload: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/api/scim/v2/Users?", bytes.NewBuffer(payloadBytes))
	w := httptest.NewRecorder()
	s.PostUsersHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var postUserRequest scim.PostUserRequest
	err = json.NewDecoder(resp.Body).Decode(&postUserRequest)
	if err != nil {
		t.Fatalf("Could not decode output: %s", err)
	}

	if postUserRequest.Id == "" {
		t.Fatalf("id is empty: %s", err)
	}

	user, err := s.UserStore.GetUserByID(postUserRequest.Id)
	if err != nil {
		t.Fatalf("Cannot get newly created user: %s", err)
	}

	// generate VPN config
	_, err = wireguard.CreateNewVPNConfig(storage)
	if err != nil {
		t.Fatalf("Cannot create vpn config: %s", err)
	}

	peerConfig, err := wireguard.NewEmptyClientConfig(storage, user.ID)
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)

	}

	if peerConfig.Disabled {
		t.Fatalf("VPN connection is disabled. Expected not disabled")
	}

	connectionID := fmt.Sprintf("%s-1", user.ID)
	userConfigFilename := storage.ConfigPath(path.Join(wireguard.VPN_CLIENTS_DIR, connectionID+".json"))
	_, err = storage.ReadFile(userConfigFilename)
	if err != nil {
		t.Fatalf("could not read user config file")
	}

	req = httptest.NewRequest("DELETE", "http://example.com/api/scim/v2/Users/"+user.ID, nil)
	req.SetPathValue("id", user.ID)
	w = httptest.NewRecorder()
	s.DeleteUserHandler(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	_, err = storage.ReadFile(userConfigFilename)
	if err == nil {
		t.Fatalf("could read user config file. Expected not to be able to read it (should have been deleted)")
	}
}
func TestCreateUserConnectionSuspendUserFlow(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}

	userStore, err := users.NewUserStore(storage, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user store: %s", err)
	}
	userStore.Empty()
	if err != nil {
		t.Fatalf("cannot empty user store")
	}
	s := scim.New(storage, userStore, "token", wireguard.DisableAllClientConfigs, wireguard.ReactivateAllClientConfigs)

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

	// create a user
	payload := scim.PostUserRequest{
		UserName: "john@domain.inv",
		Name: scim.Name{
			GivenName:  "John",
			FamilyName: "Doe",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("cannot marshal payload: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/api/scim/v2/Users?", bytes.NewBuffer(payloadBytes))
	w := httptest.NewRecorder()
	s.PostUsersHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var postUserRequest scim.PostUserRequest
	err = json.NewDecoder(resp.Body).Decode(&postUserRequest)
	if err != nil {
		t.Fatalf("Could not decode output: %s", err)
	}

	if postUserRequest.Id == "" {
		t.Fatalf("id is empty: %s", err)
	}

	user, err := s.UserStore.GetUserByID(postUserRequest.Id)
	if err != nil {
		t.Fatalf("Cannot get newly created user: %s", err)
	}

	// generate VPN config
	_, err = wireguard.CreateNewVPNConfig(storage)
	if err != nil {
		t.Fatalf("Cannot create vpn config: %s", err)
	}

	peerConfig, err := wireguard.NewEmptyClientConfig(storage, user.ID)
	if err != nil {
		t.Fatalf("NewEmptyClientConfig error: %s", err)

	}

	if peerConfig.Disabled {
		t.Fatalf("VPN connection is disabled. Expected not disabled")
	}

	connectionID := fmt.Sprintf("%s-1", user.ID)
	userConfigFilename := storage.ConfigPath(path.Join(wireguard.VPN_CLIENTS_DIR, connectionID+".json"))
	_, err = storage.ReadFile(userConfigFilename)
	if err != nil {
		t.Fatalf("could not read user config file")
	}

	// disable user

	payload.Active = false
	payload.Id = user.ID
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		t.Fatalf("cannot marshal payload: %s", err)
	}

	req = httptest.NewRequest("PUT", "http://example.com/api/scim/v2/Users/"+user.ID, bytes.NewBuffer(payloadBytes))
	req.SetPathValue("id", user.ID)
	w = httptest.NewRecorder()
	s.PutUserHandler(w, req)

	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	configBytes, err := storage.ReadFile(userConfigFilename)
	if err != nil {
		t.Fatalf("could not read user file")
	}

	var config2 wireguard.PeerConfig
	err = json.Unmarshal(configBytes, &config2)
	if err != nil {
		t.Fatalf("could not parse config: %s", err)
	}
	if !config2.Disabled {
		t.Fatalf("VPN connection is enabled. Expected disabled")
	}
}

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
	storage := &memorystorage.MockMemoryStorage{}

	v := New(storage, &users.UserStore{})

	err = v.UserStore.Empty()
	if err != nil {
		t.Fatalf("Cannot create context")
	}

	// create a user
	userToCreate := users.User{
		Login:    "john",
		Role:     "user",
		Password: "xyz",
	}
	user, err := v.UserStore.AddUser(userToCreate)
	if err != nil {
		t.Fatalf("user creation error: %s", err)
	}

	// generate VPN config
	_, err = wireguard.CreateNewVPNConfig(v.Storage)
	if err != nil {
		t.Fatalf("Cannot create vpn config: %s", err)
	}

	req := httptest.NewRequest("POST", "http://example.com/connections", nil)
	w := httptest.NewRecorder()
	v.connectionsHandler(w, req.WithContext(context.WithValue(context.Background(), rest.CustomValue("user"), user)))

	resp := w.Result()

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
	v.connectionsElementHandler(w, req.WithContext(context.WithValue(context.Background(), rest.CustomValue("user"), user)))

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

	err = v.UserStore.DeleteUserByID(user.ID)
	if err != nil {
		t.Fatalf("user deletion error: %s", err)
	}

	_, err = storage.ReadFile(userConfigFilename)
	if err == nil {
		t.Fatalf("could read user config file, expected not to")
	}
}
