package config

import "testing"

func TestLoadDefaultConfig(t *testing.T) {
	cfg := Load("/nonexistent/path/config.toml")
	if cfg.TTL != "4h" {
		t.Errorf("expected default ttl '4h', got '%s'", cfg.TTL)
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}
}

func TestVaultDir(t *testing.T) {
	dir := VaultDir()
	if dir == "" {
		t.Error("VaultDir returned empty string")
	}
}
