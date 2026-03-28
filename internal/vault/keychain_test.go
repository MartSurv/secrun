//go:build darwin

package vault

import (
	"os/exec"
	"testing"
)

func hasSecurityCLI() bool {
	_, err := exec.LookPath("security")
	return err == nil
}

func TestKeychainSaveAndLoadPassword(t *testing.T) {
	if !hasSecurityCLI() {
		t.Skip("security CLI not available")
	}
	defer KeychainDeletePassword("test-kc-project")

	err := KeychainSavePassword("test-kc-project", "my-secret-pass")
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	pw, err := KeychainLoadPassword("test-kc-project")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if pw != "my-secret-pass" {
		t.Errorf("expected 'my-secret-pass', got '%s'", pw)
	}
}

func TestKeychainHasPassword(t *testing.T) {
	if !hasSecurityCLI() {
		t.Skip("security CLI not available")
	}
	defer KeychainDeletePassword("test-kc-has")

	if KeychainHasPassword("test-kc-has") {
		t.Fatal("should not have password before save")
	}

	KeychainSavePassword("test-kc-has", "pass123")

	if !KeychainHasPassword("test-kc-has") {
		t.Fatal("should have password after save")
	}
}

func TestKeychainDeletePassword(t *testing.T) {
	if !hasSecurityCLI() {
		t.Skip("security CLI not available")
	}

	KeychainSavePassword("test-kc-del", "pass123")
	KeychainDeletePassword("test-kc-del")

	if KeychainHasPassword("test-kc-del") {
		t.Fatal("should not have password after delete")
	}
}
