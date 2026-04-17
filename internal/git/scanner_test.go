package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func makeRepoWithCommit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitInit(t, dir)
	writeFile(t, dir, "README.md", "hello")
	gitCommit(t, dir, "initial commit")
	return dir
}

func makeRepoWithCommitAndTag(t *testing.T, tagName string) string {
	t.Helper()
	dir := makeRepoWithCommit(t)
	gitTag(t, dir, tagName)
	return dir
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", dir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	cmd = exec.Command("git", "-C", dir, "config", "user.email", "test@test.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	cmd = exec.Command("git", "-C", dir, "config", "user.name", "Test")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config name: %v", err)
	}
	cmd = exec.Command("git", "-C", dir, "config", "commit.gpgsign", "false")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config gpgsign: %v", err)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "add", "-A")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	cmd = exec.Command("git", "-C", dir, "commit", "-m", msg)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}
}

func gitTag(t *testing.T, dir, name string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "tag", name)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git tag %s: %v", name, err)
	}
}

func TestScanRepoWithCommits(t *testing.T) {
	dir := makeRepoWithCommit(t)
	result := ScanRepo(dir, ScanOptions{})

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.LastCommitDate == nil {
		t.Fatal("LastCommitDate is nil")
	}
	if result.LastCommitDate.After(time.Now().UTC()) {
		t.Errorf("commit date %v is in the future", result.LastCommitDate)
	}
	if result.ScannedAt.IsZero() {
		t.Error("ScannedAt is zero")
	}
}

func TestScanRepoWithVTag(t *testing.T) {
	dir := makeRepoWithCommitAndTag(t, "v1.0.0")
	result := ScanRepo(dir, ScanOptions{})

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.LastTagDate == nil {
		t.Fatal("LastTagDate is nil for repo with v-prefixed tag")
	}
	if result.LastTagDate.After(time.Now().UTC()) {
		t.Errorf("tag date %v is in the future", result.LastTagDate)
	}
}

func TestScanRepoNoTags(t *testing.T) {
	dir := makeRepoWithCommit(t)
	result := ScanRepo(dir, ScanOptions{})

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.LastTagDate != nil {
		t.Errorf("expected nil LastTagDate, got %v", result.LastTagDate)
	}
}

func TestScanRepoNonGitDirectory(t *testing.T) {
	dir := t.TempDir()
	result := ScanRepo(dir, ScanOptions{})

	if result.Error == "" {
		t.Fatal("expected error for non-git directory")
	}
	if result.LastCommitDate != nil {
		t.Error("LastCommitDate should be nil on error")
	}
}

func TestScanRepoMissingPath(t *testing.T) {
	result := ScanRepo("/nonexistent/path/to/repo", ScanOptions{})

	if result.Error == "" {
		t.Fatal("expected error for missing path")
	}
	if result.LastCommitDate != nil {
		t.Error("LastCommitDate should be nil on error")
	}
}

func TestScanRepoEmptyGitRepo(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)

	result := ScanRepo(dir, ScanOptions{})
	if result.Error == "" {
		t.Fatal("expected error for empty repo with no commits")
	}
}

func TestScanRepoNonVTagIgnored(t *testing.T) {
	dir := makeRepoWithCommitAndTag(t, "release-1.0")
	result := ScanRepo(dir, ScanOptions{})

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.LastTagDate != nil {
		t.Errorf("expected nil LastTagDate for non-v tag, got %v", result.LastTagDate)
	}
}

func TestScanAllMultipleRepos(t *testing.T) {
	repo1 := makeRepoWithCommit(t)
	repo2 := makeRepoWithCommitAndTag(t, "v2.0.0")
	missing := "/no/such/path"

	results := ScanAll([]string{repo1, repo2, missing}, ScanOptions{})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Error != "" {
		t.Errorf("repo1 error: %s", results[0].Error)
	}
	if results[0].LastTagDate != nil {
		t.Error("repo1 should have no tag date")
	}
	if results[1].Error != "" {
		t.Errorf("repo2 error: %s", results[1].Error)
	}
	if results[1].LastTagDate == nil {
		t.Error("repo2 should have tag date")
	}
	if results[2].Error == "" {
		t.Error("missing path should have error")
	}
}

func TestDiscoverReposExpandsDirectory(t *testing.T) {
	parent := t.TempDir()
	repo1 := filepath.Join(parent, "repo1")
	repo2 := filepath.Join(parent, "repo2")
	os.MkdirAll(repo1, 0o755)
	os.MkdirAll(repo2, 0o755)
	gitInit(t, repo1)
	writeFile(t, repo1, "f.txt", "1")
	gitCommit(t, repo1, "init")
	gitInit(t, repo2)
	writeFile(t, repo2, "f.txt", "2")
	gitCommit(t, repo2, "init")

	repos := ResolveRepoPaths([]string{parent})
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %v", len(repos), repos)
	}
}

func TestDiscoverReposSingleRepo(t *testing.T) {
	repo := makeRepoWithCommit(t)
	repos := ResolveRepoPaths([]string{repo})
	if len(repos) != 1 || repos[0] != repo {
		t.Fatalf("expected [%s], got %v", repo, repos)
	}
}

func TestScanRepoAuthorFilter(t *testing.T) {
	dir := makeRepoWithCommit(t)
	result := ScanRepo(dir, ScanOptions{Author: "Test"})
	if result.Error != "" {
		t.Fatalf("unexpected error with matching author: %s", result.Error)
	}

	result = ScanRepo(dir, ScanOptions{Author: "Nobody"})
	if result.LastCommitDate != nil {
		t.Error("expected nil LastCommitDate for non-matching author filter")
	}
}

func TestScanRepoCollectsAuthors(t *testing.T) {
	dir := makeRepoWithCommit(t)
	result := ScanRepo(dir, ScanOptions{})
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if len(result.Authors) == 0 {
		t.Error("expected non-empty authors list")
	}
}
