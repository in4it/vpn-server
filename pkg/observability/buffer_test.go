package observability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"testing"

	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestIngestion(t *testing.T) {
	totalMessagesToGenerate := 1000
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(storage, 20)
	o.Storage = storage
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
		decodedMessages := decodeMessages(messages)
		totalMessages += len(decodedMessages)
	}
	if len(dirlist) == 0 {
		t.Fatalf("expected multiple files in directory, got %d", len(dirlist))
	}

	if totalMessages != totalMessagesToGenerate {
		t.Fatalf("Tried to generate total message count of: %d; got: %d", totalMessagesToGenerate, totalMessages)
	}
}

func TestIngestionMoreMessages(t *testing.T) {
	t.Skip()                            // we can skip this for general unit testing
	totalMessagesToGenerate := 10000000 // 10,000,000
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(MAX_BUFFER_SIZE)
	o.Storage = storage
	payload := IncomingData{
		{
			"date": 1720613813.197045,
			"log":  "this is string: ",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %s", err)
	}

	for i := 0; i < totalMessagesToGenerate; i++ {
		data := io.NopCloser(bytes.NewReader(payloadBytes))
		err := o.Ingest(data)
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
		decodedMessages := decodeMessages(messages)
		totalMessages += len(decodedMessages)
	}
	if len(dirlist) == 0 {
		t.Fatalf("expected multiple files in directory, got %d", len(dirlist))
	}

	if totalMessages != totalMessagesToGenerate {
		t.Fatalf("Tried to generate total message count of: %d; got: %d", totalMessagesToGenerate, totalMessages)
	}
	fmt.Printf("Buffer size (read+unread): %d in %d files\n", o.Buffer.Cap(), len(dirlist))

}

func BenchmarkIngest10000000(b *testing.B) {
	totalMessagesToGenerate := 10000000 // 10,000,000
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(MAX_BUFFER_SIZE)
	o.Storage = storage
	payload := IncomingData{
		{
			"date": 1720613813.197045,
			"log":  "this is string",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		b.Fatalf("marshal error: %s", err)
	}

	for i := 0; i < totalMessagesToGenerate; i++ {
		data := io.NopCloser(bytes.NewReader(payloadBytes))
		err := o.Ingest(data)
		if err != nil {
			b.Fatalf("ingest error: %s", err)
		}
	}

	// wait until all data is flushed
	o.ActiveBufferWriters.Wait()

	// flush remaining data that hasn't been flushed
	if n := o.Buffer.Len(); n >= 0 {
		err := o.WriteBufferToStorage(int64(n))
		if err != nil {
			b.Fatalf("write log buffer to storage error (buffer: %d): %s", o.Buffer.Len(), err)
		}
	}
}

func BenchmarkIngest100000000(b *testing.B) {
	totalMessagesToGenerate := 10000000 // 10,000,000
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(MAX_BUFFER_SIZE)
	o.Storage = storage
	payload := IncomingData{
		{
			"date": 1720613813.197045,
			"log":  "this is string",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		b.Fatalf("marshal error: %s", err)
	}

	for i := 0; i < totalMessagesToGenerate; i++ {
		data := io.NopCloser(bytes.NewReader(payloadBytes))
		err := o.Ingest(data)
		if err != nil {
			b.Fatalf("ingest error: %s", err)
		}
	}

	// wait until all data is flushed
	o.ActiveBufferWriters.Wait()

	// flush remaining data that hasn't been flushed
	if n := o.Buffer.Len(); n >= 0 {
		err := o.WriteBufferToStorage(int64(n))
		if err != nil {
			b.Fatalf("write log buffer to storage error (buffer: %d): %s", o.Buffer.Len(), err)
		}
	}
}
