package localstorage

import (
	"fmt"
	"io"
	"os"
	"path"
)

func (l *LocalStorage) WriteFile(name string, data []byte) error {
	return os.WriteFile(path.Join(l.path, name), data, 0600)
}

func (l *LocalStorage) AppendFile(name string, data []byte) error {
	f, err := os.OpenFile(path.Join(l.path, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

func (l *LocalStorage) OpenFileForWriting(name string) (io.WriteCloser, error) {
	file, err := os.Create(path.Join(l.path, name))
	if err != nil {
		return nil, fmt.Errorf("cannot open file (%s): %s", name, err)
	}
	return file, nil
}

func (l *LocalStorage) OpenFileForAppending(name string) (io.WriteCloser, error) {
	file, err := os.OpenFile(path.Join(l.path, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	return file, nil
}
