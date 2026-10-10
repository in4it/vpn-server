package configmanager

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/in4it/wireguard-server/pkg/wireguard/fsutils"
)

var (
	// variables so tests can stub them
	getImmutable = fsutils.GetImmutable
	setImmutable = fsutils.SetImmutable
	// the warning that the filesystem doesn't support the immutable flag is only logged once
	immutableNotSupportedOnce sync.Once
)

// setImmutableFlag sets or clears the immutable flag on filename. When the
// filesystem doesn't support the flag, a warning is logged (once) and nil is
// returned.
func setImmutableFlag(filename string, immutable bool) error {
	err := setImmutable(filename, immutable)
	if err != nil && errors.Is(err, fsutils.ErrImmutableNotSupported) {
		logImmutableNotSupported(err)
		return nil
	}
	return err
}

func logImmutableNotSupported(err error) {
	immutableNotSupportedOnce.Do(func() {
		log.Printf("Warning: could not set the immutable flag on the binaries: %s", err)
	})
}

// ensureImmutableBinaries sets the immutable flag on the binaries in dir when
// it's not set yet
func ensureImmutableBinaries(dir string, binaries map[string]string) {
	for filename := range binaries {
		binaryPath := path.Join(dir, filename)
		if _, err := os.Lstat(binaryPath); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: could not check binary %s: %s", binaryPath, err)
			}
			continue
		}
		immutable, err := getImmutable(binaryPath)
		if err != nil {
			if errors.Is(err, fsutils.ErrImmutableNotSupported) {
				logImmutableNotSupported(err)
				return
			}
			log.Printf("Warning: could not get the immutable flag of %s: %s", binaryPath, err)
			continue
		}
		if immutable {
			continue
		}
		if err := setImmutableFlag(binaryPath, true); err != nil {
			log.Printf("Warning: could not set the immutable flag on %s: %s", binaryPath, err)
		}
	}
}

func executableDir() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("os Executable error: %s", err)
	}
	return path.Dir(executable), nil
}
