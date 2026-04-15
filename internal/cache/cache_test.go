package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCacheIsEmpty(t *testing.T) {
	c := New()
	if len(c.Repos) != 0 {
		t.Errorf("expected empty repos, got %d entries", len(c.Repos))
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	now := time.Now().UTC().Truncate(time.Second)
	commitDate := now.Add(-48 * time.Hour)
	tagDate := now.Add(-24 * time.Hour)

	original := New()
	original.Repos["/home/user/project"] = RepoEntry{
		LastCommitDate: &commitDate,
		LastTagDate:    &tagDate,
		ScannedAt:      now,
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(loaded.Repos))
	}
	entry, ok := loaded.Repos["/home/user/project"]
	if !ok {
		t.Fatal("repo entry missing")
	}
	if entry.LastCommitDate == nil {
		t.Fatal("LastCommitDate is nil")
	}
	if !entry.LastCommitDate.Equal(commitDate) {
		t.Errorf("commit date = %v, want %v", entry.LastCommitDate, commitDate)
	}
	if entry.LastTagDate == nil {
		t.Fatal("LastTagDate is nil")
	}
	if !entry.LastTagDate.Equal(tagDate) {
		t.Errorf("tag date = %v, want %v", entry.LastTagDate, tagDate)
	}
	if !entry.ScannedAt.Equal(now) {
		t.Errorf("scanned_at = %v, want %v", entry.ScannedAt, now)
	}
}

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	c, err := Load("/nonexistent/path/cache.json")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(c.Repos) != 0 {
		t.Errorf("expected empty cache, got %d entries", len(c.Repos))
	}
}

func TestLoadEmptyFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(c.Repos) != 0 {
		t.Errorf("expected empty cache, got %d entries", len(c.Repos))
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	if err := os.WriteFile(path, []byte("{not valid json}"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "cache.json")

	c := New()
	if err := Save(path, c); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("cache file not created: %v", err)
	}
}

func TestCacheWithNilDates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	original := New()
	original.Repos["/no/tags/repo"] = RepoEntry{
		LastCommitDate: nil,
		LastTagDate:    nil,
		Error:          "",
		ScannedAt:      time.Now().UTC().Truncate(time.Second),
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	entry := loaded.Repos["/no/tags/repo"]
	if entry.LastCommitDate != nil {
		t.Errorf("expected nil LastCommitDate, got %v", entry.LastCommitDate)
	}
	if entry.LastTagDate != nil {
		t.Errorf("expected nil LastTagDate, got %v", entry.LastTagDate)
	}
}

func TestCacheWithError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	original := New()
	original.Repos["/bad/repo"] = RepoEntry{
		Error:     "not a git repository",
		ScannedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	entry := loaded.Repos["/bad/repo"]
	if entry.Error != "not a git repository" {
		t.Errorf("error = %q, want %q", entry.Error, "not a git repository")
	}
	if entry.LastCommitDate != nil {
		t.Error("LastCommitDate should be nil for errored entry")
	}
}

func TestCacheJSONFormat(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	c := New()
	c.Repos["/repo"] = RepoEntry{
		LastCommitDate: &now,
		LastTagDate:    &now,
		ScannedAt:      now,
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	repos, ok := raw["repos"].(map[string]interface{})
	if !ok {
		t.Fatal("repos is not a map")
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo entry, got %d", len(repos))
	}
}
