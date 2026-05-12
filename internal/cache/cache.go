package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RepoEntry struct {
	LastCommitDate *time.Time `json:"last_commit_date,omitempty"`
	LastTagDate    *time.Time `json:"last_tag_date,omitempty"`
	LastTag        string     `json:"last_tag,omitempty"`
	Error          string     `json:"error,omitempty"`
	ScannedAt      time.Time  `json:"scanned_at"`
}

type Cache struct {
	Repos        map[string]RepoEntry `json:"repos"`
	RefreshCount int                  `json:"refresh_count"`
}

func New() *Cache {
	return &Cache{
		Repos: make(map[string]RepoEntry),
	}
}

func Load(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, fmt.Errorf("reading cache: %w", err)
	}
	if len(data) == 0 {
		return New(), nil
	}
	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return New(), nil
	}
	if c.Repos == nil {
		c.Repos = make(map[string]RepoEntry)
	}
	return &c, nil
}

func FilterByRepos(c *Cache, repos []string) map[string]RepoEntry {
	filtered := make(map[string]RepoEntry, len(repos))
	if c == nil || len(repos) == 0 {
		return filtered
	}
	set := make(map[string]bool, len(repos))
	for _, r := range repos {
		set[r] = true
	}
	for path, entry := range c.Repos {
		if set[path] {
			filtered[path] = entry
		}
	}
	return filtered
}

func Save(path string, c *Cache) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing cache: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomic rename: %w", err)
	}
	return nil
}
