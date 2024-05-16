package network

import "testing"

func TestParseResolve(t *testing.T) {
	res, err := parseResolve([]byte("# comment\n# comment\n\nsearch abc.invalid\nnameserver 192.168.0.1\n"))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if len(res) == 0 {
		t.Fatalf("res is empty")
	}
	if res[0] != "192.168.0.1" {
		t.Fatalf("wrong ip returned: %s", res[0])
	}
}
