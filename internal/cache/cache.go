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
	Error          string     `json:"error,omitempty"`
	ScannedAt      time.Time  `json:"scanned_at"`
}

type Cache struct {
	Repos map[string]RepoEntry `json:"repos"`
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
		return nil, fmt.Errorf("parsing cache %s: %w", path, err)
	}
	if c.Repos == nil {
		c.Repos = make(map[string]RepoEntry)
	}
	return &c, nil
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
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing cache: %w", err)
	}
	return nil
}
