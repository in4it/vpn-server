package observability

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			"Date": 1720613813.197045,
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

	err = o.WriteBufferToStorage()
	if err != nil {
		t.Fatalf("write buffer to storage error: %s", err)
	}
	dirlist, err := storage.ReadDir("")
	if err != nil {
		t.Fatalf("read dir error: %s", err)
	}
	for _, filename := range dirlist {
		filenameOut, err := storage.ReadFile(filename)
		if err != nil {
			t.Fatalf("read file error: %s", err)
		}
		fmt.Printf("filenameOut: %s", filenameOut)
	}
}
