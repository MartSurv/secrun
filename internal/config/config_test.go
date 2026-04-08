package config

import (
	"strings"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}
	if !strings.Contains(dir, ".config/secrun") {
		t.Errorf("ConfigDir should contain '.config/secrun', got '%s'", dir)
	}
}

func TestVaultDir(t *testing.T) {
	dir := VaultDir()
	if dir == "" {
		t.Error("VaultDir returned empty string")
	}
	if !strings.HasSuffix(dir, "vaults") {
		t.Errorf("VaultDir should end with 'vaults', got '%s'", dir)
	}
}
