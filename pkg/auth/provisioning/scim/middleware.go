package scim

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *scim) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			writeWithStatus(w, []byte(`{"error": "token not found"}`), http.StatusUnauthorized)
			return
		}
		tokenString := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1)
		if len(tokenString) == 0 {
			returnError(w, fmt.Errorf("empty token"), http.StatusUnauthorized)
			return
		}
		if s.Token == "" {
			writeWithStatus(w, []byte(`{"error": "scim not active"}`), http.StatusUnauthorized)
			return
		}
		if s.Token != tokenString {
			writeWithStatus(w, []byte(`{"error": "authentication failed"}`), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
