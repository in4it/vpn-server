package s3storage

import (
	"io/fs"
	"strings"
)

func (l *S3Storage) FileExists(filename string) bool {
	return false
}

func (l *S3Storage) ConfigPath(filename string) string {
	return CONFIG_PATH + "/" + strings.TrimLeft(filename, "/")
}

func (s *S3Storage) GetPath() string {
	return s.prefix
}

func (l *S3Storage) EnsurePath(pathname string) error {
	return nil
}

func (l *S3Storage) EnsureOwnership(filename, login string) error {
	return nil
}

func (l *S3Storage) Remove(name string) error {
	return nil
}

func (l *S3Storage) Rename(oldName, newName string) error {
	return nil
}

func (l *S3Storage) EnsurePermissions(name string, mode fs.FileMode) error {
	return nil
}

func (l *S3Storage) FileInfo(name string) (fs.FileInfo, error) {
	return nil, nil
}
