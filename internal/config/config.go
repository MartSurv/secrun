package config

import "os"

type Config struct {
	TTL string
}

func Load(path string) *Config {
	return &Config{TTL: "4h"}
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.config/secrun"
}

func ConfigPath() string { return ConfigDir() + "/config.toml" }
func VaultDir() string   { return ConfigDir() + "/vaults" }
