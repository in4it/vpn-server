package rest

import "testing"

func TestIsAlphaNumeric(t *testing.T) {
	if !isAlphaNumeric("abc123") {
		t.Errorf("expected alphanumeric")
	}
	if isAlphaNumeric("abc@") {
		t.Errorf("expected alphanumeric")
	}
}
