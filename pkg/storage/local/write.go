package localstorage

import (
	"os"
	"path"
)

func (l *LocalStorage) WriteFile(name string, data []byte) error {
	return os.WriteFile(path.Join(l.path, name), data, 0600)
}

func (l *LocalStorage) AppendFile(name string, data []byte) error {
	f, err := os.OpenFile("text.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}
