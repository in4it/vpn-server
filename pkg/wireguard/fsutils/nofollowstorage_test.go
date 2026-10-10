package fsutils

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// setupSymlinkDir creates a base directory with a "target" directory outside
// of it, containing a file that must never be modified
func setupSymlinkDir(t *testing.T) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	base := filepath.Join(tmp, "base")
	target := filepath.Join(tmp, "target")
	for _, dir := range []string{base, target, filepath.Join(base, "stats")} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(target, "file"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	return base, target
}

func checkTargetUnchanged(t *testing.T, target string) {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(target, "file"))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "original" {
		t.Fatalf("file outside of the base directory was modified: %q", contents)
	}
	info, err := os.Stat(filepath.Join(target, "file"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions of file outside of the base directory changed: %s", info.Mode().Perm())
	}
}

func TestNoFollowStorageRefusesSymlinkedFile(t *testing.T) {
	base, target := setupSymlinkDir(t)
	if err := os.Symlink(filepath.Join(target, "file"), filepath.Join(base, "stats", "link.log")); err != nil {
		t.Fatal(err)
	}
	s, err := NewNoFollowStorage(base)
	if err != nil {
		t.Fatal(err)
	}
	name := "stats/link.log"
	operations := map[string]func() error{
		"AppendFile": func() error { return s.AppendFile(name, []byte("appended")) },
		"WriteFile":  func() error { return s.WriteFile(name, []byte("written")) },
		"OpenFileForAppending": func() error {
			f, err := s.OpenFileForAppending(name)
			if err == nil {
				f.Close() //nolint:errcheck
			}
			return err
		},
		"OpenFileForWriting": func() error {
			f, err := s.OpenFileForWriting(name)
			if err == nil {
				f.Close() //nolint:errcheck
			}
			return err
		},
		"OpenFile": func() error {
			f, err := s.OpenFile(name)
			if err == nil {
				f.Close() //nolint:errcheck
			}
			return err
		},
		"ReadFile":          func() error { _, err := s.ReadFile(name); return err },
		"EnsurePermissions": func() error { return s.EnsurePermissions(name, 0644) },
		"EnsureOwnership":   func() error { return s.EnsureOwnership(name, "vpn") },
		"EnsurePath":        func() error { return s.EnsurePath(name) },
	}
	for opName, op := range operations {
		err := op()
		if err == nil {
			t.Errorf("%s: expected error for symbolic link", opName)
		} else if !errors.Is(err, ErrSymlink) {
			t.Errorf("%s: expected ErrSymlink, got: %s", opName, err)
		}
		checkTargetUnchanged(t, target)
	}
	// the symlink itself is still there
	if info, err := os.Lstat(filepath.Join(base, name)); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink to be untouched: %v", err)
	}
}

func TestNoFollowStorageRefusesSymlinkedDirectory(t *testing.T) {
	base, target := setupSymlinkDir(t)
	if err := os.Symlink(target, filepath.Join(base, "packetlogs")); err != nil {
		t.Fatal(err)
	}
	s, err := NewNoFollowStorage(base)
	if err != nil {
		t.Fatal(err)
	}
	name := "packetlogs/file"
	operations := map[string]func() error{
		"AppendFile":        func() error { return s.AppendFile(name, []byte("appended")) },
		"WriteFile":         func() error { return s.WriteFile(name, []byte("written")) },
		"ReadFile":          func() error { _, err := s.ReadFile(name); return err },
		"EnsurePermissions": func() error { return s.EnsurePermissions(name, 0644) },
		"EnsureOwnership":   func() error { return s.EnsureOwnership(name, "vpn") },
		"EnsurePathDir":     func() error { return s.EnsurePath("packetlogs") },
		"EnsurePath":        func() error { return s.EnsurePath("packetlogs/newdir") },
		"Remove":            func() error { return s.Remove(name) },
		"Rename":            func() error { return s.Rename(name, "packetlogs/file2") },
		"AppendNewFile":     func() error { return s.AppendFile("packetlogs/new.log", []byte("appended")) },
	}
	for opName, op := range operations {
		err := op()
		if !errors.Is(err, ErrSymlink) {
			t.Errorf("%s: expected ErrSymlink, got: %v", opName, err)
		}
		checkTargetUnchanged(t, target)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected no new files in the directory outside of the base directory, got %d entries", len(entries))
	}
}

