package saml

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func getRandomString(n int) (string, error) {
	buf := make([]byte, n)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("crypto/rand Reader error: %s", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
