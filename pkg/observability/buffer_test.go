package observability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"testing"

	"github.com/in4it/wireguard-server/pkg/logging"
	memorystorage "github.com/in4it/wireguard-server/pkg/storage/memory"
)

func TestIngestion(t *testing.T) {
	logging.Loglevel = logging.LOG_DEBUG
	totalMessagesToGenerate := 20
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(storage, 20)
	o.Storage = storage
	payloads := IncomingData{}
	for i := 0; i < totalMessagesToGenerate/10; i++ {
		payloads = append(payloads, map[string]any{
			"date": 1720613813.197045,
			"log":  "this is string: " + strconv.Itoa(i),
		})
	}

	for i := 0; i < totalMessagesToGenerate/len(payloads); i++ {
		payloadBytes, err := json.Marshal(payloads)
		if err != nil {
			t.Fatalf("marshal error: %s", err)
		}
		data := io.NopCloser(bytes.NewReader(payloadBytes))
		err = o.Ingest(data)
		if err != nil {
			t.Fatalf("ingest error: %s", err)
		}
	}

	err := o.Flush()
	if err != nil {
		t.Fatalf("flush error: %s", err)
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
	o := NewWithoutMonitor(storage, MAX_BUFFER_SIZE)
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

	err = o.Flush()
	if err != nil {
		t.Fatalf("flush error: %s", err)
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
	o := NewWithoutMonitor(storage, MAX_BUFFER_SIZE)
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
	o := NewWithoutMonitor(storage, MAX_BUFFER_SIZE)
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

func TestEnsurePath(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	err := ensurePath(storage, "a/b/c/filename.txt")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
}

func TestMergeBufferPosAndPrefix(t *testing.T) {
	testCase1 := []BufferPosAndPrefix{
		{
			prefix: "abc",
			offset: 3,
		},
		{
			prefix: "abc",
			offset: 9,
		},
		{
			prefix: "abc",
			offset: 2,
		},
		{
			prefix: "abc2",
			offset: 3,
		},
		{
			prefix: "abc2",
			offset: 2,
		},
		{
			prefix: "abc3",
			offset: 2,
		},
	}
	expected1 := []BufferPosAndPrefix{
		{
			prefix: "abc",
			offset: 14,
		},
		{
			prefix: "abc2",
			offset: 5,
		},
		{
			prefix: "abc3",
			offset: 2,
		},
	}
	res := mergeBufferPosAndPrefix(testCase1)
	if !slices.Equal(res, expected1) {
		t.Fatalf("test case is not equal to expected\nGot: %+v\nExpected:%+v\n", res, expected1)
	}
}

func TestReadPrefix(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	o := NewWithoutMonitor(storage, MAX_BUFFER_SIZE)
	o.Buffer.prefix = []BufferPosAndPrefix{
		{
			prefix: "abc",
			offset: 3,
		},
		{
			prefix: "abc",
			offset: 9,
		},
		{
			prefix: "abc",
			offset: 2,
		},
		{
			prefix: "abc2",
			offset: 3,
		},
		{
			prefix: "abc2",
			offset: 2,
		},
		{
			prefix: "abc3",
			offset: 2,
		},
	}
	expected1 := []BufferPosAndPrefix{
		{
			prefix: "abc",
			offset: 3,
		},
		{
			prefix: "abc",
			offset: 9,
		},
		{
			prefix: "abc",
			offset: 2,
		},
	}
	expected2 := []BufferPosAndPrefix{
		{
			prefix: "abc2",
			offset: 3,
		},
	}
	res := o.Buffer.ReadPrefix(int64(o.Buffer.prefix[0].offset + o.Buffer.prefix[1].offset + o.Buffer.prefix[2].offset))
	if !slices.Equal(res, expected1) {
		t.Fatalf("test case is not equal to expected\nGot: %+v\nExpected:%+v\n", res, expected1)
	}
	res2 := o.Buffer.ReadPrefix(3)
	if !slices.Equal(res2, expected2) {
		t.Fatalf("test case is not equal to expected\nGot: %+v\nExpected:%+v\n", res, expected2)
	}
}
