//go:build linux
// +build linux

package processpeerconfig

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/in4it/go-devops-platform/storage"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func wgAddPeer(storage storage.Iface, peerConfig wireguard.PeerConfig) error {
	// wg set

	args := []string{"set", wireguard.VPN_INTERFACE_NAME, "peer", peerConfig.PublicKey, "allowed-ips", strings.Join(peerConfig.ServerAllowedIPs, ","), "preshared-key", path.Join(storage.GetPath(), wireguard.VPN_SERVER_SECRETS_PATH, wireguard.PRESHARED_KEY_FILENAME)}

	fmt.Printf("Executing cmd: wg %s\n", strings.Join(args, " "))

	cmd := exec.Command("wg", args...)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("wg set error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("wg set exit Status: %d", exiterr.ExitCode())
		} else {
			return fmt.Errorf("error during wg set: %v", err)
		}
	}

	// add route (needs another check first?)
	/*for _, allowedIP := range peerConfig.AllowedIPs {
		cmd := exec.Command("ip", "-4", "route", "add", allowedIP, "dev", wireguard.VPN_INTERFACE_NAME)

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("wg set error: %v", err)
		}

		if err := cmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				return fmt.Errorf("wg set exit Status: %d", exiterr.ExitCode())
			} else {
				return fmt.Errorf("error during wg set: %v", err)
			}
		}
	}*/
	return nil
}

func wgDeletePeer(peerConfig wireguard.PeerConfig) error {
	// wg set to delete a peer
	args := []string{"set", wireguard.VPN_INTERFACE_NAME, "peer", peerConfig.PublicKey, "remove"}

	fmt.Printf("Executing cmd: wg %s\n", strings.Join(args, " "))

	cmd := exec.Command("wg", args...)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("wg set error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("wg set exit Status: %d", exiterr.ExitCode())
		} else {
			return fmt.Errorf("error during wg set: %v", err)
		}
	}
	return nil
}

func wgSetServerPrivateKey(privateKeyPath string) error {
	// wg set to change private key
	args := []string{"set", wireguard.VPN_INTERFACE_NAME, "private-key", privateKeyPath}

	fmt.Printf("Executing cmd: wg %s\n", strings.Join(args, " "))

	cmd := exec.Command("wg", args...)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("wg set private-key error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("wg set exit Status: %d", exiterr.ExitCode())
		} else {
			return fmt.Errorf("error during wg set private-key: %v", err)
		}
	}
	return nil
}
