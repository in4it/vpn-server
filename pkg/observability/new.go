package observability

import "net/http"

func New() *Observability {
	return &Observability{}
}

type Iface interface {
	GetRouter() *http.ServeMux
}
