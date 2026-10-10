package vpn

import (
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
)

// interface names are interpolated in the PostUp/PostDown shell commands of the server config
var externalInterfaceRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,15}$`)

func validateAddressRange(addressRange netip.Prefix) error {
	if !addressRange.Addr().Is4() {
		return fmt.Errorf("AddressRange must be an IPv4 range (e.g. 10.189.184.1/21)")
	}
	return nil
}

// validateClientAddressPrefix checks that the client address prefix has the format /N and that it
// yields a valid prefix within the address range. The client address is the next free IP + prefix, so
// the prefix must be more specific than the address range (otherwise no address can be allocated).
func validateClientAddressPrefix(addressRange netip.Prefix, clientAddressPrefix string) error {
	if !strings.HasPrefix(clientAddressPrefix, "/") {
		return fmt.Errorf("client address prefix must be in the format /N (e.g. /32)")
	}
	bits, err := strconv.Atoi(strings.TrimPrefix(clientAddressPrefix, "/"))
	if err != nil || strconv.Itoa(bits) != strings.TrimPrefix(clientAddressPrefix, "/") {
		return fmt.Errorf("client address prefix must be in the format /N (e.g. /32)")
	}
	if bits <= addressRange.Bits() || bits > addressRange.Addr().BitLen() {
		return fmt.Errorf("client address prefix must be between /%d and /%d for address range %s", addressRange.Bits()+1, addressRange.Addr().BitLen(), addressRange.String())
	}
	if _, err := netip.ParsePrefix(addressRange.Addr().String() + clientAddressPrefix); err != nil {
		return fmt.Errorf("client address prefix is invalid: %s", err)
	}
	return nil
}

func validatePort(port string) (int, error) {
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("port in wrong format: %s", err)
	}
	if portNumber < 1 || portNumber > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}
	return portNumber, nil
}

// validateExternalInterface: an empty interface is allowed (default gateway interface couldn't be determined)
func validateExternalInterface(externalInterface string) error {
	if externalInterface == "" {
		return nil
	}
	if !externalInterfaceRegex.MatchString(externalInterface) {
		return fmt.Errorf("external interface name is invalid (allowed: letters, digits, '_', '.', '-', maximum 15 characters)")
	}
	return nil
}
