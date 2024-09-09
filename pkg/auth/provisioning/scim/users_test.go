package scim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
	"github.com/in4it/wireguard-server/pkg/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

const USERSTORE_MAX_USERS = 1000

func TestUsersGetCount100EmptyResult(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}

	userStore, err := users.NewUserStore(storage, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user store")
	}
	userStore.Empty()
	if err != nil {
		t.Fatalf("cannot empty user store")
	}

	s := New(storage, userStore, "token")
	req := httptest.NewRequest("GET", "http://example.com/api/scim/v2/Users?count=100&startIndex=1&", nil)
	w := httptest.NewRecorder()
	s.getUsersHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	response, err := listUserResponse([]users.User{}, "", 100, 1)
	if err != nil {
		t.Fatalf("userResponse error: %s", err)
	}
	if string(body) != string(response) {
		t.Fatalf("expected empty input. Got %s\n", string(body))
	}
}

func TestUsersGetCount10(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	userStore, err := users.NewUserStore(storage, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user store")
	}
	err = userStore.Empty()
	if err != nil {
		t.Fatalf("cannot empty user store")
	}
	totalUserCount := 150
	usersToCreate := make([]users.User, totalUserCount)
	for i := 0; i < totalUserCount; i++ {
		usersToCreate[i] = users.User{
			Login: fmt.Sprintf("user-%d@domain.inv", i),
		}
	}
	users, err := userStore.AddUsers(usersToCreate)
	if err != nil {
		t.Fatalf("cannot create users: %s", err)
	}
	s := New(storage, userStore, "token")
	req := httptest.NewRequest("GET", "http://example.com/api/scim/v2/Users?count=10&startIndex=1&", nil)
	w := httptest.NewRecorder()
	s.getUsersHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	response, err := listUserResponse(users, "", 10, 1)
	if err != nil {
		t.Fatalf("userResponse error: %s", err)
	}
	if string(body) != string(response) {
		t.Fatalf("Unexpected output: Got: %s\nExpected: %s\n\n", string(body), string(response))
	}
}

func TestUsersGetCount10Start5(t *testing.T) {
	count := 10
	start := 5
	storage := &memorystorage.MockMemoryStorage{}
	userStore, err := users.NewUserStore(storage, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user store")
	}
	err = userStore.Empty()
	if err != nil {
		t.Fatalf("cannot empty user store")
	}
	totalUserCount := 150
	usersToCreate := make([]users.User, totalUserCount)
	for i := 0; i < totalUserCount; i++ {
		usersToCreate[i] = users.User{
			Login: fmt.Sprintf("user-%d@domain.inv", i),
		}
	}
	users, err := userStore.AddUsers(usersToCreate)
	if err != nil {
		t.Fatalf("cannot create users: %s", err)
	}
	s := New(storage, userStore, "token")
	req := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/api/scim/v2/Users?count=%d&startIndex=%d&", count, start), nil)
	w := httptest.NewRecorder()
	s.getUsersHandler(w, req)

	resp := w.Result()

	var userResponse UserResponse
	err = json.NewDecoder(resp.Body).Decode(&userResponse)
	if err != nil {
		t.Fatalf("Could not decode output: %s", err)
	}
	if userResponse.TotalResults != totalUserCount-start {
		t.Fatalf("Wrong user count: %d", userResponse.TotalResults)
	}
	if userResponse.ItemsPerPage != count {
		t.Fatalf("Wrong page count: %d", userResponse.TotalResults)
	}
	if userResponse.StartIndex != start {
		t.Fatalf("Wrong user start: %d", userResponse.StartIndex)
	}
	if len(userResponse.Resources) != count {
		t.Fatalf("Wrong response count: %d", len(userResponse.Resources))
	}
	if userResponse.Resources[0].UserName != users[5].Login {
		t.Fatalf("Wrong first login: %s (actual) vs %s (expected)", userResponse.Resources[0].UserName, users[5].Login)
	}
}

func TestUsersGetNonExistentUser(t *testing.T) {
	userStore, err := users.NewUserStore(&memorystorage.MockMemoryStorage{}, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user stoer")
	}

	s := New(&memorystorage.MockMemoryStorage{}, userStore, "token")
	req := httptest.NewRequest("GET", "http://example.com/api/scim/v2/Users?filter=userName+eq+%22ward%40in4it.io%22&", nil)
	w := httptest.NewRecorder()
	s.getUsersHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	response, err := listUserResponse([]users.User{}, "", -1, -1)
	if err != nil {
		t.Fatalf("userResponse error: %s", err)
	}
	if string(body) != string(response) {
		t.Fatalf("expected empty input. Got %s\n", string(body))
	}
}

func TestAddUser(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	userStore, err := users.NewUserStore(storage, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user store: %s", err)
	}
	userStore.Empty()
	if err != nil {
		t.Fatalf("cannot empty user store")
	}
	s := New(storage, userStore, "token")
	payload := PostUserRequest{
		UserName: "john@domain.inv",
		Name: Name{
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
	s.postUsersHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Fatalf("User not added. StatusCode: %d", resp.StatusCode)
	}

	var postUserRequest PostUserRequest
	err = json.NewDecoder(resp.Body).Decode(&postUserRequest)
	if err != nil {
		t.Fatalf("Could not decode output: %s", err)
	}

	if postUserRequest.Id == "" {
		t.Fatalf("id is empty: %s", err)
	}
	if postUserRequest.UserName != payload.UserName {
		t.Fatalf("username mismatch: %s (actual) vs %s (expected)", postUserRequest.UserName, payload.UserName)
	}

}

func TestCreateUserConnectionDeleteUserFlow(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	userStore, err := users.NewUserStore(storage, USERSTORE_MAX_USERS)
	if err != nil {
		t.Fatalf("cannot create new user store: %s", err)
	}
	userStore.Empty()
	if err != nil {
		t.Fatalf("cannot empty user store")
	}
	s := New(storage, userStore, "token")

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
	payload := PostUserRequest{
		UserName: "john@domain.inv",
		Name: Name{
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
	s.postUsersHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var postUserRequest PostUserRequest
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
	s.deleteUserHandler(w, req)

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
	s := New(storage, userStore, "token")

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
	payload := PostUserRequest{
		UserName: "john@domain.inv",
		Name: Name{
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
	s.postUsersHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var postUserRequest PostUserRequest
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
	s.putUserHandler(w, req)

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
