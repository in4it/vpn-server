package testingmocks

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type MockReadWriter struct {
	Data map[string][]byte
}

func (m *MockReadWriter) ConfigPath(filename string) string {
	return path.Join("config", filename)
}
func (m *MockReadWriter) FileExists(name string) bool {
	if m.Data == nil {
		m.Data = make(map[string][]byte)
	}
	_, ok := m.Data[name]
	return ok
}

func (m *MockReadWriter) ReadFile(name string) ([]byte, error) {
	if m.Data == nil {
		m.Data = make(map[string][]byte)
	}
	val, ok := m.Data[name]
	if !ok {
		return val, fmt.Errorf("file does not exist")
	}
	return val, nil
}
func (m *MockReadWriter) WriteFile(name string, data []byte) error {
	if m.Data == nil {
		m.Data = make(map[string][]byte)
	}
	m.Data[name] = data
	return nil
}

type MockMemoryStorage struct {
	Data map[string][]byte
}

func (m *MockMemoryStorage) ConfigPath(filename string) string {
	return path.Join("config", filename)
}
func (m *MockMemoryStorage) FileExists(name string) bool {
	if m.Data == nil {
		m.Data = make(map[string][]byte)
	}
	_, ok := m.Data[name]
	return ok
}

func (m *MockMemoryStorage) ReadFile(name string) ([]byte, error) {
	if m.Data == nil {
		m.Data = make(map[string][]byte)
	}
	val, ok := m.Data[name]
	if !ok {
		return val, fmt.Errorf("file does not exist")
	}
	return val, nil
}
func (m *MockMemoryStorage) WriteFile(name string, data []byte) error {
	if m.Data == nil {
		m.Data = make(map[string][]byte)
	}
	m.Data[name] = data
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
		m.Data = make(map[string][]byte)
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
		m.Data = make(map[string][]byte)
	}
	delete(m.Data, name)
	return nil
}
