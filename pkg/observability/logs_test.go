package observability

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestGetLogs(t *testing.T) {
	totalMessagesToGenerate := 100
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(storage, 20)
	payload := IncomingData{
		{
			"date": 1720613813.197045,
			"log":  "this is string: ",
		},
	}

	for i := 0; i < totalMessagesToGenerate; i++ {
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

	// wait until all data is flushed
	o.ActiveBufferWriters.Wait()

	// flush remaining data that hasn't been flushed
	if n := o.Buffer.Len(); n >= 0 {
		err := o.WriteBufferToStorage(int64(n))
		if err != nil {
			t.Fatalf("write log buffer to storage error (buffer: %d): %s", o.Buffer.Len(), err)
		}
	}

	now := time.Now()
	maxLogLines := 100
	search := ""

	logEntryResponse, err := o.getLogs(now, now, 0, 0, maxLogLines, search)
	if err != nil {
		t.Fatalf("get logs error: %s", err)
	}
	if len(logEntryResponse.LogEntries) != totalMessagesToGenerate {
		t.Fatalf("didn't get the same log entries as messaged we generated: got: %d, expected: %d", len(logEntryResponse.LogEntries), totalMessagesToGenerate)
	}
	if logEntryResponse.LogEntries[0].Timestamp != floatToDate(1720613813.197045).Format(TIMESTAMP_FORMAT) {
		t.Fatalf("unexpected timestamp")
	}
}

func TestFloatToDate(t *testing.T) {
	now := time.Now()
	floatDate := float64(now.Unix()) + float64(now.Nanosecond())/1e9
	floatToDate := floatToDate(floatDate)
	if !now.Equal(floatToDate) {
		t.Fatalf("times are not equal. Got: %s, expected: %s", floatToDate, now)
	}
}

func TestKeyValue(t *testing.T) {
	logEntryResponse := LogEntryResponse{
		Keys: KeyValueInt{
			{Key: "k", Value: "v", Total: 4},
		},
	}
	out, err := json.Marshal(logEntryResponse)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if !strings.Contains(string(out), `"keys":[{"key":"k","value":"v","total":4}]`) {
		t.Fatalf("wrong output: %s", out)
	}
}
