package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Display struct {
	Color string `toml:"color" comment:"color mode: auto, always, never"`
	Width int    `toml:"width" comment:"max display width (0 = auto-detect)"`
}

type Config struct {
	Repos        []string `toml:"repos" comment:"paths to git repositories to track"`
	RefreshEvery int      `toml:"refresh_every" comment:"auto-refresh after N executions (0 = disabled)"`
	Display      Display  `toml:"display" comment:"display preferences"`
}

func Default() *Config {
	return &Config{
		Repos: []string{},
		Display: Display{
			Color: "auto",
			Width: 0,
		},
	}
}

func DefaultTOML() string {
	cfg := Default()
	b, _ := toml.Marshal(cfg)
	return string(b)
}

func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "gitfetch"), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func DataDir() (string, error) {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine data directory: %w", err)
		}
		dir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dir, "gitfetch"), nil
}

func CachePath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cache.json"), nil
}

func ExpandPath(p string) string {
	if len(p) >= 2 && p[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	for i, r := range cfg.Repos {
		cfg.Repos[i] = ExpandPath(r)
	}
	return &cfg, nil
}

func LoadOrCreate() (*Config, string, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, "", fmt.Errorf("resolving config path: %w", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := Default()
		if createErr := Save(path, cfg); createErr != nil {
			return nil, "", fmt.Errorf("creating default config: %w", createErr)
		}
		return cfg, path, nil
	}

	cfg, err := Load(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
