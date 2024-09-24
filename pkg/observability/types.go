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

type KeyValueInt []KeyValueTotal

type KeyValueTotal struct {
	Key   string
	Value string
	Total int
}
type KeyValue struct {
	Key   string
	Value string
}

func (kv KeyValueInt) MarshalJSON() ([]byte, error) {
	res := "["
	for _, v := range kv {
		res += `{ "key" : "` + v.Key + `", "value": "` + v.Value + `", "total": ` + strconv.Itoa(v.Total) + ` },`
	}
	res = strings.TrimRight(res, ",")
	res += "]"
	return []byte(res), nil
}

func (kv KeyValueInt) Len() int {
	return len(kv)
}
func (kv KeyValueInt) Less(i, j int) bool {
	if kv[i].Key == kv[j].Key {
		return kv[i].Value < kv[j].Value
	}
	return kv[i].Key < kv[j].Key
}
func (kv KeyValueInt) Swap(i, j int) {
	kv[i], kv[j] = kv[j], kv[i]
}
