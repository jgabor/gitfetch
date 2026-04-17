package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ScanResult struct {
	RepoPath       string     `json:"repo_path"`
	LastCommitDate *time.Time `json:"last_commit_date,omitempty"`
	LastTagDate    *time.Time `json:"last_tag_date,omitempty"`
	LastTag        string     `json:"last_tag,omitempty"`
	Authors        []string   `json:"authors,omitempty"`
	Remotes        []string   `json:"remotes,omitempty"`
	Error          string     `json:"error,omitempty"`
	ScannedAt      time.Time  `json:"scanned_at"`
}

type ScanOptions struct {
	Author string
	Remote string
}

func ScanRepo(repoPath string, opts ScanOptions) ScanResult {
	result := ScanResult{
		RepoPath:  repoPath,
		ScannedAt: time.Now().UTC().Truncate(time.Second),
	}

	if opts.Author != "" {
		authors, err := listAuthors(repoPath)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		result.Authors = authors
		if !matchFilter(authors, opts.Author) {
			return result
		}
	}

	if opts.Remote != "" {
		remotes, err := listRemotes(repoPath)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		result.Remotes = remotes
		if !matchFilter(remotes, opts.Remote) {
			return result
		}
	}

	commitDate, err := lastCommitDate(repoPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.LastCommitDate = &commitDate

	tagName, tagDate, _ := lastTagInfo(repoPath)
	result.LastTag = tagName
	result.LastTagDate = tagDate

	return result
}

func ScanAll(repoPaths []string, opts ScanOptions) []ScanResult {
	expanded := ResolveRepoPaths(repoPaths)
	results := make([]ScanResult, 0, len(expanded))
	for _, path := range expanded {
		r := ScanRepo(path, opts)
		if opts.Author != "" || opts.Remote != "" {
			if r.LastCommitDate == nil && r.Error == "" {
				continue
			}
		}
		results = append(results, r)
	}
	return results
}

func ResolveRepoPaths(paths []string) []string {
	var resolved []string
	seen := make(map[string]bool)
	for _, p := range paths {
		repos, err := DiscoverRepos(p)
		if err != nil {
			repos = []string{p}
		}
		for _, repo := range repos {
			if !seen[repo] {
				seen[repo] = true
				resolved = append(resolved, repo)
			}
		}
	}
	return resolved
}

func DiscoverRepos(path string) ([]string, error) {
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		return []string{path}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("discover repos %s: %w", path, err)
	}
	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(path, e.Name())
		if _, err := os.Stat(filepath.Join(sub, ".git")); err == nil {
			repos = append(repos, sub)
		}
	}
	if len(repos) == 0 {
		return []string{path}, nil
	}
	return repos, nil
}

func listAuthors(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "log", "--format=%aN")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing authors: %w", sanitizeGitError(err, out))
	}
	seen := make(map[string]bool)
	var authors []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name := strings.TrimSpace(line)
		if name != "" && !seen[name] {
			seen[name] = true
			authors = append(authors, name)
		}
	}
	return authors, nil
}

func listRemotes(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "-v")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing remotes: %w", sanitizeGitError(err, out))
	}
	seen := make(map[string]bool)
	var remotes []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && !seen[parts[1]] {
			seen[parts[1]] = true
			remotes = append(remotes, parts[1])
		}
	}
	return remotes, nil
}

func matchFilter(values []string, pattern string) bool {
	for _, v := range values {
		if strings.Contains(strings.ToLower(v), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func lastCommitDate(repoPath string) (time.Time, error) {
	cmd := exec.Command("git", "-C", repoPath, "log", "-1", "--format=%ct")
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("last commit date: %w", sanitizeGitError(err, out))
	}
	ts := strings.TrimSpace(string(out))
	if ts == "" {
		return time.Time{}, fmt.Errorf("last commit date: empty repository")
	}
	return parseTimestamp(ts)
}

func lastTagInfo(repoPath string) (string, *time.Time, error) {
	cmd := exec.Command("git", "-C", repoPath, "for-each-ref",
		"--sort=-creatordate", "--format=%(refname:short) %(creatordate:unix)",
		"--count=1", "refs/tags/v*",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", nil, fmt.Errorf("last tag: %w", sanitizeGitError(err, out))
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return "", nil, nil
	}
	parts := strings.SplitN(line, " ", 2)
	tagName := parts[0]
	if len(parts) < 2 {
		return tagName, nil, nil
	}
	t, err := parseTimestamp(strings.TrimSpace(parts[1]))
	if err != nil {
		return tagName, nil, err
	}
	return tagName, &t, nil
}

func parseTimestamp(s string) (time.Time, error) {
	unixSec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp %q: %w", s, err)
	}
	return time.Unix(unixSec, 0).UTC(), nil
}

func sanitizeGitError(err error, out []byte) error {
	msg := strings.TrimSpace(string(out))
	if msg != "" {
		return fmt.Errorf("%s", msg)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		if stderr != "" {
			return fmt.Errorf("%s", stderr)
		}
	}
	return err
}
