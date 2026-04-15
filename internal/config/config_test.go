package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	content := `
repos = ["/home/user/projects", "/home/user/work/api"]

[display]
color = "always"
width = 120
`
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0] != "/home/user/projects" {
		t.Errorf("repos[0] = %q", cfg.Repos[0])
	}
	if cfg.Display.Color != "always" {
		t.Errorf("display.color = %q", cfg.Display.Color)
	}
	if cfg.Display.Width != 120 {
		t.Errorf("display.width = %d", cfg.Display.Width)
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	content := `this is not [valid toml {{{`
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultReturnsSaneValues(t *testing.T) {
	cfg := Default()
	if len(cfg.Repos) == 0 {
		t.Error("default config should have at least one repo path")
	}
	if cfg.Display.Color != "auto" {
		t.Errorf("default color = %q, want auto", cfg.Display.Color)
	}
	if cfg.Display.Width != 0 {
		t.Errorf("default width = %d, want 0", cfg.Display.Width)
	}
}

func TestDefaultTOMLIsParseable(t *testing.T) {
	raw := DefaultTOML()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("DefaultTOML() produced unparseable output: %v", err)
	}
	if len(cfg.Repos) == 0 {
		t.Error("parsed default should have repos")
	}
}

func TestConfigDirResolves(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}
	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}
	if filepath.Base(dir) != "gitfetch" {
		t.Errorf("ConfigDir() = %q, want base gitfetch", dir)
	}
}

func TestDataDirRespectsXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdgtest/data")
	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir() error: %v", err)
	}
	expected := filepath.Join("/tmp/xdgtest/data", "gitfetch")
	if dir != expected {
		t.Errorf("DataDir() = %q, want %q", dir, expected)
	}
}

func TestDataDirFallsBackToHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", "/tmp/fakehome")
	defer t.Setenv("HOME", origHome)

	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir() error: %v", err)
	}
	expected := filepath.Join("/tmp/fakehome", ".local", "share", "gitfetch")
	if dir != expected {
		t.Errorf("DataDir() = %q, want %q", dir, expected)
	}
}

func TestCachePathUnderDataDir(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdgtest/data")
	path, err := CachePath()
	if err != nil {
		t.Fatalf("CachePath() error: %v", err)
	}
	expected := filepath.Join("/tmp/xdgtest/data", "gitfetch", "cache.json")
	if path != expected {
		t.Errorf("CachePath() = %q, want %q", path, expected)
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.toml")

	original := &Config{
		Repos: []string{"/a", "/b"},
		Display: Display{
			Color: "never",
			Width: 80,
		},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.Repos) != 2 {
		t.Errorf("repos length = %d, want 2", len(loaded.Repos))
	}
	if loaded.Display.Color != "never" {
		t.Errorf("color = %q, want never", loaded.Display.Color)
	}
	if loaded.Display.Width != 80 {
		t.Errorf("width = %d, want 80", loaded.Display.Width)
	}
}

func TestLoadOrCreateCreatesDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, path, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("config file not created at %s: %v", path, statErr)
	}
}

func TestLoadOrCreateLoadsExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "gitfetch")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `repos = ["/custom/path"]` + "\n"
	path := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	cfg, gotPath, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate() error: %v", err)
	}
	if gotPath != path {
		t.Errorf("path = %q, want %q", gotPath, path)
	}
	if len(cfg.Repos) != 1 || cfg.Repos[0] != "/custom/path" {
		t.Errorf("repos = %v, want [/custom/path]", cfg.Repos)
	}
}
