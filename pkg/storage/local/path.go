package localstorage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path"
	"strconv"

	"github.com/in4it/wireguard-server/pkg/logging"
)

func (l *LocalStorage) FileExists(filename string) bool {
	if _, err := os.Stat(path.Join(l.path, filename)); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (l *LocalStorage) ConfigPath(filename string) string {
	return path.Join(CONFIG_PATH, filename)
}

func (l *LocalStorage) GetPath() string {
	return l.path
}

func (l *LocalStorage) EnsurePath(pathname string) error {
	fullPathname := path.Join(l.path, pathname)
	if _, err := os.Stat(fullPathname); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(fullPathname, 0700)
		if err != nil {
			return fmt.Errorf("create directory error: %s", err)
		}
	}
	return nil
}

func (l *LocalStorage) EnsureOwnership(filename, login string) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %s", err)
	}
	if currentUser.Username != "root" {
		logging.DebugLog(fmt.Errorf("cannot ensure ownership of file %s when not user root (current user: %s)", filename, currentUser.Username))
		return nil
	}
	vpnUser, err := user.Lookup(login)
	if err != nil {
		return fmt.Errorf("user lookup error (vpn): %s", err)
	}
	vpnUserUid, err := strconv.Atoi(vpnUser.Uid)
	if err != nil {
		return fmt.Errorf("user lookup error (uid): %s", err)
	}
	vpnUserGid, err := strconv.Atoi(vpnUser.Gid)
	if err != nil {
		return fmt.Errorf("user lookup error (gid): %s", err)
	}

	err = os.Chown(path.Join(l.path, filename), vpnUserUid, vpnUserGid)
	if err != nil {
		return fmt.Errorf("vpn chown error: %s", err)
	}
	return nil
}

func (l *LocalStorage) ReadDir(pathname string) ([]string, error) {
	res, err := os.ReadDir(path.Join(l.path, pathname))
	if err != nil {
		return []string{}, err
	}
	resNames := make([]string, len(res))
	for k, v := range res {
		resNames[k] = v.Name()
	}
	return resNames, nil
}

func (l *LocalStorage) Remove(name string) error {
	return os.Remove(path.Join(l.path, name))
}

func (l *LocalStorage) Rename(oldName, newName string) error {
	return os.Rename(path.Join(l.path, oldName), path.Join(l.path, newName))
}

func (l *LocalStorage) EnsurePermissions(name string, mode fs.FileMode) error {
	return os.Chmod(path.Join(l.path, name), mode)
}
