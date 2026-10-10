package fsutils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path"
	"strconv"

	"github.com/in4it/go-devops-platform/logging"
	"github.com/in4it/go-devops-platform/storage"
	localstorage "github.com/in4it/go-devops-platform/storage/local"
)

// NoFollowStorage is the local storage used by the configmanager. Files are
// written, appended, read, chowned and chmodded without following symbolic
// links (see nofollow.go). The remaining operations are handled by the local
// storage.
type NoFollowStorage struct {
	storage.Iface
	base string
}

// NewNoFollowStorage returns a NoFollowStorage with pathname as base directory
func NewNoFollowStorage(pathname string) (*NoFollowStorage, error) {
	local, err := localstorage.NewWithPath(pathname)
	if err != nil {
		return nil, err
	}
	return &NoFollowStorage{Iface: local, base: pathname}, nil
}

// NewNoFollowStorageWithConfigDirs returns a NoFollowStorage and makes sure
// the config directories exist and are owned by the vpn user (like
// localstorage.New does)
func NewNoFollowStorageWithConfigDirs(pathname string) (*NoFollowStorage, error) {
	s, err := NewNoFollowStorage(pathname)
	if err != nil {
		return nil, err
	}
	for _, dir := range []string{localstorage.CONFIG_PATH, path.Join(localstorage.CONFIG_PATH, localstorage.VPN_CLIENTS_DIR)} {
		if err := s.EnsurePath(dir); err != nil {
			return nil, fmt.Errorf("cannot create storage directories: %s", err)
		}
		if err := s.EnsureOwnership(dir, VPN_USER); err != nil {
			return nil, fmt.Errorf("cannot ensure vpn ownership of config directory: %s", err)
		}
	}
	return s, nil
}

func (s *NoFollowStorage) GetPath() string {
	return s.base
}

func (s *NoFollowStorage) EnsurePath(pathname string) error {
	return MkdirNoFollow(s.base, pathname, 0700)
}

func (s *NoFollowStorage) EnsureOwnership(filename, login string) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %s", err)
	}
	if currentUser.Uid != "0" {
		// only root can change ownership. Still refuse symbolic links, so
		// the behaviour is the same as when running as root.
		dirfd, last, err := openParentDir(s.base, filename)
		if err != nil {
			return err
		}
		defer closeFd(dirfd)
		if isSymlinkAt(dirfd, last) {
			return symlinkError(s.base, filename)
		}
		logging.DebugLog(fmt.Errorf("cannot ensure ownership of file %s when not user root (current user: %s)", filename, currentUser.Username))
		return nil
	}
	u, err := user.Lookup(login)
	if err != nil {
		return fmt.Errorf("user lookup error (%s): %s", login, err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("user lookup error (uid): %s", err)
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("user lookup error (gid): %s", err)
	}
	if err := LchownNoFollow(s.base, filename, uid, gid); err != nil {
		return fmt.Errorf("vpn chown error: %s", err)
	}
	return nil
}

func (s *NoFollowStorage) EnsurePermissions(name string, mode fs.FileMode) error {
	return ChmodNoFollow(s.base, name, mode)
}

func (s *NoFollowStorage) Remove(name string) error {
	return RemoveNoFollow(s.base, name)
}

func (s *NoFollowStorage) Rename(oldName, newName string) error {
	return RenameNoFollow(s.base, oldName, newName)
}

func (s *NoFollowStorage) AppendFile(name string, data []byte) error {
	return AppendFileNoFollow(s.base, name, data, 0660)
}

func (s *NoFollowStorage) WriteFile(name string, data []byte) error {
	return WriteFileNoFollow(s.base, name, data, 0600)
}

func (s *NoFollowStorage) ReadFile(name string) ([]byte, error) {
	f, err := OpenFileNoFollow(s.base, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck
	return io.ReadAll(f)
}

func (s *NoFollowStorage) OpenFile(name string) (io.ReadCloser, error) {
	return OpenFileNoFollow(s.base, name, os.O_RDONLY, 0)
}

func (s *NoFollowStorage) OpenFileForWriting(name string) (io.WriteCloser, error) {
	f, err := OpenFileNoFollow(s.base, name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("cannot open file (%s): %w", name, err)
	}
	return f, nil
}

func (s *NoFollowStorage) OpenFileForAppending(name string) (io.WriteCloser, error) {
	return OpenFileNoFollow(s.base, name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
}
