package config

import (
	"os"
	"path/filepath"
)

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "secrun")
}

func VaultDir() string { return filepath.Join(ConfigDir(), "vaults") }
