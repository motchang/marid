package config

import (
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// PromptForPassword displays a password prompt and reads securely from stdin
func PromptForPassword() (string, error) {
	fmt.Print("Enter MySQL password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after password input

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return strings.TrimSpace(string(passwordBytes)), nil
}
