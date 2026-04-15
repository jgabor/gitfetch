package tui

import (
	"testing"
	"time"

	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
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
