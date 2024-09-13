package observability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestIngestion(t *testing.T) {
	t.Skip() // working on this test
	storage := &memorystorage.MockMemoryStorage{}
	o := &Observability{
		Storage: storage,
	}
	payload := IncomingData{
		{
			"date": 1720613813.197045,
			"log":  "this is string: ",
		},
	}

	for i := 0; i < MAX_BUFFER_SIZE; i++ {
		payload[0]["log"] = "this is string: " + strconv.Itoa(i)
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal error: %s", err)
		}
		data := io.NopCloser(bytes.NewReader(payloadBytes))
		err = o.Ingest(data)
		if err != nil {
			t.Fatalf("ingest error: %s", err)
		}
	}

	// flush remaining data
	time.Sleep(1 * time.Second)
	if o.Buffer.Len() >= MAX_BUFFER_SIZE {
		if o.FlushOverflow.CompareAndSwap(false, true) {
			if n := o.Buffer.Len(); n >= MAX_BUFFER_SIZE {
				err := o.WriteBufferToStorage(int64(n))
				if err != nil {
					t.Fatalf("write log buffer to storage error (buffer: %d): %s", o.Buffer.Len(), err)
				}
			}
			o.FlushOverflow.Swap(false)
		}
	}

	dirlist, err := storage.ReadDir("")
	if err != nil {
		t.Fatalf("read dir error: %s", err)
	}

	totalMessages := 0
	for _, file := range dirlist {
		messages, err := storage.ReadFile(file)
		if err != nil {
			t.Fatalf("read file error: %s", err)
		}
		decodedMessages := decodeMessage(messages)
		for _, message := range decodedMessages {
			fmt.Printf("decoded message: %s\n", message.Data["log"])
		}
		totalMessages += len(decodedMessages)
	}
	fmt.Printf("totalmessages: %d", totalMessages)
	if len(dirlist) != 3 {
		t.Fatalf("expected 3 files in directory, got %d", len(dirlist))
	}
}
