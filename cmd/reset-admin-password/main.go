package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/in4it/wireguard-server/pkg/commands"
	"golang.org/x/term"
)

func main() {
	var (
		appDir              string
		err                 error
		newAdminUserCreated bool
	)
	flag.StringVar(&appDir, "vpn-dir", "/vpn", "directory where vpn files are located")
	flag.Parse()

	password, _ := getPassword()
	if newAdminUserCreated, err = commands.ResetPassword(appDir, password); err != nil {
		fmt.Printf("Failed to changed admin password: %s", err)
		os.Exit(1)
	}
	if !newAdminUserCreated {
		resetMFA, err := getLine("Also remove MFA if present? [Y/n] ")
		if err != nil {
			fmt.Printf("Failed to changed admin password: %s", err)
			os.Exit(1)
		}
		if strings.TrimSpace(strings.ToUpper(resetMFA)) == "" || strings.TrimSpace(strings.ToUpper(resetMFA)) == "Y" {
			err = commands.ResetAdminMFA(appDir)
			if err != nil {
				fmt.Printf("Failed to reset admin MFA: %s", err)
				os.Exit(1)
			}
			fmt.Printf("Admin MFA removed.\n")
		}
	}
	if err := reloadConfig(appDir); err != nil {
		fmt.Printf("Admin password reset, but could not restart rest-server: %s.\nTry to restart rest-server manually with systemctl restart vpn-rest-server\n", err)
		os.Exit(1)
	}
	fmt.Printf("Admin password succesfully changed!\n")
}

func reloadConfig(appDir string) error {
	// read pid file
	pid, err := os.ReadFile(path.Join(appDir, "rest-server.pid"))
	if err != nil {
		return fmt.Errorf("could not read pidfile: %s", err)
	}
	pidInt, err := strconv.Atoi(string(pid))
	if err != nil {
		return fmt.Errorf("pid is not a number: %s", err)
	}
	err = syscall.Kill(pidInt, syscall.SIGHUP)
	if err != nil {
		return fmt.Errorf("sighup rest-server failed: %s", err)
	}
	return nil
}

func getPassword() (string, error) {
	fmt.Print("Enter new password for the admin user: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}
func getLine(question string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", question)
	return reader.ReadString('\n')
}
