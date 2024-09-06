package localstorage

import (
	"fmt"
	"io"
	"os"
	"path"
)

func (l *LocalStorage) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(path.Join(l.path, name))
}

func (l *LocalStorage) OpenFilesFromPos(names []string, pos int64) ([]io.ReadCloser, error) {
	readers := []io.ReadCloser{}
	for _, name := range names {
		file, err := os.Open(path.Join(l.path, name))
		if err != nil {
			return nil, fmt.Errorf("cannot open file (%s): %s", name, err)
		}
		stat, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("cannot get file stat (%s): %s", name, err)
		}
		if stat.Size() <= pos {
			pos -= stat.Size()
		} else {
			_, err := file.Seek(pos, 0)
			if err != nil {
				return nil, fmt.Errorf("could not seek to pos (file: %s): %s", name, err)
			}
			pos = 0
			readers = append(readers, file)
		}
	}
	return readers, nil
}
