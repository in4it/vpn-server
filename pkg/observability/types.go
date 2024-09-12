package observability

import (
	"bytes"

	"github.com/in4it/wireguard-server/pkg/storage"
)

type IncomingData []map[string]any

type FluentBitMessage struct {
	Date float64           `json:"date"`
	Data map[string]string `json:"data"`
}

type Observability struct {
	Storage storage.Iface
	Buffer  bytes.Buffer
}
