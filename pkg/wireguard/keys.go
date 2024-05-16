package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
)

const keyLength = 32

func GenerateKeys() (string, string, error) {
	key := make([]byte, keyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", "", fmt.Errorf("GenerateKeys seed error: %s", err)
	}

	key[0] &= 248
	key[31] &= 127
	key[31] |= 64

	privKey := [keyLength]byte(key)
	pubKey := [keyLength]byte{}

	curve25519.ScalarBaseMult(&pubKey, &privKey)

	return base64.StdEncoding.EncodeToString(privKey[:]), base64.StdEncoding.EncodeToString(pubKey[:]), nil

}

func GeneratePresharedKey() (string, error) {
	key := make([]byte, keyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("GenerateKeys seed error: %s", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
