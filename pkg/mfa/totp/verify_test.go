package totp

import (
	"testing"
	"time"
)

func TestVerify(t *testing.T) { // validated with https://2fa.glitch.me/
	interval := int64(57275699)
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	token, err := GetToken(secret, interval)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if token != "840823" {
		t.Fatalf("wrong token. Got: %s", token)
	}
}

func TestVerifyWrongSecret(t *testing.T) {
	interval := int64(57275699)
	secret := "wrong secret"
	_, err := GetToken(secret, interval)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestVerifyMultipleIntervals(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	ok, err := verifyMultipleIntervals(secret, "312137", 20, time.Unix(1718272397, 0))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if !ok {
		t.Fatalf("no token matched")
	}
}

func TestVerifyMultipleIntervalsWrongToken(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	ok, err := verifyMultipleIntervals(secret, "312137", 20, time.Unix(1718272000, 0))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if ok {
		t.Fatalf("token matched, but shouldn't have")
	}
}
