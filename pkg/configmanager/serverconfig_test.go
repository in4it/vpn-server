package configmanager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/in4it/go-devops-platform/storage"
	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func TestWriteServerConfigKeepsExistingConfig(t *testing.T) {
	dir := t.TempDir()
	orig := serverConfigPath
	t.Cleanup(func() { serverConfigPath = orig })
	serverConfigPath = filepath.Join(dir, "vpn.conf")
	notAllowed := func(storage.Iface) error {
		return fmt.Errorf("could not generate wireguard server config: %w", wireguard.ErrServerConfigNotAllowed)
	}
	otherError := func(storage.Iface) error { return fmt.Errorf("failed to get vpn config") }
	ok := func(storage.Iface) error { return nil }
	s := &memorystorage.MockMemoryStorage{}

	// no existing config: the error is returned
	if err := writeServerConfig(s, notAllowed); err == nil {
		t.Fatalf("expected error without existing config")
	}
	if err := os.WriteFile(serverConfigPath, []byte("[Interface]\n"), 0600); err != nil {
		t.Fatalf("write error: %s", err)
	}
	// existing config: the vpn starts with the existing config
	if err := writeServerConfig(s, notAllowed); err != nil {
		t.Fatalf("expected existing config to be used, got: %s", err)
	}
	// other errors are still returned
	if err := writeServerConfig(s, otherError); err == nil {
		t.Fatalf("expected other errors to be returned")
	}
	if err := writeServerConfig(s, ok); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	// a symlink isn't used as existing config
	link := filepath.Join(dir, "link.conf")
	if err := os.Symlink(serverConfigPath, link); err != nil {
		t.Fatalf("symlink error: %s", err)
	}
	serverConfigPath = link
	if err := writeServerConfig(s, notAllowed); err == nil {
		t.Fatalf("expected error when the existing config is a symlink")
	}
}

// TestGenerateServerConfigNotAllowedError: a template with custom commands
// returns ErrServerConfigNotAllowed, which writeServerConfig uses to keep the
// existing config.
func TestGenerateServerConfigNotAllowedError(t *testing.T) {
	s := &memorystorage.MockMemoryStorage{}
	if _, err := wireguard.CreateNewVPNConfig(s); err != nil {
		t.Fatalf("CreateNewVPNConfig error: %s", err)
	}
	template, err := wireguard.GetServerTemplate(s)
	if err != nil {
		t.Fatalf("get server template: %s", err)
	}
	if err := wireguard.WriteServerTemplate(s, append(template, []byte("\nPostUp = touch /tmp/x\n")...)); err != nil {
		t.Fatalf("write server template: %s", err)
	}
	err = wireguard.CheckWireGuardServerConfig(s)
	if err == nil {
		t.Fatalf("expected error for custom PostUp")
	}
	if !errors.Is(err, wireguard.ErrServerConfigNotAllowed) {
		t.Fatalf("expected ErrServerConfigNotAllowed, got: %s", err)
	}
}
