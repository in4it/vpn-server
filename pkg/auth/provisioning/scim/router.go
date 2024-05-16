package scim

import (
	"net/http"
)

func (s *scim) GetRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/api/scim/", s.authMiddleware(http.HandlerFunc(notFoundHandler)))
	mux.Handle("/api/scim/v2/Users", s.authMiddleware(http.HandlerFunc(s.usersHandler)))
	mux.Handle("/api/scim/v2/Users/{id}", s.authMiddleware(http.HandlerFunc(s.userHandler)))

	return mux
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error": "page not found"}`))
}
