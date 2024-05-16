package observability

import "net/http"

func New() *Observability {
	return &Observability{}
}

type Observability struct {
}

type Iface interface {
	GetRouter() *http.ServeMux
}
