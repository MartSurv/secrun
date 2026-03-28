//go:build darwin

package vault

import (
	"os/exec"
	"strings"
)

const keychainService = "secrun"

// KeychainSavePassword stores a project's master password in macOS Keychain.
func KeychainSavePassword(project, password string) error {
	// Delete existing entry first (security CLI errors on duplicate)
	exec.Command("security", "delete-generic-password",
		"-s", keychainService, "-a", project).Run()
	// Password passed as -w argument. On macOS /proc doesn't exist, and the
	// process runs for milliseconds, so cmdline exposure risk is minimal.
	cmd := exec.Command("security", "add-generic-password",
		"-s", keychainService, "-a", project, "-w", password)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// KeychainLoadPassword retrieves a project's master password from macOS Keychain.
// Returns empty string and error if not found.
func KeychainLoadPassword(project string) (string, error) {
	out, err := exec.Command("security", "find-generic-password",
		"-s", keychainService, "-a", project, "-w").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// KeychainHasPassword checks if a password is stored in Keychain for the project.
func KeychainHasPassword(project string) bool {
	_, err := KeychainLoadPassword(project)
	return err == nil
}

// KeychainDeletePassword removes a project's master password from macOS Keychain.
func KeychainDeletePassword(project string) error {
	return exec.Command("security", "delete-generic-password",
		"-s", keychainService, "-a", project).Run()
}
