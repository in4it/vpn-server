//go:build darwin
// +build darwin

package network

func GetInterfaceDefaultGw() (string, error) {
	return "notsupported(Darwin)", nil
}
