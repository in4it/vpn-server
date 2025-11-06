//go:build linux

package network

import (
	"fmt"
	"os"
)

func GetNameservers() ([]string, error) {
	resolveData, err := os.ReadFile("/run/systemd/resolve/resolv.conf")
	if err != nil {
		return []string{}, fmt.Errorf("couldn't read resolv.conf")
	}
	return parseResolve(resolveData)
}
