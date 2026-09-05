package core

import (
	"testing"
	"time"

	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/decay"
)

func makeDate(daysAgo int) time.Time {
	return time.Now().UTC().Add(-time.Duration(daysAgo) * 24 * time.Hour)
}

func TestBuildRowsEmpty(t *testing.T) {
	rows, _ := BuildRows(map[string]cache.RepoEntry{})
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestBuildRowsFreshRepo(t *testing.T) {
	commitDate := makeDate(5)
	tagDate := makeDate(2)
	repos := map[string]cache.RepoEntry{
		"/home/user/project": {
			LastCommitDate: &commitDate,
			LastTagDate:    &tagDate,
			ScannedAt:      time.Now().UTC(),
		},
	}
	rows, _ := BuildRows(repos)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.Name != "project" {
		t.Errorf("name = %q, want project", r.Name)
	}
	if r.CommitTier != decay.Fresh {
		t.Errorf("commit tier = %s, want fresh", r.CommitTier)
	}
	if !r.HasTag {
		t.Error("expected HasTag = true")
	}
	if r.TagTier != decay.Fresh {
		t.Errorf("tag tier = %s, want fresh", r.TagTier)
	}
}

func TestBuildRowsErrorRepo(t *testing.T) {
	repos := map[string]cache.RepoEntry{
		"/bad/repo": {
			Error:     "not a git repository",
			ScannedAt: time.Now().UTC(),
		},
	}
	rows, _ := BuildRows(repos)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.Error != "not a git repository" {
		t.Errorf("error = %q, want not a git repository", r.Error)
	}
}

func TestBuildRowsNoTag(t *testing.T) {
	commitDate := makeDate(60)
	repos := map[string]cache.RepoEntry{
		"/home/user/notags": {
			LastCommitDate: &commitDate,
			ScannedAt:      time.Now().UTC(),
		},
	}
	rows, _ := BuildRows(repos)
	r := rows[0]
	if r.HasTag {
		t.Error("expected HasTag = false")
	}
	if r.CommitTier != decay.Stale {
		t.Errorf("commit tier = %s, want stale", r.CommitTier)
	}
}

func TestBuildRowsMultipleTiers(t *testing.T) {
	freshDate := makeDate(5)
	staleDate := makeDate(60)
	decayedDate := makeDate(120)
	deadDate := makeDate(200)
	repos := map[string]cache.RepoEntry{
		"/r1": {LastCommitDate: &freshDate, ScannedAt: time.Now().UTC()},
		"/r2": {LastCommitDate: &staleDate, ScannedAt: time.Now().UTC()},
		"/r3": {LastCommitDate: &decayedDate, ScannedAt: time.Now().UTC()},
		"/r4": {LastCommitDate: &deadDate, ScannedAt: time.Now().UTC()},
	}
	rows, _ := BuildRows(repos)
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}
	wantTiers := map[string]decay.Tier{}
	for _, row := range rows {
		switch row.Name {
		case "r1":
			wantTiers[row.Name] = decay.Fresh
		case "r2":
			wantTiers[row.Name] = decay.Stale
		case "r3":
			wantTiers[row.Name] = decay.Decayed
		case "r4":
			wantTiers[row.Name] = decay.Dead
		}
		if row.CommitTier != wantTiers[row.Name] {
			t.Errorf("%s tier = %s, want %s", row.Name, row.CommitTier, wantTiers[row.Name])
		}
	}
}

func TestBuildRowsSortsByMostDecayed(t *testing.T) {
	freshDate := makeDate(5)
	staleDate := makeDate(60)
	decayedDate := makeDate(120)
	deadDate := makeDate(300)
	repos := map[string]cache.RepoEntry{
		"/zzz-fresh":   {LastCommitDate: &freshDate, ScannedAt: time.Now().UTC()},
		"/aaa-dead":    {LastCommitDate: &deadDate, ScannedAt: time.Now().UTC()},
		"/mid-decayed": {LastCommitDate: &decayedDate, ScannedAt: time.Now().UTC()},
		"/mid-stale":   {LastCommitDate: &staleDate, ScannedAt: time.Now().UTC()},
		"/bad":         {Error: "scan failed", ScannedAt: time.Now().UTC()},
	}
	rows, _ := BuildRows(repos)
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}
	wantOrder := []string{"aaa-dead", "mid-decayed", "mid-stale", "zzz-fresh", "bad"}
	for i, want := range wantOrder {
		if rows[i].Name != want {
			t.Errorf("rows[%d] = %s, want %s (full order %v)", i, rows[i].Name, want,
				[]string{rows[0].Name, rows[1].Name, rows[2].Name, rows[3].Name, rows[4].Name})
		}
	}
}

func TestRepoColumnWidthUsesLongestName(t *testing.T) {
	rows := []RepoRow{
		{Name: "short"},
		{Name: "a-very-long-repository-name"},
		{Name: "mid"},
	}
	got := RepoColumnWidth(rows)
	if got != len("a-very-long-repository-name") {
		t.Errorf("repoColumnWidth = %d, want %d", got, len("a-very-long-repository-name"))
	}
}

func TestRepoColumnWidthRespectsMin(t *testing.T) {
	if got := RepoColumnWidth(nil); got != colRepoMin {
		t.Errorf("repoColumnWidth(nil) = %d, want %d", got, colRepoMin)
	}
	if got := RepoColumnWidth([]RepoRow{{Name: "x"}}); got != colRepoMin {
		t.Errorf("repoColumnWidth(short) = %d, want %d", got, colRepoMin)
	}
}

func TestAlignedActivity(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	entry := cache.RepoEntry{ScannedAt: now.AddDate(0, 0, -14), WeeklyCommits: []int{1, 2, 3, 4, 5, 6, 7, 8}}
	got := alignedActivity(entry, now)
	want := []int{3, 4, 5, 6, 7, 8, -1, -1}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	entry.ScannedAt = now.AddDate(0, 0, -70)
	for _, n := range alignedActivity(entry, now) {
		if n != -1 {
			t.Fatal("expired bins must be unknown")
		}
	}
	if alignedActivity(cache.RepoEntry{}, now) != nil {
		t.Fatal("legacy data must be unavailable")
	}
}
