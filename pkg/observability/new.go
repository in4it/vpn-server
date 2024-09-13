package observability

import "net/http"

func New() *Observability {
	o := &Observability{}
	go o.monitorBuffer()
	return o
}
func NewWithoutMonitor() *Observability {
	o := &Observability{}
	return o
}

type Iface interface {
	GetRouter() *http.ServeMux
}