func TestNoFollowStorageRefusesHardlinkedFile(t *testing.T) {
	base, target := setupSymlinkDir(t)
	if err := os.Link(filepath.Join(target, "file"), filepath.Join(base, "stats", "link.log")); err != nil {
		t.Skipf("hard links not supported: %s", err)
	}
	s, err := NewNoFollowStorage(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AppendFile("stats/link.log", []byte("appended")); err == nil {
		t.Fatalf("expected error for file with multiple hard links")
	}
	if _, err := s.OpenFileForWriting("stats/link.log"); err == nil {
		t.Fatalf("expected error for file with multiple hard links")
	}
	checkTargetUnchanged(t, target)
}

func TestNoFollowStorageOutsideBase(t *testing.T) {
	base, target := setupSymlinkDir(t)
	s, err := NewNoFollowStorage(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AppendFile("../target/file", []byte("appended")); err == nil {
		t.Fatalf("expected error for path outside of base directory")
	}
	if err := s.WriteFile(filepath.Join(target, "file"), []byte("written")); err == nil {
		t.Fatalf("expected error for absolute path outside of base directory")
	}
	checkTargetUnchanged(t, target)
}

func TestNoFollowStorageOperations(t *testing.T) {
	base, _ := setupSymlinkDir(t)
	s, err := NewNoFollowStorageWithConfigDirs(base)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(base, "config", "clients")); err != nil || !info.IsDir() {
		t.Fatalf("expected config/clients directory: %v", err)
	}
	if err := s.EnsurePath("stats/packetlogs"); err != nil {
		t.Fatalf("EnsurePath error: %s", err)
	}
	if err := s.EnsurePath("stats/packetlogs"); err != nil {
		t.Fatalf("EnsurePath (existing) error: %s", err)
	}
	if err := s.EnsurePermissions("stats/packetlogs", 0770|os.ModeSetgid); err != nil {
		t.Fatalf("EnsurePermissions error: %s", err)
	}
	if err := s.EnsureOwnership("stats/packetlogs", "vpn"); err != nil && os.Geteuid() != 0 {
		t.Fatalf("EnsureOwnership error: %s", err)
	}
	// write (new file, and replacing an existing file keeps the mode)
	if err := s.WriteFile("config/vpn-config.json", []byte("one")); err != nil {
		t.Fatalf("WriteFile error: %s", err)
	}
	if err := os.Chmod(filepath.Join(base, "config/vpn-config.json"), 0640); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteFile(filepath.Join(base, "config/vpn-config.json"), []byte("two")); err != nil {
		t.Fatalf("WriteFile (absolute path) error: %s", err)
	}
	body, err := s.ReadFile("config/vpn-config.json")
	if err != nil || string(body) != "two" {
		t.Fatalf("ReadFile: %q, %v", body, err)
	}
	if info, err := os.Stat(filepath.Join(base, "config/vpn-config.json")); err != nil || info.Mode().Perm() != 0640 {
		t.Fatalf("expected mode to be kept: %v %v", info.Mode(), err)
	}
	// append
	for _, data := range []string{"a", "b"} {
		if err := s.AppendFile("stats/user.log", []byte(data)); err != nil {
			t.Fatalf("AppendFile error: %s", err)
		}
	}
	w, err := s.OpenFileForAppending("stats/user.log")
	if err != nil {
		t.Fatalf("OpenFileForAppending error: %s", err)
	}
	if _, err := w.Write([]byte("c")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := s.OpenFile("stats/user.log")
	if err != nil {
		t.Fatalf("OpenFile error: %s", err)
	}
	body, err = io.ReadAll(r)
	r.Close() //nolint:errcheck
	if err != nil || string(body) != "abc" {
		t.Fatalf("unexpected contents: %q, %v", body, err)
	}
	// overwrite, rename, remove
	w, err = s.OpenFileForWriting("stats/user.log.tmp")
	if err != nil {
		t.Fatalf("OpenFileForWriting error: %s", err)
	}
	if _, err := w.Write([]byte("tmp")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Rename("stats/user.log.tmp", "stats/user.log.gz"); err != nil {
		t.Fatalf("Rename error: %s", err)
	}
	if err := s.Remove("stats/user.log"); err != nil {
		t.Fatalf("Remove error: %s", err)
	}
	if err := s.Remove("stats/packetlogs"); err != nil {
		t.Fatalf("Remove (directory) error: %s", err)
	}
	files, err := s.ReadDir("stats")
	if err != nil || len(files) != 1 || files[0] != "user.log.gz" {
		t.Fatalf("unexpected files: %v, %v", files, err)
	}
	if err := s.Remove("stats/doesnotexist"); !os.IsNotExist(err) {
		t.Fatalf("expected not exist error, got: %v", err)
	}
}
