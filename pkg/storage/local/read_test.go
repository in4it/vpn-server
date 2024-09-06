package localstorage

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"
)

func TestOpenFilesFromPos(t *testing.T) {
	pwd, err := os.Executable()
	if err != nil {
		t.Fatalf("os Executable error: %s", err)
	}
	l := LocalStorage{
		path: path.Dir(pwd),
	}
	contents1 := []byte(`this is the first file`)
	contents2 := []byte(`this is the second file`)
	err = l.WriteFile("1.txt", contents1)
	if err != nil {
		t.Fatalf("write file error: %s", err)
	}
	err = l.WriteFile("2.txt", contents2)
	if err != nil {
		t.Fatalf("write file error: %s", err)
	}
	t.Cleanup(func() {
		err = os.Remove(path.Join(l.path, "1.txt"))
		if err != nil {
			t.Fatalf("file delete error: %s", err)
		}
		err = os.Remove(path.Join(l.path, "2.txt"))
		if err != nil {
			t.Fatalf("file delete error: %s", err)
		}
	})
	expected := []string{
		"this is the first filethis is the second file",
		"is the first filethis is the second file",
		"this is the second file",
		"ethis is the second file",
		"his is the second file",
		"",
		"",
		"",
	}
	expextedOpenFiles := []int{
		2,
		2,
		1,
		2,
		1,
		0,
		0,
		0,
	}
	tests := []int64{
		0,
		5,
		int64(len(contents1)),
		int64(len(contents1) - 1),
		int64(len(contents1) + 1),
		int64(len(contents1) + len(contents2)),
		int64(len(contents1) + len(contents2) + 1),
		-5,
	}
	for k, pos := range tests {
		files, err := l.OpenFilesFromPos([]string{"1.txt", "2.txt"}, pos)
		if err != nil {
			t.Fatalf("open file error: %s", err)
		}
		contents := bytes.NewBuffer([]byte{})
		for _, file := range files {
			defer file.Close()
			body, err := io.ReadAll(file)
			if err != nil {
				t.Fatalf("could not read file: %s", err)
			}
			contents.Write(body)
		}
		if expected[k] != contents.String() {
			t.Fatalf("unexpected output: expected '%s' got '%s'", expected[k], contents.String())
		}
		if expextedOpenFiles[k] != len(files) {
			t.Fatalf("unexpected open files: expected %d got %d", expextedOpenFiles[k], len(files))
		}
	}

}
