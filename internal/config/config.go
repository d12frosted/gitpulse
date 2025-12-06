package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Repos []string `toml:"repos"`
	Theme string   `toml:"theme,omitempty"`
}

type RepoConfig struct {
	Path string
	Name string
}

func (c *Config) RepoConfigs() []RepoConfig {
	configs := make([]RepoConfig, 0, len(c.Repos))
	for _, path := range c.Repos {
		expanded := expandPath(path)
		name := filepath.Base(expanded)
		configs = append(configs, RepoConfig{
			Path: expanded,
			Name: name,
		})
	}
	return configs
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "gitpulse")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "gitpulse")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &ConfigNotFoundError{Path: path}
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	f, err := os.Create(ConfigPath())
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func ExampleConfig() string {
	return `# gitpulse configuration

# Color theme: dracula, nord, catppuccin, gruvbox, tokyonight, mono, jrpg-dark, jrpg-light
theme = "dracula"

# Repository paths to monitor
repos = [
    "~/Developer/project1",
    "~/Developer/project2",
    "~/work/important-repo",
]
`
}

type ConfigNotFoundError struct {
	Path string
}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("config not found at %s", e.Path)
}
