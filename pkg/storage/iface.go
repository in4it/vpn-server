package storage

import (
	"io"
	"io/fs"
)

type Iface interface {
	GetPath() string
	EnsurePath(path string) error
	EnsureOwnership(filename, login string) error
	ReadDir(name string) ([]string, error)
	Remove(name string) error
	Rename(oldName, newName string) error
	AppendFile(name string, data []byte) error
	EnsurePermissions(name string, mode fs.FileMode) error
	FileInfo(name string) (fs.FileInfo, error)
	ReadWriter
	Seeker
}

type ReadWriter interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte) error
	FileExists(filename string) bool
	ConfigPath(filename string) string
	OpenFile(name string) (io.ReadCloser, error)
	OpenFileForWriting(name string) (io.WriteCloser, error)
	OpenFileForAppending(name string) (io.WriteCloser, error)
}

type Seeker interface {
	OpenFilesFromPos(names []string, pos int64) ([]io.ReadCloser, error)
}
