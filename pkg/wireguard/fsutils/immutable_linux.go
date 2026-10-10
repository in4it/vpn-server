//go:build linux

package fsutils

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// FS_IMMUTABLE_FL from linux/fs.h (not defined in golang.org/x/sys/unix)
const fsImmutableFl = 0x00000010

// GetImmutable returns whether the immutable inode flag is set on filename.
// ErrImmutableNotSupported is returned when the filesystem doesn't support it.
func GetImmutable(filename string) (bool, error) {
	f, err := openForInodeFlags(filename)
	if err != nil {
		return false, err
	}
	defer f.Close() //nolint:errcheck
	flags, err := unix.IoctlGetUint32(int(f.Fd()), unix.FS_IOC_GETFLAGS)
	if err != nil {
		return false, inodeFlagsError(filename, err)
	}
	return flags&fsImmutableFl != 0, nil
}

// SetImmutable sets or clears the immutable inode flag on filename.
// ErrImmutableNotSupported is returned when the filesystem doesn't support it.
func SetImmutable(filename string, immutable bool) error {
	f, err := openForInodeFlags(filename)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	flags, err := unix.IoctlGetUint32(int(f.Fd()), unix.FS_IOC_GETFLAGS)
	if err != nil {
		return inodeFlagsError(filename, err)
	}
	newFlags := flags &^ fsImmutableFl
	if immutable {
		newFlags = flags | fsImmutableFl
	}
	if newFlags == flags {
		return nil
	}
	if err := unix.IoctlSetPointerInt(int(f.Fd()), unix.FS_IOC_SETFLAGS, int(int32(newFlags))); err != nil {
		return inodeFlagsError(filename, err)
	}
	return nil
}

func openForInodeFlags(filename string) (*os.File, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close() //nolint:errcheck
		return nil, err
	}
	if !info.Mode().IsRegular() {
		f.Close() //nolint:errcheck
		return nil, fmt.Errorf("not a regular file: %s", filename)
	}
	return f, nil
}

func inodeFlagsError(filename string, err error) error {
	if errors.Is(err, unix.ENOTTY) || errors.Is(err, unix.EOPNOTSUPP) || errors.Is(err, unix.EINVAL) {
		return fmt.Errorf("%w (%s): %s", ErrImmutableNotSupported, filename, err)
	}
	return fmt.Errorf("inode flags error (%s): %s", filename, err)
}
