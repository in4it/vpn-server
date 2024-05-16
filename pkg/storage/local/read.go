package localstorage

import (
	"os"
	"path"
)

func (l *LocalStorage) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(path.Join(l.path, name))
}
