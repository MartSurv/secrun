//go:build darwin

package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/MartSurv/secrun/internal/prompt"
	"github.com/MartSurv/secrun/internal/vault"
)

// offerKeychainSave prompts the user to save the master password to macOS Keychain.
// Called after successful init with the password already confirmed.
func offerKeychainSave(project string) {
	fmt.Fprintf(os.Stderr, "Save password to macOS Keychain? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "" && answer != "y" && answer != "yes" {
		return
	}

	pw, err := prompt.Password("Re-enter master password to save to Keychain")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read password: %v\n", err)
		return
	}
	if err := vault.KeychainSavePassword(project, pw); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save to Keychain: %v\n", err)
		return
	}
	fmt.Fprintln(os.Stderr, "Password saved to Keychain. You won't be prompted again for this project.")
}

// keychainPasswordFn returns a PasswordFunc that tries Keychain first, falls back to prompt.
func keychainPasswordFn(project string) vault.PasswordFunc {
	var cached string
	return func() (string, error) {
		if cached != "" {
			return cached, nil
		}
		// Try keychain first
		if pw, err := vault.KeychainLoadPassword(project); err == nil && pw != "" {
			cached = pw
			return cached, nil
		}
		// Fall back to interactive prompt
		pw, err := prompt.Password("Master password")
		if err != nil {
			return "", err
		}
		cached = pw
		return cached, nil
	}
}
