package wireguard

import (
	"fmt"
	"os/exec"
)

func StartVPN() error {
	cmd := exec.Command("wg-quick", "up", "vpn")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("VPN start error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("exit Status: %d", exiterr.ExitCode())
		} else {
			return fmt.Errorf("error while waiting for the VPN to start: %v", err)
		}
	}
	return nil
}
