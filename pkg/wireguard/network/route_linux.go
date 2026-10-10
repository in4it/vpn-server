//go:build linux

package network

import (
	"fmt"
	"os"
)

func GetInterfaceDefaultGw() (string, error) {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "", fmt.Errorf("could not open /proc/net/route: %s", err)
	}
	defer file.Close() //nolint:errcheck

	return parseDefaultGwInterface(file)
}
