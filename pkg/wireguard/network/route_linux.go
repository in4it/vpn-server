//go:build linux
// +build linux

package network

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetInterfaceDefaultGw() (string, error) {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "", fmt.Errorf("could not open /proc/net/route: %s", err)
	}
	defer file.Close() //nolint:errcheck

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		for i >= 1 { // 1 is first line, which contains the default gateway
			elements := strings.Split(scanner.Text(), "\t")
			return elements[0], nil
		}
	}
	return "", fmt.Errorf("could not determine default gw")
}
