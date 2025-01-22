package vpn

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/in4it/go-devops-platform/rest"
)

func TestIsAdmin(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "<html><body>Hello World!</body></html>")
		if err != nil {
			t.Fatalf("write error: %s", err)
		}
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	rest.IsAdminMiddleware(http.HandlerFunc(handler)).ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != 403 {
		t.Fatalf("expected permission denied")
	}

}
