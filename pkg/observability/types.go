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
	Buffer                *ConcurrentRWBuffer
	LastFlushed           time.Time
	FlushOverflow         atomic.Bool
	FlushOverflowSequence atomic.Uint64
	ActiveBufferWriters   sync.WaitGroup
	MaxBufferSize         int
}

type ConcurrentRWBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

type LogEntryResponse struct {
	LogEntries []LogEntry `json:"logEntries"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}
