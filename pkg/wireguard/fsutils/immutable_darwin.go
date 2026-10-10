//go:build darwin

package fsutils

// GetImmutable is not implemented on darwin
func GetImmutable(filename string) (bool, error) {
	return false, nil
}

// SetImmutable is not implemented on darwin
func SetImmutable(filename string, immutable bool) error {
	return nil
}
