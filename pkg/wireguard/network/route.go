package network

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parseDefaultGwInterface parses /proc/net/route and returns the interface of
// the default route (destination and mask 00000000). When there are multiple
// default routes, the one with the lowest metric is returned.
func parseDefaultGwInterface(r io.Reader) (string, error) {
	const (
		colIface       = 0
		colDestination = 1
		colFlags       = 3
		colMetric      = 6
		colMask        = 7
		rtfUp          = 0x1
	)
	scanner := bufio.NewScanner(r)
	iface := ""
	lowestMetric := int64(-1)
	for i := 0; scanner.Scan(); i++ {
		if i == 0 { // header
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) <= colMask {
			continue
		}
		if fields[colDestination] != "00000000" || fields[colMask] != "00000000" {
			continue
		}
		flags, err := strconv.ParseInt(fields[colFlags], 16, 64)
		if err != nil || flags&rtfUp == 0 {
			continue
		}
		metric, err := strconv.ParseInt(fields[colMetric], 10, 64)
		if err != nil {
			continue
		}
		if iface == "" || metric < lowestMetric {
			iface = fields[colIface]
			lowestMetric = metric
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("could not read routes: %s", err)
	}
	if iface == "" {
		return "", fmt.Errorf("could not determine default gw: no default route found")
	}
	return iface, nil
}
