package fsutils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/in4it/go-devops-platform/logging"
	"golang.org/x/sys/unix"
)

// ErrSymlink is returned when a file operation is refused because the target
// (or one of the directories leading to it) is a symbolic link.
var ErrSymlink = errors.New("refusing to follow symbolic link")

// The functions in this file are used by the configmanager to write, append
// and change ownership of files in the vpn directory. Paths are resolved
// relative to a base directory, one component at a time, without following
// symbolic links. Only the base directory itself is resolved normally.

func symlinkError(base, name string) error {
	err := fmt.Errorf("%w: %s", ErrSymlink, filepath.Join(base, name))
	logging.ErrorLog(err)
	return err
}

// splitRelPath cleans name and returns its components relative to base. name
// can be relative to base, or an absolute path within base.
func splitRelPath(base, name string) ([]string, error) {
	base = filepath.Clean(base)
	rel := name
	if filepath.IsAbs(name) {
		var err error
		rel, err = filepath.Rel(base, filepath.Clean(name))
		if err != nil {
			return nil, fmt.Errorf("path %s is outside of %s", name, base)
		}
	}
	rel = filepath.Clean(rel)
	if rel == "." {
		return []string{}, nil
	}
	if filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path %s is outside of %s", name, base)
	}
	return strings.Split(rel, string(filepath.Separator)), nil
}

func isSymlinkAt(dirfd int, name string) bool {
	var stat unix.Stat_t
	if err := unix.Fstatat(dirfd, name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return false
	}
	return stat.Mode&unix.S_IFMT == unix.S_IFLNK
}

