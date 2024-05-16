package observability

import (
	"bytes"
	"testing"
)

func TestDecoding(t *testing.T) {
	data := `[{"date":1720613813.197045,"rand_value":5523152494216581654}]`
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
	if val.(float64) != 5523152494216581654 {
		t.Fatalf("wrong data returned")
	}
}
