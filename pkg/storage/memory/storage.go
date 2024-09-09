package memorystorage

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

type MyWriteCloser struct {
	*bufio.Writer
}

func (mwc *MyWriteCloser) Close() error {
	return nil
}

type MockReadWriterData []byte

func (m *MockReadWriterData) Close() error {
	return nil
}
func (m *MockReadWriterData) Write(p []byte) (nn int, err error) {
	*m = append(*m, p...)
	return len(p), nil
}

type MockMemoryStorage struct {
	Data map[string]*MockReadWriterData
}

func (m *MockMemoryStorage) ConfigPath(filename string) string {
	return path.Join("config", filename)
}
func (m *MockMemoryStorage) Rename(oldName, newName string) error {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	_, ok := m.Data[oldName]
	if !ok {
		return fmt.Errorf("file doesn't exist")
	}
	m.Data[newName] = m.Data[oldName]
	delete(m.Data, oldName)
	return nil
}
func (m *MockMemoryStorage) FileExists(name string) bool {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	_, ok := m.Data[name]
	return ok
}

func (m *MockMemoryStorage) ReadFile(name string) ([]byte, error) {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	val, ok := m.Data[name]
	if !ok {
		return nil, fmt.Errorf("file does not exist")
	}
	return *val, nil
}
func (m *MockMemoryStorage) WriteFile(name string, data []byte) error {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	m.Data[name] = (*MockReadWriterData)(&data)
	return nil
}
func (m *MockMemoryStorage) AppendFile(name string, data []byte) error {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	if m.Data[name] == nil {
		m.Data[name] = (*MockReadWriterData)(&data)
	} else {
		*m.Data[name] = append(*m.Data[name], data...)
	}

	return nil
}

func (m *MockMemoryStorage) GetPath() string {
	pwd, _ := os.Executable()
	return path.Dir(pwd)
}

func (m *MockMemoryStorage) EnsurePath(pathname string) error {
	return nil
}

func (m *MockMemoryStorage) EnsureOwnership(filename, login string) error {
	return nil
}

func (m *MockMemoryStorage) ReadDir(path string) ([]string, error) {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	res := []string{}
	for k := range m.Data {
		if strings.HasPrefix(k, path+"/") {
			res = append(res, strings.ReplaceAll(k, path+"/", ""))
		}
	}
	return res, nil
}

func (m *MockMemoryStorage) Remove(name string) error {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	_, ok := m.Data[name]
	if !ok {
		return fmt.Errorf("file does not exist")
	}
	delete(m.Data, name)
	return nil
}

func (m *MockMemoryStorage) OpenFilesFromPos(names []string, pos int64) ([]io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockMemoryStorage) OpenFile(name string) (io.ReadCloser, error) {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	val, ok := m.Data[name]
	if !ok {
		return nil, fmt.Errorf("file does not exist")
	}

	return io.NopCloser(bytes.NewBuffer(*val)), nil
}
func (m *MockMemoryStorage) OpenFileForWriting(name string) (io.WriteCloser, error) {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	m.Data[name] = (*MockReadWriterData)(&[]byte{})
	return m.Data[name], nil
}
func (m *MockMemoryStorage) OpenFileForAppending(name string) (io.WriteCloser, error) {
	if m.Data == nil {
		m.Data = make(map[string]*MockReadWriterData)
	}
	val, ok := m.Data[name]
	if !ok {
		m.Data[name] = (*MockReadWriterData)(&[]byte{})
		return m.Data[name], nil
	}
	m.Data[name] = (*MockReadWriterData)(val)
	return m.Data[name], nil
}
func (m *MockMemoryStorage) EnsurePermissions(name string, mode fs.FileMode) error {
	return nil
}
