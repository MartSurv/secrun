package config

import (
	"os"
	"github.com/BurntSushi/toml"
)

type Defaults struct {
	Store string `toml:"store"`
	TTL   string `toml:"ttl"`
}

type ProjectConfig struct {
	Store string `toml:"store"`
}

type Config struct {
	Defaults Defaults                 `toml:"defaults"`
	Projects map[string]ProjectConfig `toml:"projects"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Defaults: Defaults{Store: "file", TTL: "4h"},
		Projects: make(map[string]ProjectConfig),
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) { return cfg, nil }
		return nil, err
	}
	if err := toml.Unmarshal(data, cfg); err != nil { return nil, err }
	if cfg.Defaults.Store == "" { cfg.Defaults.Store = "file" }
	if cfg.Defaults.TTL == "" { cfg.Defaults.TTL = "4h" }
	if cfg.Projects == nil { cfg.Projects = make(map[string]ProjectConfig) }
	return cfg, nil
}

func (c *Config) StoreForProject(project, flagOverride string) string {
	if flagOverride != "" { return flagOverride }
	if pc, ok := c.Projects[project]; ok && pc.Store != "" { return pc.Store }
	return c.Defaults.Store
}

func (c *Config) TTLForProject(project, flagOverride string) string {
	if flagOverride != "" { return flagOverride }
	return c.Defaults.TTL
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.config/secrun"
}

func ConfigPath() string { return ConfigDir() + "/config.toml" }
func VaultDir() string { return ConfigDir() + "/vaults" }
