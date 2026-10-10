package wireguard

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/in4it/wireguard-server/pkg/wireguard/fsutils"
)

func newNoFollowTestStorage(t *testing.T) (*fsutils.NoFollowStorage, string, string) {
	t.Helper()
	base := t.TempDir()
	outside := path.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	storage, err := fsutils.NewNoFollowStorage(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{VPN_STATS_DIR, path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR)} {
		if err := storage.EnsurePath(dir); err != nil {
			t.Fatal(err)
		}
	}
	return storage, base, outside
}

func checkOutsideFileUnchanged(t *testing.T, outside string) {
	t.Helper()
	contents, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "outside" {
		t.Fatalf("file outside of the storage directory was modified: %q", contents)
	}
	info, err := os.Stat(outside)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions of file outside of the storage directory changed: %s", info.Mode().Perm())
	}
}

func TestParsePacketDoesNotFollowSymlinks(t *testing.T) {
	storage, base, outside := newNoFollowTestStorage(t)
	now := time.Now()
	clientCache := &ClientCache{
		Addresses: []ClientCacheAddresses{
			{
				Address:  net.IPNet{IP: net.ParseIP("10.189.184.2"), Mask: net.IPMask(net.ParseIP("255.255.255.255").To4())},
				ClientID: "1-2-3-4",
			},
		},
	}
	logFile := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR, "1-2-3-4-"+now.Format("2006-01-02")+".log")
	if err := os.Symlink(outside, path.Join(base, logFile)); err != nil {
		t.Fatal(err)
	}
	// dns request from 10.189.184.2
	data, err := hex.DecodeString("45000037e04900004011cdab0abdb8020a000002e60d00350023d6861e1501000001000000000000056170706c6503636f6d0000010001")
	if err != nil {
		t.Fatal(err)
	}
	openFiles := make(PacketLoggerOpenFiles)
	err = parsePacket(storage, data, clientCache, openFiles, map[string]bool{"dns": true}, now)
	if !errors.Is(err, fsutils.ErrSymlink) {
		t.Fatalf("expected symlink error, got: %v", err)
	}
	if len(openFiles) != 0 {
		t.Fatalf("expected no open files")
	}
	checkOutsideFileUnchanged(t, outside)

	// a regular file is written
	if err := os.Remove(path.Join(base, logFile)); err != nil {
		t.Fatal(err)
	}
	if err := parsePacket(storage, data, clientCache, openFiles, map[string]bool{"dns": true}, now); err != nil {
		t.Fatalf("parsePacket error: %s", err)
	}
	closeOpenFiles(openFiles)
	body, err := storage.ReadFile(logFile)
	if err != nil || len(body) == 0 {
		t.Fatalf("expected packet log to be written: %q, %v", body, err)
	}
}

func TestPacketLoggerLogRotationDoesNotFollowSymlinks(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	prefix := path.Join(VPN_STATS_DIR, VPN_PACKETLOGGER_DIR)

	// the log file to compress is a symlink
	storage, base, outside := newNoFollowTestStorage(t)
	logFile := path.Join(prefix, fmt.Sprintf("1-2-3-4-%s.log", yesterday))
	if err := os.Symlink(outside, path.Join(base, logFile)); err != nil {
		t.Fatal(err)
	}
	if err := packetLoggerLogRotation(storage); !errors.Is(err, fsutils.ErrSymlink) {
		t.Fatalf("expected symlink error, got: %v", err)
	}
	if storage.FileExists(logFile + ".gz") {
		t.Fatalf("symlinked log file should not be compressed")
	}
	checkOutsideFileUnchanged(t, outside)

	// the temporary file is a symlink
	storage, base, outside = newNoFollowTestStorage(t)
	if err := storage.WriteFile(logFile, []byte("log line")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, path.Join(base, logFile+".gz.tmp")); err != nil {
		t.Fatal(err)
	}
	if err := packetLoggerLogRotation(storage); !errors.Is(err, fsutils.ErrSymlink) {
		t.Fatalf("expected symlink error, got: %v", err)
	}
	checkOutsideFileUnchanged(t, outside)

	// the packet log directory is a symlink
	storage, base, outside = newNoFollowTestStorage(t)
	if err := os.RemoveAll(path.Join(base, prefix)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(path.Dir(outside), path.Join(base, prefix)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(path.Dir(outside), fmt.Sprintf("1-2-3-4-%s.log", yesterday)), []byte("log line"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := packetLoggerLogRotation(storage); !errors.Is(err, fsutils.ErrSymlink) {
		t.Fatalf("expected symlink error, got: %v", err)
	}
	if _, err := os.Stat(path.Join(path.Dir(outside), fmt.Sprintf("1-2-3-4-%s.log.gz.tmp", yesterday))); !os.IsNotExist(err) {
		t.Fatalf("no files should be created outside of the storage directory")
	}
	checkOutsideFileUnchanged(t, outside)
}
