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
			return fmt.Errorf("start vpn exit Status: %d", exiterr.ExitCode())
		} else {
			return fmt.Errorf("error while waiting for the VPN to start: %v", err)
		}
	}
	return nil
}

func StopVPN() error {
	cmd := exec.Command("wg-quick", "down", "vpn")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("VPN stop error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("stop vpn exit Status: %d", exiterr.ExitCode())
		} else {
			return fmt.Errorf("error while waiting for the VPN to stop: %v", err)
		}
	}
	return nil
}
