package s3storage

import (
	"io"
)

func (l *S3Storage) ReadFile(name string) ([]byte, error) {
	return nil, nil
}

func (l *S3Storage) OpenFilesFromPos(names []string, pos int64) ([]io.ReadCloser, error) {
	return nil, nil
}

func (l *S3Storage) OpenFile(name string) (io.ReadCloser, error) {
	return nil, nil
}
