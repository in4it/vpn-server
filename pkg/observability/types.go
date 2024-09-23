package observability

import (
	"bytes"
	"strconv"
	"strings"
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
	Enabled      bool        `json:"enabled"`
	LogEntries   []LogEntry  `json:"logEntries"`
	Environments []string    `json:"environments"`
	Keys         KeyValueInt `json:"keys"`
	NextPos      int64       `json:"nextPos"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Data      string `json:"data"`
}

type KeyValueInt map[KeyValue]int

type KeyValue struct {
	Key   string
	Value string
}

func (kv KeyValueInt) MarshalJSON() ([]byte, error) {
	res := "["
	for k, v := range kv {
		res += `{ "key" : "` + k.Key + `", "value": "` + k.Value + `", "total": ` + strconv.Itoa(v) + ` },`
	}
	res = strings.TrimRight(res, ",")
	res += "]"
	return []byte(res), nil
}
