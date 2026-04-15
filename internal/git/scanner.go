package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type ScanResult struct {
	RepoPath       string     `json:"repo_path"`
	LastCommitDate *time.Time `json:"last_commit_date,omitempty"`
	LastTagDate    *time.Time `json:"last_tag_date,omitempty"`
	Error          string     `json:"error,omitempty"`
	ScannedAt      time.Time  `json:"scanned_at"`
}

func ScanRepo(repoPath string) ScanResult {
	result := ScanResult{
		RepoPath:  repoPath,
		ScannedAt: time.Now().UTC().Truncate(time.Second),
	}

	commitDate, err := lastCommitDate(repoPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.LastCommitDate = &commitDate

	tagDate, _ := lastTagDate(repoPath)
	result.LastTagDate = tagDate

	return result
}

func ScanAll(repoPaths []string) []ScanResult {
	results := make([]ScanResult, 0, len(repoPaths))
	for _, path := range repoPaths {
		results = append(results, ScanRepo(path))
	}
	return results
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

func lastTagDate(repoPath string) (*time.Time, error) {
	cmd := exec.Command("git", "-C", repoPath, "for-each-ref",
		"--sort=-creatordate", "--format=%(creatordate:unix)",
		"--count=1", "refs/tags/v*",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("last tag date: %w", sanitizeGitError(err, out))
	}
	ts := strings.TrimSpace(string(out))
	if ts == "" {
		return nil, nil
	}
	t, err := parseTimestamp(ts)
	if err != nil {
		return nil, err
	}
	return &t, nil
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
