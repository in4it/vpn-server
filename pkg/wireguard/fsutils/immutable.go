package fsutils

import "errors"

// ErrImmutableNotSupported is returned when the filesystem doesn't support the
// immutable inode flag
var ErrImmutableNotSupported = errors.New("immutable flag not supported by the filesystem")
