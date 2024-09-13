package observability

import (
	"net/http"
)

func New() *Observability {
	o := NewWithoutMonitor(MAX_BUFFER_SIZE)
	go o.monitorBuffer()
	return o
}
func NewWithoutMonitor(maxBufferSize int) *Observability {
	o := &Observability{
		Buffer:        &ConcurrentRWBuffer{},
		MaxBufferSize: maxBufferSize,
	}
	return o
}

type Iface interface {
	GetRouter() *http.ServeMux
}
