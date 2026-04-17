package display

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

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
	commitDate := makeDate(10)
	tagDate := makeDate(5)
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

func TestFormatDashboardEmpty(t *testing.T) {
	out := FormatDashboard(map[string]cache.RepoEntry{})
	if !strings.Contains(out, "No repos tracked") {
		t.Errorf("empty dashboard missing hint: %q", out)
	}
}

func TestFormatDashboardHeader(t *testing.T) {
	out := FormatDashboard(map[string]cache.RepoEntry{})
	if !strings.Contains(out, "gitfetch") {
		t.Errorf("dashboard missing header: %q", out)
	}
}

func TestFormatDashboardLegend(t *testing.T) {
	commitDate := makeDate(10)
	repos := map[string]cache.RepoEntry{
		"/r1": {LastCommitDate: &commitDate, ScannedAt: time.Now().UTC()},
	}
	out := FormatDashboard(repos)
	if !strings.Contains(out, "Legend:") {
		t.Errorf("dashboard missing legend: %q", out)
	}
}

func TestFormatDashboardMultipleRepos(t *testing.T) {
	commitDate := makeDate(10)
	tagDate := makeDate(5)
	staleDate := makeDate(60)
	repos := map[string]cache.RepoEntry{
		"/home/user/fresh-project": {
			LastCommitDate: &commitDate,
			LastTagDate:    &tagDate,
			ScannedAt:      time.Now().UTC(),
		},
		"/home/user/stale-project": {
			LastCommitDate: &staleDate,
			ScannedAt:      time.Now().UTC(),
		},
	}
	out := FormatDashboard(repos)
	if !strings.Contains(out, "fresh-project") {
		t.Errorf("missing fresh-project: %q", out)
	}
	if !strings.Contains(out, "stale-project") {
		t.Errorf("missing stale-project: %q", out)
	}
}

func TestPlainBarWidth(t *testing.T) {
	bar := plainBar(decay.Fresh, 0.5)
	if utf8.RuneCountInString(bar) != barWidth {
		t.Errorf("bar width = %d runes, want %d", utf8.RuneCountInString(bar), barWidth)
	}
}

func TestPlainBarFull(t *testing.T) {
	bar := plainBar(decay.Dead, 1.0)
	for _, ch := range bar {
		if ch != '#' {
			t.Errorf("expected all filled, got %c", ch)
			break
		}
	}
}

func TestPlainBarEmpty(t *testing.T) {
	bar := plainBar(decay.Fresh, 0.0)
	for _, ch := range bar {
		if ch != '-' {
			t.Errorf("expected all empty, got %c", ch)
			break
		}
	}
}

func TestNewTableCreatesColumns(t *testing.T) {
	commitDate := makeDate(10)
	repos := map[string]cache.RepoEntry{
		"/r1": {LastCommitDate: &commitDate, ScannedAt: time.Now().UTC()},
	}
	tm, _ := NewTable(repos, 10, false)
	cols := Columns()
	if len(cols) != 6 {
		t.Errorf("expected 6 columns, got %d", len(cols))
	}
	rows := tm.Rows()
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) == 0 || rows[0][0] != "/r1" {
		t.Errorf("expected hidden path column to contain /r1, got %v", rows[0])
	}
}

