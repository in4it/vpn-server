package network

import (
	"bufio"
	"bytes"
	"fmt"
	"slices"
	"strings"
)

func parseResolve(buf []byte) ([]string, error) {
	nameservers := []string{}
	scanner := bufio.NewScanner(bytes.NewBuffer(buf))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver ") {
			nameservers = append(nameservers, strings.Replace(line, "nameserver ", "", -1))
		}
	}
	if err := scanner.Err(); err != nil {
		return []string{}, fmt.Errorf("scanner error: %s", err)
	}
	slices.Sort(nameservers)
	return slices.Compact(nameservers), nil
}