// openParentDir opens the parent directory of name (relative to base) without
// following symbolic links, and returns its file descriptor and the last path
// component. The caller must close the returned file descriptor.
func openParentDir(base, name string) (int, string, error) {
	parts, err := splitRelPath(base, name)
	if err != nil {
		return -1, "", err
	}
	if len(parts) == 0 {
		return -1, "", fmt.Errorf("invalid path: %s", name)
	}
	dirfd, err := unix.Open(filepath.Clean(base), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, "", &os.PathError{Op: "open", Path: base, Err: err}
	}
	for i, part := range parts[:len(parts)-1] {
		fd, err := unix.Openat(dirfd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if err != nil {
			symlink := isSymlinkAt(dirfd, part)
			unix.Close(dirfd) //nolint:errcheck
			if symlink {
				return -1, "", symlinkError(base, filepath.Join(parts[:i+1]...))
			}
			return -1, "", &os.PathError{Op: "open", Path: filepath.Join(base, filepath.Join(parts[:i+1]...)), Err: err}
		}
		unix.Close(dirfd) //nolint:errcheck
		dirfd = fd
	}
	return dirfd, parts[len(parts)-1], nil
}

func closeFd(fd int) {
	unix.Close(fd) //nolint:errcheck
}

// lstatAt returns the stat of name in dirfd, without following symbolic links
func lstatAt(dirfd int, name string) (unix.Stat_t, error) {
	var stat unix.Stat_t
	err := unix.Fstatat(dirfd, name, &stat, unix.AT_SYMLINK_NOFOLLOW)
	return stat, err
}

// OpenFileNoFollow opens name (relative to base) with the given flags. It
// refuses symbolic links and anything else than regular files.
func OpenFileNoFollow(base, name string, flag int, perm fs.FileMode) (*os.File, error) {
	dirfd, last, err := openParentDir(base, name)
	if err != nil {
		return nil, err
	}
	defer unix.Close(dirfd) //nolint:errcheck
	return openAt(base, name, dirfd, last, flag, perm)
}

func openAt(base, name string, dirfd int, last string, flag int, perm fs.FileMode) (*os.File, error) {
	fullPath := filepath.Join(base, name)
	// truncate only after the checks below
	truncate := flag&unix.O_TRUNC != 0
	flag &^= unix.O_TRUNC
	fd, err := unix.Openat(dirfd, last, flag|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, uint32(perm.Perm()))
	if err != nil {
		if isSymlinkAt(dirfd, last) {
			return nil, symlinkError(base, name)
		}
		return nil, &os.PathError{Op: "open", Path: fullPath, Err: err}
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		unix.Close(fd) //nolint:errcheck
		return nil, &os.PathError{Op: "stat", Path: fullPath, Err: err}
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		unix.Close(fd) //nolint:errcheck
		return nil, fmt.Errorf("not a regular file: %s", fullPath)
	}
	if flag&(unix.O_WRONLY|unix.O_RDWR) != 0 && stat.Nlink > 1 {
		unix.Close(fd) //nolint:errcheck
		return nil, fmt.Errorf("refusing to write to a file with multiple hard links: %s", fullPath)
	}
	if err := unix.SetNonblock(fd, false); err != nil {
		unix.Close(fd) //nolint:errcheck
		return nil, &os.PathError{Op: "open", Path: fullPath, Err: err}
	}
	if truncate {
		if err := unix.Ftruncate(fd, 0); err != nil {
			unix.Close(fd) //nolint:errcheck
			return nil, &os.PathError{Op: "truncate", Path: fullPath, Err: err}
		}
	}
	return os.NewFile(uintptr(fd), fullPath), nil
}

// WriteFileNoFollow writes data to name (relative to base) atomically: the
// data is written to a temporary file in the same directory, which is then
// renamed to name. When the file already exists, its mode and (when running
// as root) its ownership are kept. New files are created with mode perm.
func WriteFileNoFollow(base, name string, data []byte, perm fs.FileMode) error {
	dirfd, last, err := openParentDir(base, name)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd) //nolint:errcheck

	mode := uint32(perm.Perm())
	uid, gid := -1, -1
	existing, err := lstatAt(dirfd, last)
	if err == nil {
		switch existing.Mode & unix.S_IFMT {
		case unix.S_IFLNK:
			return symlinkError(base, name)
		case unix.S_IFREG:
			mode = uint32(existing.Mode) & 07777
			if os.Geteuid() == 0 {
				uid, gid = int(existing.Uid), int(existing.Gid)
			}
		default:
			return fmt.Errorf("not a regular file: %s", filepath.Join(base, name))
		}
	} else if !errors.Is(err, unix.ENOENT) {
		return &os.PathError{Op: "lstat", Path: filepath.Join(base, name), Err: err}
	}

	tmpName, tmpFile, err := createTempAt(base, dirfd, last)
	if err != nil {
		return err
	}
	renamed := false
	defer func() {
		if !renamed {
			tmpFile.Close()                  //nolint:errcheck
			unix.Unlinkat(dirfd, tmpName, 0) //nolint:errcheck
		}
	}()
	if uid != -1 {
		if err := tmpFile.Chown(uid, gid); err != nil {
			return fmt.Errorf("chown error: %s", err)
		}
	}
	if err := unix.Fchmod(int(tmpFile.Fd()), mode); err != nil {
		return fmt.Errorf("chmod error: %s", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := unix.Renameat(dirfd, tmpName, dirfd, last); err != nil {
		return &os.PathError{Op: "rename", Path: filepath.Join(base, name), Err: err}
	}
	renamed = true
	unix.Fsync(dirfd) //nolint:errcheck
	return nil
}

func createTempAt(base string, dirfd int, last string) (string, *os.File, error) {
	for i := 0; i < 10; i++ {
		random := make([]byte, 8)
		if _, err := rand.Read(random); err != nil {
			return "", nil, err
		}
		tmpName := "." + last + ".tmp-" + hex.EncodeToString(random)
		fd, err := unix.Openat(dirfd, tmpName, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
		if errors.Is(err, unix.EEXIST) {
			continue
		}
		if err != nil {
			return "", nil, &os.PathError{Op: "create", Path: filepath.Join(base, tmpName), Err: err}
		}
		return tmpName, os.NewFile(uintptr(fd), filepath.Join(base, tmpName)), nil
	}
	return "", nil, fmt.Errorf("could not create temporary file for %s", filepath.Join(base, last))
}

// AppendFileNoFollow appends data to name (relative to base), creating it when
// it doesn't exist.
func AppendFileNoFollow(base, name string, data []byte, perm fs.FileMode) error {
	f, err := OpenFileNoFollow(base, name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close() //nolint:errcheck
		return err
	}
	return f.Close()
}

// MkdirNoFollow creates the directory name (relative to base) when it doesn't
// exist yet. An existing symbolic link is refused.
func MkdirNoFollow(base, name string, perm fs.FileMode) error {
	dirfd, last, err := openParentDir(base, name)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd) //nolint:errcheck
	stat, err := lstatAt(dirfd, last)
	if err == nil {
		switch stat.Mode & unix.S_IFMT {
		case unix.S_IFDIR:
			return nil
		case unix.S_IFLNK:
			return symlinkError(base, name)
		default:
			return fmt.Errorf("not a directory: %s", filepath.Join(base, name))
		}
	}
	if err := unix.Mkdirat(dirfd, last, uint32(perm.Perm())); err != nil {
		if errors.Is(err, unix.EEXIST) && isSymlinkAt(dirfd, last) {
			return symlinkError(base, name)
		}
		if errors.Is(err, unix.EEXIST) {
			return nil
		}
		return fmt.Errorf("create directory error: %s", &os.PathError{Op: "mkdir", Path: filepath.Join(base, name), Err: err})
	}
	return nil
}

// LchownNoFollow changes the ownership of name (relative to base). Symbolic
// links and files with multiple hard links are refused.
func LchownNoFollow(base, name string, uid, gid int) error {
	dirfd, last, err := openParentDir(base, name)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd) //nolint:errcheck
	stat, err := lstatAt(dirfd, last)
	if err != nil {
		return &os.PathError{Op: "lstat", Path: filepath.Join(base, name), Err: err}
	}
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFLNK:
		return symlinkError(base, name)
	case unix.S_IFDIR:
	case unix.S_IFREG:
		if stat.Nlink > 1 {
			return fmt.Errorf("refusing to change ownership of a file with multiple hard links: %s", filepath.Join(base, name))
		}
	default:
		return fmt.Errorf("not a regular file or directory: %s", filepath.Join(base, name))
	}
	if err := unix.Fchownat(dirfd, last, uid, gid, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return &os.PathError{Op: "lchown", Path: filepath.Join(base, name), Err: err}
	}
	return nil
}

