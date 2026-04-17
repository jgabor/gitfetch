package tui

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
	gitscanner "github.com/jgabor/gitfetch/internal/git"
)

func TestNewModel(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/a", "/b"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg/path", c, "/cache/path")
	if len(m.repos) != 2 {
		t.Errorf("repos = %d, want 2", len(m.repos))
	}
	if m.cfgPath != "/cfg/path" {
		t.Errorf("cfgPath = %q, want /cfg/path", m.cfgPath)
	}
	if m.cachePath != "/cache/path" {
		t.Errorf("cachePath = %q, want /cache/path", m.cachePath)
	}
}

func TestViewEmpty(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if len(view) == 0 {
		t.Error("view should not be empty")
	}
}

func TestViewWithRepos(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/project"}}
	c := cache.New()
	commitDate := time.Now().UTC().Add(-10 * 24 * time.Hour)
	c.Repos["/home/user/project"] = cache.RepoEntry{
		LastCommitDate: &commitDate,
		ScannedAt:      time.Now().UTC(),
	}
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if len(view) == 0 {
		t.Error("view should not be empty")
	}
}

func TestViewShowsHelp(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if len(view) == 0 {
		t.Fatal("view is empty")
	}
}

func TestViewShowsStatus(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.statusMsg = "test status"
	view := m.View()
	if len(view) == 0 {
		t.Error("view should not be empty")
	}
}

func TestRefreshKeyTriggersScan(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/some-repo"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache/cache.json")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd after 'r' keypress")
	}
	um := updated.(model)
	if !um.scanning {
		t.Error("expected scanning state active after 'r' keypress")
	}
	if um.statusMsg != "scanning…" {
		t.Errorf("status = %q, want %q", um.statusMsg, "scanning…")
	}
}

func TestRefreshKeyNoopWhenNoRepos(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache/cache.json")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	um := updated.(model)
	if um.scanning {
		t.Error("scanning should not start when no repos are tracked")
	}
	if um.statusMsg == "scanning…" {
		t.Error("status should not show scanning when no repos are tracked")
	}
}

func TestRefreshKeyIgnoredWhileScanning(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/p"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache/cache.json")
	m.scanning = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd != nil {
		t.Error("expected nil cmd while already scanning (deduplicate)")
	}
}

func TestScanDoneWritesCacheAndClearsScanning(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")

	cfg := &config.Config{Repos: []string{"/home/user/x"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, cachePath)
	m.scanning = true

	now := time.Now().UTC()
	msg := scanDoneMsg{results: []gitscanner.ScanResult{{
		RepoPath:       "/home/user/x",
		LastCommitDate: &now,
		ScannedAt:      now,
	}}}
	updated, _ := m.Update(msg)
	um := updated.(model)
	if um.scanning {
		t.Error("scanning should be false after scanDoneMsg")
	}
	loaded, err := cache.Load(cachePath)
	if err != nil {
		t.Fatalf("cache not persisted: %v", err)
	}
	if _, ok := loaded.Repos["/home/user/x"]; !ok {
		t.Error("scan result not written to cache on disk")
	}
}

func TestViewQuitting(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.quitting = true
	view := m.View()
	if view != "" {
		t.Errorf("quitting view should be empty, got %q", view)
	}
}
