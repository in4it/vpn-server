package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/in4it/wireguard-server/pkg/license"
	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
	"github.com/in4it/wireguard-server/pkg/users"
)

func TestContextHandlerSetupSecret(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}

	storage.WriteFile(SETUP_CODE_FILE, []byte(`secret setup code`))

	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("new user store error")
	}
	c, err := getEmptyContext("appdir")
	if err != nil {
		t.Fatalf("cannot create empty context")
	}
	c.Storage = &Storage{Client: storage}
	c.UserStore = userStore

	payload := ContextRequest{
		Secret:        "secret setup code",
		AdminPassword: "adminPassword",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/setup", bytes.NewBuffer(payloadBytes))
	w := httptest.NewRecorder()
	c.contextHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got read error: %s", err)
	}

	var contextSetupResponse ContextSetupResponse
	err = json.Unmarshal(body, &contextSetupResponse)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if !contextSetupResponse.SetupCompleted {
		t.Fatalf("expected setup to be completed")
	}
}

func TestContextHandlerSetupWrongSecret(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}

	storage.WriteFile(SETUP_CODE_FILE, []byte(`secret setup code`))

	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("new user store error")
	}
	c, err := getEmptyContext("appdir")
	if err != nil {
		t.Fatalf("cannot create empty context")
	}
	c.Storage = &Storage{Client: storage}
	c.UserStore = userStore

	payload := ContextRequest{
		AdminPassword: "adminPassword",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/setup", bytes.NewBuffer(payloadBytes))
	w := httptest.NewRecorder()
	c.contextHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 401 {
		t.Fatalf("status code is not 401: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got read error: %s", err)
	}

	var contextSetupResponse ContextSetupResponse
	err = json.Unmarshal(body, &contextSetupResponse)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if contextSetupResponse.SetupCompleted {
		t.Fatalf("expected setup to not be completed")
	}
}
func TestContextHandlerSetupWrongSecretPartial(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}

	storage.WriteFile(SETUP_CODE_FILE, []byte(`secret setup code`))

	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("new user store error")
	}
	c, err := getEmptyContext("appdir")
	if err != nil {
		t.Fatalf("cannot create empty context")
	}
	c.Storage = &Storage{Client: storage}
	c.UserStore = userStore

	payload := ContextRequest{
		Secret:        "secret setup cod",
		AdminPassword: "adminPassword",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/setup", bytes.NewBuffer(payloadBytes))
	w := httptest.NewRecorder()
	c.contextHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 401 {
		t.Fatalf("status code is not 401: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got read error: %s", err)
	}

	var contextSetupResponse ContextSetupResponse
	err = json.Unmarshal(body, &contextSetupResponse)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if contextSetupResponse.SetupCompleted {
		t.Fatalf("expected setup to not be completed")
	}
}

func TestContextHandlerSetupAWSInstanceID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/latest/api/token" {
			w.Write([]byte("this is a test token"))
			return
		}
		if r.RequestURI == "/latest/meta-data/instance-id" {
			w.Write([]byte("i-012aaaaaaaaaaaaa1"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()
	license.MetadataIP = strings.TrimPrefix(ts.URL, "http://")

	storage := &memorystorage.MockMemoryStorage{}

	storage.WriteFile(SETUP_CODE_FILE, []byte(`secret setup code`))

	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("new user store error")
	}
	c, err := getEmptyContext("appdir")
	if err != nil {
		t.Fatalf("cannot create empty context")
	}
	c.Storage = &Storage{Client: storage}
	c.UserStore = userStore
	c.CloudType = "aws"

	payload := ContextRequest{
		InstanceID:    "i-012aaaaaaaaaaaaa1",
		AdminPassword: "adminPassword",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/setup", bytes.NewBuffer(payloadBytes))
	w := httptest.NewRecorder()
	c.contextHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got read error: %s", err)
	}

	var contextSetupResponse ContextSetupResponse
	err = json.Unmarshal(body, &contextSetupResponse)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if !contextSetupResponse.SetupCompleted {
		t.Fatalf("expected setup to be completed")
	}
}
func TestContextHandlerSetupDigitalOceanTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/metadata/v1/tags" {
			w.Write([]byte("vpnsecret-this-is-a-secret-tag"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()
	license.MetadataIP = strings.TrimPrefix(ts.URL, "http://")

	storage := &memorystorage.MockMemoryStorage{}

	storage.WriteFile(SETUP_CODE_FILE, []byte(`secret setup code`))

	userStore, err := users.NewUserStore(storage, -1)
	if err != nil {
		t.Fatalf("new user store error")
	}
	c, err := getEmptyContext("appdir")
	if err != nil {
		t.Fatalf("cannot create empty context")
	}
	c.Storage = &Storage{Client: storage}
	c.UserStore = userStore
	c.CloudType = "digitalocean"

	payload := ContextRequest{
		TagHash:       "vpnsecret-this-is-a-secret-tag",
		AdminPassword: "adminPassword",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	req := httptest.NewRequest("POST", "http://example.com/setup", bytes.NewBuffer(payloadBytes))
	w := httptest.NewRecorder()
	c.contextHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("got read error: %s", err)
	}

	var contextSetupResponse ContextSetupResponse
	err = json.Unmarshal(body, &contextSetupResponse)
	if err != nil {
		t.Fatalf("unmarshal error: %s", err)
	}
	if !contextSetupResponse.SetupCompleted {
		t.Fatalf("expected setup to be completed")
	}
}