// ChmodNoFollow changes the mode of name (relative to base), which needs to
// be a regular file or a directory. Symbolic links are refused.
func ChmodNoFollow(base, name string, mode fs.FileMode) error {
	dirfd, last, err := openParentDir(base, name)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd) //nolint:errcheck
	fd, err := unix.Openat(dirfd, last, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		if isSymlinkAt(dirfd, last) {
			return symlinkError(base, name)
		}
		return &os.PathError{Op: "open", Path: filepath.Join(base, name), Err: err}
	}
	defer unix.Close(fd) //nolint:errcheck
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return &os.PathError{Op: "stat", Path: filepath.Join(base, name), Err: err}
	}
	if t := stat.Mode & unix.S_IFMT; t != unix.S_IFREG && t != unix.S_IFDIR {
		return fmt.Errorf("not a regular file or directory: %s", filepath.Join(base, name))
	}
	if err := unix.Fchmod(fd, unixMode(mode)); err != nil {
		return &os.PathError{Op: "chmod", Path: filepath.Join(base, name), Err: err}
	}
	return nil
}

func unixMode(mode fs.FileMode) uint32 {
	m := uint32(mode.Perm())
	if mode&fs.ModeSetuid != 0 {
		m |= unix.S_ISUID
	}
	if mode&fs.ModeSetgid != 0 {
		m |= unix.S_ISGID
	}
	if mode&fs.ModeSticky != 0 {
		m |= unix.S_ISVTX
	}
	return m
}

// RemoveNoFollow removes the file or empty directory name (relative to base).
// A symbolic link itself is removed, not its target.
func RemoveNoFollow(base, name string) error {
	dirfd, last, err := openParentDir(base, name)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd) //nolint:errcheck
	err = unix.Unlinkat(dirfd, last, 0)
	if err == nil {
		return nil
	}
	if errDir := unix.Unlinkat(dirfd, last, unix.AT_REMOVEDIR); errDir == nil {
		return nil
	}
	return &os.PathError{Op: "remove", Path: filepath.Join(base, name), Err: err}
}

// RenameNoFollow renames oldName to newName (both relative to base)
func RenameNoFollow(base, oldName, newName string) error {
	oldDirfd, oldLast, err := openParentDir(base, oldName)
	if err != nil {
		return err
	}
	defer unix.Close(oldDirfd) //nolint:errcheck
	newDirfd, newLast, err := openParentDir(base, newName)
	if err != nil {
		return err
	}
	defer unix.Close(newDirfd) //nolint:errcheck
	if err := unix.Renameat(oldDirfd, oldLast, newDirfd, newLast); err != nil {
		return &os.LinkError{Op: "rename", Old: filepath.Join(base, oldName), New: filepath.Join(base, newName), Err: err}
	}
	return nil
}
