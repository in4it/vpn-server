package observability

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecoding(t *testing.T) {
	data := `[{"date":1720613813.197045,"rand_value":"rand"}]`
	messages, err := Decode(bytes.NewBuffer([]byte(data)))
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

func TestDecodeMsg(t *testing.T) {
	msgs := []FluentBitMessage{
		{
			Date: 1720613813.197045,
			Data: map[string]string{
				"mykey":      "this is myvalue",
				"second key": "this is my second value",
				"third key":  "this is my third value",
			},
		},
		/*{
			Date: 1720613813.197099,
			Data: map[string]string{
				"second data set": "my value",
			},
		},*/
	}
	encoded := encodeMessage(msgs)
	decoded := decodeMessage(encoded)
	fmt.Printf("decoded: %+v\n", decoded)

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
