package observability

import (
	"net/http"

	"github.com/in4it/wireguard-server/pkg/storage"
)

func New(defaultStorage storage.Iface) *Observability {
	o := NewWithoutMonitor(defaultStorage, MAX_BUFFER_SIZE)
	go o.monitorBuffer()
	return o
}
func NewWithoutMonitor(storage storage.Iface, maxBufferSize int) *Observability {
	o := &Observability{
		Buffer:        &ConcurrentRWBuffer{},
		MaxBufferSize: maxBufferSize,
		Storage:       storage,
	}
	return o
}

type Iface interface {
	GetRouter() *http.ServeMux
}
