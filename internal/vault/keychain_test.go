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

func cleanupKeychainProject(t *testing.T, project string) {
	t.Helper()
	kb := &KeychainBackend{servicePrefix: "secrun-test"}
	keys, _ := kb.List(project)
	for _, k := range keys { kb.Delete(project, k) }
}

func TestKeychainBackendSetAndGet(t *testing.T) {
	if !hasSecurityCLI() { t.Skip("security CLI not available") }
	kb := &KeychainBackend{servicePrefix: "secrun-test"}
	defer cleanupKeychainProject(t, "kc-test-project")
	err := kb.Set("kc-test-project", "TEST_KEY", "test-value-123")
	if err != nil { t.Fatalf("set failed: %v", err) }
	val, err := kb.Get("kc-test-project", "TEST_KEY")
	if err != nil { t.Fatalf("get failed: %v", err) }
	if val != "test-value-123" { t.Errorf("expected 'test-value-123', got '%s'", val) }
}

func TestKeychainBackendList(t *testing.T) {
	if !hasSecurityCLI() { t.Skip("security CLI not available") }
	kb := &KeychainBackend{servicePrefix: "secrun-test"}
	defer cleanupKeychainProject(t, "kc-test-list")
	kb.Set("kc-test-list", "KEY_A", "val_a")
	kb.Set("kc-test-list", "KEY_B", "val_b")
	keys, err := kb.List("kc-test-list")
	if err != nil { t.Fatalf("list failed: %v", err) }
	if len(keys) != 2 { t.Fatalf("expected 2 keys, got %d", len(keys)) }
}

func TestKeychainBackendDelete(t *testing.T) {
	if !hasSecurityCLI() { t.Skip("security CLI not available") }
	kb := &KeychainBackend{servicePrefix: "secrun-test"}
	defer cleanupKeychainProject(t, "kc-test-del")
	kb.Set("kc-test-del", "DEL_KEY", "del_val")
	err := kb.Delete("kc-test-del", "DEL_KEY")
	if err != nil { t.Fatalf("delete failed: %v", err) }
	_, err = kb.Get("kc-test-del", "DEL_KEY")
	if err == nil { t.Fatal("expected error after delete") }
}
