package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestIngestionHandler(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(20)
	o.Storage = storage
	payload := IncomingData{
		{
			"date": 1720613813.197045,
			"log":  "this is a string",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/observability/ingestion/json", bytes.NewReader(payloadBytes))
	w := httptest.NewRecorder()
	o.ingestionHandler(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code OK. Got: %d", res.StatusCode)
	}

	// wait until all data is flushed
	o.ActiveBufferWriters.Wait()

	// flush remaining data that hasn't been flushed
	if n := o.Buffer.Len(); n >= 0 {
		err := o.WriteBufferToStorage(int64(n))
		if err != nil {
			t.Fatalf("write log buffer to storage error (buffer: %d): %s", o.Buffer.Len(), err)
		}
	}

	dirlist, err := storage.ReadDir("")
	if err != nil {
		t.Fatalf("read dir error: %s", err)
	}
	if len(dirlist) == 0 {
		t.Fatalf("dir is empty")
	}
	messages, err := storage.ReadFile(dirlist[0])
	if err != nil {
		t.Fatalf("read file error: %s", err)
	}
	decodedMessages := decodeMessages(messages)
	if decodedMessages[0].Date != 1720613813.197045 {
		t.Fatalf("unexpected date. Got %f, expected: %f", decodedMessages[0].Date, 1720613813.197045)
	}
	if decodedMessages[0].Data["log"] != "this is a string" {
		t.Fatalf("unexpected log data")
	}
}
