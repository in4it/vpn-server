package configmanager

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"slices"
	"testing"
	"time"
)

func stubSetCapabilities(t *testing.T) *[]string {
	t.Helper()
	called := &[]string{}
	orig := setCapabilities
	t.Cleanup(func() { setCapabilities = orig })
	setCapabilities = func(filename string) error {
		*called = append(*called, filename)
		return nil
	}
	return called
}

func TestInstallBinariesReplacesExisting(t *testing.T) {
	capabilities := stubSetCapabilities(t)
	dir := t.TempDir()
	binaries := map[string]string{"rest-server": "restserver-test", "configmanager": "configmanager-test"}
	for name, downloaded := range binaries {
		if err := os.WriteFile(path.Join(dir, name), []byte("old"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path.Join(dir, downloaded), []byte("new "+name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := installBinaries(dir, binaries); err != nil {
		t.Fatalf("installBinaries error: %s", err)
	}
	for name, downloaded := range binaries {
		contents, err := os.ReadFile(path.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(contents) != "new "+name {
			t.Fatalf("binary %s not replaced: %q", name, contents)
		}
		info, err := os.Stat(path.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0700 {
			t.Fatalf("expected 0700 permissions on %s, got %s", name, info.Mode().Perm())
		}
		if _, err := os.Stat(path.Join(dir, downloaded)); !os.IsNotExist(err) {
			t.Fatalf("downloaded file %s should be renamed", downloaded)
		}
	}
	if runtime.GOOS == "linux" {
		// setcap is set on the downloaded file, before it replaces the old binary
		if len(*capabilities) != 1 || (*capabilities)[0] != path.Join(dir, "restserver-test") {
			t.Fatalf("unexpected setcap calls: %v", *capabilities)
		}
	}
}

// TestInstallBinariesWithoutExistingBinary: the old Remove+Rename failed when a
// binary didn't exist yet (e.g. a previous upgrade was interrupted).
func TestInstallBinariesWithoutExistingBinary(t *testing.T) {
	stubSetCapabilities(t)
	dir := t.TempDir()
	if err := os.WriteFile(path.Join(dir, "reset-admin-password-test"), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installBinaries(dir, map[string]string{"reset-admin-password": "reset-admin-password-test"}); err != nil {
		t.Fatalf("installBinaries error: %s", err)
	}
	if contents, err := os.ReadFile(path.Join(dir, "reset-admin-password")); err != nil || string(contents) != "new" {
		t.Fatalf("binary not installed: %q, %v", contents, err)
	}
}

// TestRestartServicesCommand: both services are restarted with one systemctl
// call, with configmanager last (systemctl runs in configmanager's cgroup).
func TestRestartServicesCommand(t *testing.T) {
	args := restartServicesCommand()
	if args[0] != "systemctl" || !slices.Contains(args, "restart") {
		t.Fatalf("unexpected command: %v", args)
	}
	restServer := slices.Index(args, "vpn-rest-server")
	configManager := slices.Index(args, "vpn-configmanager")
	if restServer == -1 || configManager == -1 || configManager != len(args)-1 || restServer > configManager {
		t.Fatalf("expected vpn-rest-server before vpn-configmanager (last): %v", args)
	}
}

func TestDownloadClientTimeouts(t *testing.T) {
	client := newDownloadClient()
	if client.Timeout < time.Minute {
		t.Fatalf("download deadline too short for the binaries: %s", client.Timeout)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport")
	}
	if transport.ResponseHeaderTimeout == 0 || transport.ResponseHeaderTimeout > time.Minute || transport.TLSHandshakeTimeout == 0 {
		t.Fatalf("expected short connect timeouts: %+v", transport)
	}
}

// TestDownloadFileSlowBody: a download that takes longer than the connect
// timeout must succeed, as long as it's within the download deadline.
func TestDownloadFileSlowBody(t *testing.T) {
	origDownload, origConnect := upgradeDownloadTimeout, upgradeConnectTimeout
	t.Cleanup(func() { upgradeDownloadTimeout, upgradeConnectTimeout = origDownload, origConnect })
	upgradeDownloadTimeout = 5 * time.Second
	upgradeConnectTimeout = 200 * time.Millisecond

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		for i := 0; i < 5; i++ {
			_, _ = w.Write([]byte("chunk"))
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		}
	}))
	defer ts.Close()

	dest := path.Join(t.TempDir(), "binary")
	if err := downloadFile(ts.URL, dest); err != nil {
		t.Fatalf("downloadFile error: %s", err)
	}
	contents, err := os.ReadFile(dest)
	if err != nil || string(contents) != "chunkchunkchunkchunkchunk" {
		t.Fatalf("unexpected download: %q, %v", contents, err)
	}
}

// TestUpgradeConcurrentPost: a second upgrade request while an upgrade runs is refused
func TestUpgradeConcurrentPost(t *testing.T) {
	origRunUpgrade := runUpgrade
	t.Cleanup(func() { runUpgrade = origRunUpgrade })
	release := make(chan struct{})
	started := make(chan struct{}, 2)
	runUpgrade = func() error {
		started <- struct{}{}
		<-release
		return nil
	}
	c := &ConfigManager{}
	post := func() int {
		w := httptest.NewRecorder()
		c.upgrade(w, httptest.NewRequest(http.MethodPost, "http://127.0.0.1/upgrade", nil))
		return w.Result().StatusCode
	}
	if status := post(); status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	<-started
	if status := post(); status != http.StatusConflict {
		t.Fatalf("expected 409 for concurrent upgrade, got %d", status)
	}
	close(release)
	// after the upgrade finished, a new upgrade can start
	deadline := time.Now().Add(2 * time.Second)
	for {
		if upgradeMutex.TryLock() {
			upgradeMutex.Unlock()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("upgrade mutex not released")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(started) != 0 {
		t.Fatalf("upgrade ran more than once")
	}
}
