package main

import (
	"fmt"
	"strings"
)

const MIN_PASSWORD_LENGTH = 8

// readNewPassword asks for the new password twice and validates it. An error
// reading the password (e.g. no terminal) must never result in an empty password.
func readNewPassword(readPassword func(prompt string) (string, error)) (string, error) {
	password, err := readPassword("Enter new password for the admin user: ")
	if err != nil {
		return "", fmt.Errorf("could not read password (run this command in an interactive terminal): %s", err)
	}
	if err := validatePassword(password); err != nil {
		return "", err
	}
	confirmation, err := readPassword("\nRepeat new password: ")
	if err != nil {
		return "", fmt.Errorf("could not read password (run this command in an interactive terminal): %s", err)
	}
	if password != confirmation {
		return "", fmt.Errorf("passwords don't match")
	}
	return password, nil
}

func validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password can't be empty")
	}
	if len(password) < MIN_PASSWORD_LENGTH {
		return fmt.Errorf("password needs to be at least %d characters", MIN_PASSWORD_LENGTH)
	}
	return nil
}
