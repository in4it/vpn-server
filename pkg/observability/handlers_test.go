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
	o := &Observability{
		Storage: storage,
	}
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

	err = o.WriteBufferToStorage(int64(o.Buffer.Len()))
	if err != nil {
		t.Fatalf("write buffer to storage error: %s", err)
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
	decodedMessages := decodeMessage(messages)
	if decodedMessages[0].Date != 1720613813.197045 {
		t.Fatalf("unexpected date")
	}
	if decodedMessages[0].Data["log"] != "this is a string" {
		t.Fatalf("unexpected log data")
	}
}
