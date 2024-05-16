package localstorage

import (
	"fmt"
	"os"
	"path"
)

func New() (*LocalStorage, error) {
	pwd, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("os Executable error: %s", err)
	}

	pathname := path.Dir(pwd)
	storage, err := NewWithPath(pathname)
	if err != nil {
		return storage, err
	}
	err = storage.EnsurePath(CONFIG_PATH)
	if err != nil {
		return nil, fmt.Errorf("cannot create storage directories: %s", err)
	}
	err = storage.EnsurePath(path.Join(CONFIG_PATH, VPN_CLIENTS_DIR))
	if err != nil {
		return nil, fmt.Errorf("cannot create storage directories: %s", err)
	}
	err = storage.EnsureOwnership(CONFIG_PATH, "vpn")
	if err != nil {
		return nil, fmt.Errorf("cannot ensure vpn ownership of config directory: %s", err)
	}
	err = storage.EnsureOwnership(path.Join(CONFIG_PATH, VPN_CLIENTS_DIR), "vpn")
	if err != nil {
		return nil, fmt.Errorf("cannot ensure vpn ownership of config directory: %s", err)
	}

	return NewWithPath(pathname)
}

func NewWithPath(pathname string) (*LocalStorage, error) {
	return &LocalStorage{
		path: pathname,
	}, nil
}
