package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if cfg.Defaults.Store != "file" { t.Errorf("expected default store 'file', got '%s'", cfg.Defaults.Store) }
	if cfg.Defaults.TTL != "4h" { t.Errorf("expected default ttl '4h', got '%s'", cfg.Defaults.TTL) }
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[defaults]
store = "keychain"
ttl = "2h"

[projects.page-probe]
store = "keychain"

[projects.backend]
store = "file"
`
	os.WriteFile(path, []byte(content), 0600)
	cfg, err := Load(path)
	if err != nil { t.Fatalf("load failed: %v", err) }
	if cfg.Defaults.Store != "keychain" { t.Errorf("expected 'keychain', got '%s'", cfg.Defaults.Store) }
	if cfg.Defaults.TTL != "2h" { t.Errorf("expected '2h', got '%s'", cfg.Defaults.TTL) }
	if cfg.Projects["page-probe"].Store != "keychain" { t.Errorf("expected page-probe store 'keychain'") }
	if cfg.Projects["backend"].Store != "file" { t.Errorf("expected backend store 'file'") }
}

func TestStoreForProject(t *testing.T) {
	cfg := &Config{
		Defaults: Defaults{Store: "file", TTL: "4h"},
		Projects: map[string]ProjectConfig{"page-probe": {Store: "keychain"}},
	}
	if s := cfg.StoreForProject("page-probe", ""); s != "keychain" { t.Errorf("expected 'keychain', got '%s'", s) }
	if s := cfg.StoreForProject("other", ""); s != "file" { t.Errorf("expected 'file', got '%s'", s) }
	if s := cfg.StoreForProject("page-probe", "file"); s != "file" { t.Errorf("flag override: expected 'file', got '%s'", s) }
}

func TestTTLForProject(t *testing.T) {
	cfg := &Config{Defaults: Defaults{Store: "file", TTL: "4h"}}
	if ttl := cfg.TTLForProject("any", ""); ttl != "4h" { t.Errorf("expected '4h', got '%s'", ttl) }
	if ttl := cfg.TTLForProject("any", "1h"); ttl != "1h" { t.Errorf("flag override: expected '1h', got '%s'", ttl) }
}
