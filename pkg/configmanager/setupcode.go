package configmanager

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os/user"

	"github.com/in4it/go-devops-platform/storage"
)

func writeSetupCode(storage storage.Iface) error {
	randomString, err := getRandomString(128)
	if err != nil {
		return fmt.Errorf("GetRandomString error: %s", randomString)
	}

	err = storage.WriteFile("setup-code.txt", []byte(randomString+"\n"))
	if err != nil {
		return fmt.Errorf("vpn config write error: %s", err)
	}
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %s", err)
	}
	if currentUser.Username != "vpn" {
		err = storage.EnsureOwnership("setup-code.txt", "vpn")
		if err != nil {
			return fmt.Errorf("config write error: %s", err)
		}
	}
	return nil
}

func getRandomString(n int) (string, error) {
	buf := make([]byte, n)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("crypto/rand Reader error: %s", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
