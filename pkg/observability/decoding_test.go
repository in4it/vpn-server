package observability

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestDecoding(t *testing.T) {
	payload := IncomingData{
		{
			"date":       1720613813.197045,
			"rand_value": "rand",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %s", err)
	}
	messages, err := Decode(bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if len(messages) == 0 {
		t.Fatalf("no messages returned")
	}
	if messages[0].Date != 1720613813.197045 {
		t.Fatalf("wrong date returned")
	}
	val, ok := messages[0].Data["rand_value"]
	if !ok {
		t.Fatalf("rand_value key not found")
	}
	if string(val) != "rand" {
		t.Fatalf("wrong data returned: %s", val)
	}
}

func TestDecodingMultiMessage(t *testing.T) {
	payload := IncomingData{
		{
			"date":           1727119152.0,
			"container_name": "/fluentbit-nginx-1",
			"source":         "stdout",
			"log":            "/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration",
			"container_id":   "7a9c8ae0ca6c5434b778fa0a2e74e038710b3f18dedb3478235291832f121186",
		},
		{
			"date":           1727119152.0,
			"source":         "stdout",
			"log":            "/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/",
			"container_id":   "7a9c8ae0ca6c5434b778fa0a2e74e038710b3f18dedb3478235291832f121186",
			"container_name": "/fluentbit-nginx-1",
		},
		{
			"date":           1727119152.0,
			"container_id":   "7a9c8ae0ca6c5434b778fa0a2e74e038710b3f18dedb3478235291832f121186",
			"container_name": "/fluentbit-nginx-1",
			"source":         "stdout",
			"log":            "/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %s", err)
	}
	messages, err := Decode(bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if len(messages) != len(payload) {
		t.Fatalf("incorrect messages returned. Got %d, expected %d", len(messages), len(payload))
	}
	val, ok := messages[2].Data["container_id"]
	if !ok {
		t.Fatalf("container_id key not found")
	}
	if string(val) != "7a9c8ae0ca6c5434b778fa0a2e74e038710b3f18dedb3478235291832f121186" {
		t.Fatalf("wrong data returned: %s", val)
	}
}

func TestDecodeMessages(t *testing.T) {
	msgs := []FluentBitMessage{
		{
			Date: 1720613813.197045,
			Data: map[string]string{
				"mykey":      "this is myvalue",
				"second key": "this is my second value",
				"third key":  "this is my third value",
			},
		},
		{
			Date: 1720613813.197099,
			Data: map[string]string{
				"second data set": "my value",
			},
		},
	}
	encoded := encodeMessage(msgs)
	decoded := decodeMessages(encoded)

	if len(msgs) != len(decoded) {
		t.Fatalf("length doesn't match")
	}
	for k := range decoded {
		if msgs[k].Date != decoded[k].Date {
			t.Fatalf("date doesn't match")
		}
		if len(msgs[k].Data) != len(decoded[k].Data) {
			t.Fatalf("length of data doesn't match")
		}
		for kk := range decoded[k].Data {
			if msgs[k].Data[kk] != decoded[k].Data[kk] {
				t.Fatalf("key/value pair in data doesn't match: key: %s. Data: %s vs %s", kk, msgs[k].Data[kk], decoded[k].Data[kk])
			}
		}
	}
}

func TestDecodeMessage(t *testing.T) {
	msgs := []FluentBitMessage{
		{
			Date: 1720613813.197099,
			Data: map[string]string{
				"second data set": "my value",
			},
		},
	}
	encoded := encodeMessage(msgs)
	message := decodeMessage(encoded)

	if message.Date != message.Date {
		t.Fatalf("date doesn't match")
	}
	if len(msgs[0].Data) != len(message.Data) {
		t.Fatalf("length of data doesn't match")
	}
	for kk := range message.Data {
		if msgs[0].Data[kk] != message.Data[kk] {
			t.Fatalf("key/value pair in data doesn't match: key: %s. Data: %s vs %s", kk, message.Data[kk], message.Data[kk])
		}
	}
}
func TestDecodeMessageWithoutTerminator(t *testing.T) {
	msgs := []FluentBitMessage{
		{
			Date: 1720613813.197099,
			Data: map[string]string{
				"second data set": "my value",
			},
		},
	}
	encoded := encodeMessage(msgs)
	message := decodeMessage(bytes.TrimSuffix(encoded, []byte{0xff, 0xff}))

	if message.Date != message.Date {
		t.Fatalf("date doesn't match")
	}
	if len(msgs[0].Data) != len(message.Data) {
		t.Fatalf("length of data doesn't match: got: '%s', expected '%s'", message.Data, msgs[0].Data)
	}
	for kk := range message.Data {
		if msgs[0].Data[kk] != message.Data[kk] {
			t.Fatalf("key/value pair in data doesn't match: key: %s. Data: %s vs %s", kk, message.Data[kk], msgs[0].Data[kk])
		}
	}
}

func TestScanMessage(t *testing.T) {
	msgs := []FluentBitMessage{
		{
			Date: 1720613813.197045,
			Data: map[string]string{
				"mykey":      "this is myvalue",
				"second key": "this is my second value",
				"third key":  "this is my third value",
			},
		},
		{
			Date: 1720613813.197099,
			Data: map[string]string{
				"second data set": "my value",
			},
		},
	}
	encoded := encodeMessage(msgs)
	// first record
	advance, record1, err := scanMessage(encoded, false)
	if err != nil {
		t.Fatalf("scan lines error: %s", err)
	}
	firstRecord := decodeMessages(append(record1, []byte{0xff, 0xff}...))
	if len(firstRecord) == 0 {
		t.Fatalf("first record is empty")
	}
	if firstRecord[0].Data["mykey"] != "this is myvalue" {
		t.Fatalf("wrong data returned")
	}
	// second record
	advance2, record2, err := scanMessage(encoded[advance:], false)
	if err != nil {
		t.Fatalf("scan lines error: %s", err)
	}
	secondRecord := decodeMessages(append(record2, []byte{0xff, 0xff}...))
	if len(secondRecord) == 0 {
		t.Fatalf("first record is empty")
	}
	if secondRecord[0].Data["second data set"] != "my value" {
		t.Fatalf("wrong data returned in second record")
	}
	// third call
	advance3, record3, err := scanMessage(encoded[advance+advance2:], false)
	if err != nil {
		t.Fatalf("scan lines error: %s", err)
	}
	if advance3 != 0 {
		t.Fatalf("third record should be empty. Got: %+v", record3)
	}
}
