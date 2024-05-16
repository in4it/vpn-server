package rest

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"

	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
)

func TestGetJWTKeys(t *testing.T) {
	mockStorage := testingmocks.MockMemoryStorage{}
	keys, err := getJWTKeys(&mockStorage)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	privateKeyFromFile, err := mockStorage.ReadFile(mockStorage.ConfigPath("pki/private.pem"))
	if err != nil {
		t.Fatalf("read error: %s", err)
	}
	_, err = mockStorage.ReadFile(mockStorage.ConfigPath("pki/public.pem"))
	if err != nil {
		t.Fatalf("read error: %s", err)
	}

	var buf bytes.Buffer
	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(keys.PrivateKey),
	}
	err = pem.Encode(&buf, privateKey)
	if err != nil {
		t.Fatalf("pem encode error: %s", err)
	}

	if !bytes.Equal(privateKeyFromFile, buf.Bytes()) {
		t.Fatalf("private keys don't match")
	}
}
