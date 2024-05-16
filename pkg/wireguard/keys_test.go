package wireguard

import (
	"fmt"
	"testing"
)

func TestGenerateKeys(t *testing.T) {
	priv, pub, err := GenerateKeys()
	if err != nil {
		t.Errorf("error: %s", err)
	}
	fmt.Printf("%s\n%s", priv, pub)
}
