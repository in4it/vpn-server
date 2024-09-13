package observability

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"

	"github.com/in4it/wireguard-server/pkg/storage"
)

type IncomingData []map[string]any

type FluentBitMessage struct {
	Date float64           `json:"date"`
	Data map[string]string `json:"data"`
}

type Observability struct {
	Storage               storage.Iface
	Buffer                bytes.Buffer
	LastFlushed           time.Time
	BufferMu              sync.Mutex
	FlushOverflow         atomic.Bool
	FlushOverflowSequence atomic.Uint64
}
