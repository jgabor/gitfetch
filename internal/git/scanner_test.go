package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
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

func TestScanRepoCollectsAuthorsWhenFilterSet(t *testing.T) {
	dir := makeRepoWithCommit(t)
	result := ScanRepo(dir, ScanOptions{Author: "Test"})
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if len(result.Authors) == 0 {
		t.Error("expected non-empty authors list when author filter is set")
	}
}

func TestScanRepoSkipsAuthorsWhenNoFilter(t *testing.T) {
	dir := makeRepoWithCommit(t)
	result := ScanRepo(dir, ScanOptions{})
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if len(result.Authors) != 0 {
		t.Errorf("expected no authors collected when filter empty (perf guard); got %v", result.Authors)
	}
}

func TestScanRepoSkipsRemotesWhenNoFilter(t *testing.T) {
	dir := makeRepoWithCommit(t)
	result := ScanRepo(dir, ScanOptions{})
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if len(result.Remotes) != 0 {
		t.Errorf("expected no remotes collected when filter empty (perf guard); got %v", result.Remotes)
	}
}

func TestDiscoverReposBareDirectory(t *testing.T) {
	dir := t.TempDir()
	repos, err := DiscoverRepos(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0] != dir {
		t.Errorf("bare directory should return [dir] as fallback; got %v", repos)
	}
}

func TestDiscoverReposMultipleNested(t *testing.T) {
	parent := t.TempDir()
	for _, name := range []string{"r1", "r2", "r3"} {
		sub := filepath.Join(parent, name)
		os.MkdirAll(sub, 0o755)
		gitInit(t, sub)
	}
	repos, err := DiscoverRepos(parent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("expected 3 nested repos, got %d: %v", len(repos), repos)
	}
}

func TestDiscoverReposSingleRepoDirect(t *testing.T) {
	repo := makeRepoWithCommit(t)
	repos, err := DiscoverRepos(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0] != repo {
		t.Errorf("single repo should return itself; got %v", repos)
	}
}

func TestDiscoverReposUnreadableDirectoryReturnsError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	parent := t.TempDir()
	locked := filepath.Join(parent, "locked")
	if err := os.Mkdir(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })

	_, err := DiscoverRepos(locked)
	if err == nil {
		t.Fatal("expected error for unreadable directory")
	}
	if !strings.Contains(err.Error(), "discover repos") {
		t.Errorf("error missing context wrapping; got: %v", err)
	}
}

func TestMatchFilterCaseInsensitive(t *testing.T) {
	if !matchFilter([]string{"Alice Example"}, "alice") {
		t.Error("lowercase query against Titlecase haystack should match")
	}
	if !matchFilter([]string{"alice"}, "ALICE") {
		t.Error("uppercase query against lowercase haystack should match")
	}
	if !matchFilter([]string{"GitHub.com/owner"}, "github") {
		t.Error("mixed-case substring should match")
	}
}

func TestMatchFilterNoMatch(t *testing.T) {
	if matchFilter([]string{"alice", "bob"}, "charlie") {
		t.Error("unrelated query should not match")
	}
	if matchFilter(nil, "anything") {
		t.Error("nil haystack should not match")
	}
	if matchFilter([]string{}, "anything") {
		t.Error("empty haystack should not match")
	}
}

func TestListRemotesErrorWrapped(t *testing.T) {
	dir := t.TempDir()
	_, err := listRemotes(dir)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
	if !strings.Contains(err.Error(), "listing remotes") {
		t.Errorf("error missing context wrapping; got: %v", err)
	}
}

func TestListRemotesSucceedsOnGitRepo(t *testing.T) {
	dir := makeRepoWithCommit(t)
	remotes, err := listRemotes(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(remotes) != 0 {
		t.Errorf("expected no remotes on fresh repo; got %v", remotes)
	}
}

func TestWeeklyCommitsBoundariesAndNonMonotonicDates(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	scannedAt := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	// The window is July 6 through August 31. Older ancestors behind an
	// out-of-order timestamp must still be traversed; the current week is out.
	dates := []string{
		"2026-07-06T00:00:00Z", "2026-07-12T23:59:59Z",
		"2026-07-13T00:00:00Z", "2026-08-30T23:59:59Z",
		"2026-07-05T23:59:59Z", "2026-08-31T00:00:00Z",
	}
	for _, date := range dates {
		cmd := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", date)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2020-01-01T00:00:00Z", "GIT_COMMITTER_DATE="+date)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("commit: %v: %s", err, out)
		}
	}
	got, err := weeklyCommits(dir, scannedAt)
	if err != nil {
		t.Fatal(err)
	}
	want := []int{2, 1, 0, 0, 0, 0, 0, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("weekly counts = %v, want %v", got, want)
	}
	result := ScanRepo(dir, ScanOptions{})
	if result.Error != "" || len(result.WeeklyCommits) != 8 {
		t.Fatalf("scan activity: %+v", result)
	}
}

func TestScanRepoEmptyActivity(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	result := ScanRepo(dir, ScanOptions{})
	if !reflect.DeepEqual(result.WeeklyCommits, make([]int, 8)) {
		t.Fatalf("empty repository activity = %v", result.WeeklyCommits)
	}
	invalid := ScanRepo(t.TempDir(), ScanOptions{})
	if invalid.WeeklyCommits != nil {
		t.Fatalf("invalid repository should have unknown activity: %v", invalid.WeeklyCommits)
	}
}
