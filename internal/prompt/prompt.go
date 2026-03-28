package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Password prompts the user for a password without echoing input.
func Password(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(pw), nil
}

// PasswordConfirm prompts for a password twice and checks they match.
func PasswordConfirm(label string) (string, error) {
	pw1, err := Password(label)
	if err != nil { return "", err }
	pw2, err := Password("Confirm password")
	if err != nil { return "", err }
	if pw1 != pw2 { return "", fmt.Errorf("passwords do not match") }
	if len(pw1) < 8 { return "", fmt.Errorf("password must be at least 8 characters") }
	return pw1, nil
}

// Value prompts the user for a value, showing it as they type.
func Value(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	reader := bufio.NewReader(os.Stdin)
	val, err := reader.ReadString('\n')
	if err != nil { return "", fmt.Errorf("read value: %w", err) }
	return strings.TrimSpace(val), nil
}
