package localstorage

import (
	"os"
	"path"
)

func (l *LocalStorage) WriteFile(name string, data []byte) error {
	return os.WriteFile(path.Join(l.path, name), data, 0600)
}
