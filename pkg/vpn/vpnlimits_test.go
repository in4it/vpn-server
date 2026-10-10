package vpn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/in4it/go-devops-platform/rest"
	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/go-devops-platform/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

// TestConnectionsHandlerMaxConnectionsPerUser: a user can't create unlimited connections (address pool exhaustion)
func TestConnectionsHandlerMaxConnectionsPerUser(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	if _, err := wireguard.CreateNewVPNConfig(storage); err != nil {
		t.Fatalf("Cannot create vpn config: %s", err)
	}
	v := New(storage, &users.UserStore{})
	user := users.User{ID: "3df97301-5f73-407a-a26b-91829f1e7f48", Login: "john", Role: "admin"}
	otherUser := users.User{ID: "5f2b3a8e-1111-4c4c-9d9d-123456789abc", Login: "jane", Role: "user"}

	createConnection := func(u users.User) int {
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/vpn/connections", nil)
		w := httptest.NewRecorder()
		v.connectionsHandler(w, req.WithContext(context.WithValue(context.Background(), rest.CustomValue("user"), u)))
		return w.Result().StatusCode
	}

	for i := 0; i < MAX_CONNECTIONS_PER_USER; i++ {
		if status := createConnection(user); status != http.StatusOK {
			t.Fatalf("connection %d: expected 200, got %d", i+1, status)
		}
	}
	if status := createConnection(user); status != http.StatusForbidden {
		t.Fatalf("expected 403 after %d connections, got %d", MAX_CONNECTIONS_PER_USER, status)
	}
	configNumbers, err := wireguard.GetConfigNumbers(storage, user.ID)
	if err != nil {
		t.Fatalf("GetConfigNumbers error: %s", err)
	}
	if len(configNumbers) != MAX_CONNECTIONS_PER_USER {
		t.Fatalf("expected %d connections, got %d", MAX_CONNECTIONS_PER_USER, len(configNumbers))
	}
	// limit is per user
	if status := createConnection(otherUser); status != http.StatusOK {
		t.Fatalf("other user: expected 200, got %d", status)
	}
}

func TestConnectionsElementHandlerInvalidID(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	v := New(storage, &users.UserStore{})
	user := users.User{ID: "3df97301-5f73-407a-a26b-91829f1e7f48", Login: "john"}

	invalidIDs := []string{
		user.ID,               // no connection number
		user.ID + "-",         // empty connection number
		user.ID + "1",         // no separator
		user.ID + "-1\\..\\x", // backslash
		user.ID + "-1\\x",     // backslash
		user.ID + "-1/../x",   // slash
		user.ID + "-1.json",   // dot
		user.ID + "-abc",      // not a number
		user.ID + "-1-2",      // extra parts
		"other-user-1",        // other user
		user.ID + "-1 ",       // trailing space
		"-1",                  // empty user
	}
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		for _, id := range invalidIDs {
			req := httptest.NewRequest(method, "http://example.com/api/vpn/connection/x", nil)
			req.SetPathValue("id", id)
			w := httptest.NewRecorder()
			v.connectionsElementHandler(w, req.WithContext(context.WithValue(context.Background(), rest.CustomValue("user"), user)))
			if w.Result().StatusCode != http.StatusBadRequest {
				t.Fatalf("%s %q: expected 400, got %d", method, id, w.Result().StatusCode)
			}
		}
	}
	if err := validateConnectionID(user.ID+"-12", user.ID); err != nil {
		t.Fatalf("expected valid connection id: %s", err)
	}
}
