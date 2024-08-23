package storage

type Iface interface {
	GetPath() string
	EnsurePath(path string) error
	EnsureOwnership(filename, login string) error
	ReadDir(name string) ([]string, error)
	Remove(name string) error
	AppendFile(name string, data []byte) error
	ReadWriter
}

type ReadWriter interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte) error
	FileExists(filename string) bool
	ConfigPath(filename string) string
}
