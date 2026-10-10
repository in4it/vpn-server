//go:build darwin

package network

// GetInterfaceDefaultGw is not implemented on darwin (development only). It
// returns a valid interface name, so the server config can be rendered.
func GetInterfaceDefaultGw() (string, error) {
	return "en0", nil
}
