package configmanager

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/in4it/wireguard-server/pkg/wireguard/fsutils"
)

type immutableCall struct {
	filename  string
	immutable bool
	contents  string // contents of the file at the time of the call
}

type immutableStub struct {
	calls     []immutableCall
	flags     map[string]bool
	getErr    error
	setErr    error
	getCalled []string
}

func stubImmutable(t *testing.T) *immutableStub {
	t.Helper()
	stub := &immutableStub{flags: map[string]bool{}}
	origGet, origSet := getImmutable, setImmutable
	t.Cleanup(func() {
		getImmutable, setImmutable = origGet, origSet
		immutableNotSupportedOnce = sync.Once{}
	})
	getImmutable = func(filename string) (bool, error) {
		stub.getCalled = append(stub.getCalled, filename)
		if stub.getErr != nil {
			return false, stub.getErr
		}
		return stub.flags[filename], nil
	}
	setImmutable = func(filename string, immutable bool) error {
		contents, _ := os.ReadFile(filename)
		stub.calls = append(stub.calls, immutableCall{filename: filename, immutable: immutable, contents: string(contents)})
		if stub.setErr != nil {
			return stub.setErr
		}
		stub.flags[filename] = immutable
		return nil
	}
	return stub
}

func TestInstallBinariesClearsAndSetsImmutableFlag(t *testing.T) {
	stubSetCapabilities(t)
	stub := stubImmutable(t)
	dir := t.TempDir()
	binaries := map[string]string{"rest-server": "restserver-test", "configmanager": "configmanager-test", "reset-admin-password": "reset-admin-password-test"}
	for name, downloaded := range binaries {
		if err := os.WriteFile(path.Join(dir, name), []byte("old"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path.Join(dir, downloaded), []byte("new "+name), 0600); err != nil {
			t.Fatal(err)
		}
		stub.flags[path.Join(dir, name)] = true
	}
	if err := installBinaries(dir, binaries); err != nil {
		t.Fatalf("installBinaries error: %s", err)
	}
	if len(stub.calls) != 2*len(binaries) {
		t.Fatalf("expected a clear and a set call per binary, got: %+v", stub.calls)
	}
	for name := range binaries {
		binaryPath := path.Join(dir, name)
		var calls []immutableCall
		for _, call := range stub.calls {
			if call.filename == binaryPath {
				calls = append(calls, call)
			}
		}
		if len(calls) != 2 {
			t.Fatalf("expected 2 calls for %s, got: %+v", name, calls)
		}
		// cleared on the old binary (before the rename), set on the new binary (after the rename)
		if calls[0].immutable || calls[0].contents != "old" {
			t.Fatalf("expected flag to be cleared on the old binary first, got: %+v", calls[0])
		}
		if !calls[1].immutable || calls[1].contents != "new "+name {
			t.Fatalf("expected flag to be set on the new binary, got: %+v", calls[1])
		}
		if !stub.flags[binaryPath] {
			t.Fatalf("expected immutable flag on %s", name)
		}
	}
}

func TestInstallBinariesSetsImmutableFlagOnNewBinary(t *testing.T) {
	stubSetCapabilities(t)
	stub := stubImmutable(t)
	dir := t.TempDir()
	if err := os.WriteFile(path.Join(dir, "reset-admin-password-test"), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installBinaries(dir, map[string]string{"reset-admin-password": "reset-admin-password-test"}); err != nil {
		t.Fatalf("installBinaries error: %s", err)
	}
	if len(stub.calls) != 1 || !stub.calls[0].immutable || stub.calls[0].contents != "new" {
		t.Fatalf("expected only a set call on the new binary, got: %+v", stub.calls)
	}
}

func TestInstallBinariesImmutableFlagNotSupported(t *testing.T) {
	stubSetCapabilities(t)
	stub := stubImmutable(t)
	stub.setErr = fmt.Errorf("%w: test", fsutils.ErrImmutableNotSupported)
	dir := t.TempDir()
	if err := os.WriteFile(path.Join(dir, "configmanager"), []byte("old"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "configmanager-test"), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installBinaries(dir, map[string]string{"configmanager": "configmanager-test"}); err != nil {
		t.Fatalf("installBinaries should continue when the immutable flag is not supported: %s", err)
	}
	if contents, err := os.ReadFile(path.Join(dir, "configmanager")); err != nil || string(contents) != "new" {
		t.Fatalf("binary not installed: %q, %v", contents, err)
	}
}

func TestInstallBinariesClearImmutableFlagError(t *testing.T) {
	stubSetCapabilities(t)
	stub := stubImmutable(t)
	stub.setErr = errors.New("operation not permitted")
	dir := t.TempDir()
	if err := os.WriteFile(path.Join(dir, "configmanager"), []byte("old"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "configmanager-test"), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	err := installBinaries(dir, map[string]string{"configmanager": "configmanager-test"})
	if err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("expected immutable flag error, got: %v", err)
	}
	if contents, err := os.ReadFile(path.Join(dir, "configmanager")); err != nil || string(contents) != "old" {
		t.Fatalf("binary should not be replaced: %q, %v", contents, err)
	}
}

func TestInstallBinariesRefusesSymlink(t *testing.T) {
	stubSetCapabilities(t)
	stubImmutable(t)
	dir := t.TempDir()
	outside := path.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, path.Join(dir, "configmanager-test")); err != nil {
		t.Fatal(err)
	}
	if err := installBinaries(dir, map[string]string{"configmanager": "configmanager-test"}); !errors.Is(err, fsutils.ErrSymlink) {
		t.Fatalf("expected symlink error, got: %v", err)
	}
	info, err := os.Stat(outside)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions of file outside of the directory changed: %s", info.Mode().Perm())
	}
	if _, err := os.Lstat(path.Join(dir, "configmanager")); !os.IsNotExist(err) {
		t.Fatalf("symlink should not be installed as binary")
	}
}

func TestEnsureImmutableBinaries(t *testing.T) {
	stub := stubImmutable(t)
	dir := t.TempDir()
	for _, name := range []string{"rest-server", "configmanager"} { // reset-admin-password doesn't exist
		if err := os.WriteFile(path.Join(dir, name), []byte(name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	stub.flags[path.Join(dir, "configmanager")] = true
	ensureImmutableBinaries(dir, getBinaries())
	if len(stub.calls) != 1 || stub.calls[0].filename != path.Join(dir, "rest-server") || !stub.calls[0].immutable {
		t.Fatalf("expected the immutable flag to be set on rest-server only, got: %+v", stub.calls)
	}
	// second run: nothing to do
	stub.calls = nil
	ensureImmutableBinaries(dir, getBinaries())
	if len(stub.calls) != 0 {
		t.Fatalf("expected no calls, got: %+v", stub.calls)
	}
}

func TestEnsureImmutableBinariesNotSupported(t *testing.T) {
	stub := stubImmutable(t)
	stub.getErr = fmt.Errorf("%w: test", fsutils.ErrImmutableNotSupported)
	dir := t.TempDir()
	for name := range getBinaries() {
		if err := os.WriteFile(path.Join(dir, name), []byte(name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	ensureImmutableBinaries(dir, getBinaries())
	if len(stub.getCalled) != 1 || len(stub.calls) != 0 {
		t.Fatalf("expected to stop after the first unsupported error, got get: %v, set: %+v", stub.getCalled, stub.calls)
	}
}

func TestDownloadFileDoesNotFollowSymlink(t *testing.T) {
	ts := newTestFileServer(t, "binary contents")
	dir := t.TempDir()
	outside := path.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	dest := path.Join(dir, "binary")
	if err := os.Symlink(outside, dest); err != nil {
		t.Fatal(err)
	}
	if err := downloadFile(ts, dest); err != nil {
		t.Fatalf("downloadFile error: %s", err)
	}
	if contents, err := os.ReadFile(outside); err != nil || string(contents) != "outside" {
		t.Fatalf("file outside of the directory was modified: %q, %v", contents, err)
	}
	info, err := os.Lstat(dest)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("expected a regular file: %v", err)
	}
	if contents, err := os.ReadFile(dest); err != nil || string(contents) != "binary contents" {
		t.Fatalf("unexpected download: %q, %v", contents, err)
	}
}
