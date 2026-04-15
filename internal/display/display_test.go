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
	t := time.Now().UTC().Add(-time.Duration(daysAgo) * 24 * time.Hour)
	return t
}

func TestBuildRowsEmpty(t *testing.T) {
	rows := BuildRows(map[string]cache.RepoEntry{})
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
	rows := BuildRows(repos)
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
	rows := BuildRows(repos)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.Error != "not a git repository" {
		t.Errorf("error = %q, want not a git repository", r.Error)
	}
	if r.CommitTier != decay.Fresh {
		t.Errorf("commit tier should be default (fresh) for error repos, got %s", r.CommitTier)
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
	rows := BuildRows(repos)
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
	rows := BuildRows(repos)
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

func TestFormatRowContainsName(t *testing.T) {
	row := RepoRow{
		Name:           "myproject",
		CommitTier:     decay.Fresh,
		CommitDays:     10,
		CommitProgress: 0.33,
	}
	out := FormatRow(row)
	if !strings.Contains(out, "myproject") {
		t.Errorf("output missing repo name: %q", out)
	}
}

func TestFormatRowErrorRepo(t *testing.T) {
	row := RepoRow{
		Name:  "broken",
		Error: "not a git repository",
	}
	out := FormatRow(row)
	if !strings.Contains(out, "broken") {
		t.Errorf("output missing repo name: %q", out)
	}
	if !strings.Contains(out, "error:") {
		t.Errorf("output missing error prefix: %q", out)
	}
}

func TestFormatRowWithTag(t *testing.T) {
	row := RepoRow{
		Name:           "tagged",
		CommitTier:     decay.Stale,
		CommitDays:     45,
		CommitProgress: 0.25,
		TagTier:        decay.Fresh,
		TagDays:        10,
		TagProgress:    0.33,
		HasTag:         true,
	}
	out := FormatRow(row)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (commit + tag), got %d: %q", len(lines), out)
	}
}

func TestFormatRowNoTag(t *testing.T) {
	row := RepoRow{
		Name:           "untagged",
		CommitTier:     decay.Fresh,
		CommitDays:     5,
		CommitProgress: 0.17,
		HasTag:         false,
	}
	out := FormatRow(row)
	if strings.Contains(out, "\n") {
		t.Errorf("single repo without tag should be one line: %q", out)
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

func TestStyledBarWidth(t *testing.T) {
	bar := styledBar(decay.Fresh, 0.5)
	stripped := stripANSI(bar)
	if utf8.RuneCountInString(stripped) != barWidth {
		t.Errorf("bar width = %d runes, want %d", utf8.RuneCountInString(stripped), barWidth)
	}
}

func TestStyledBarFull(t *testing.T) {
	bar := styledBar(decay.Dead, 1.0)
	stripped := stripANSI(bar)
	for _, ch := range stripped {
		if ch != '█' {
			t.Errorf("expected all filled, got %c", ch)
			break
		}
	}
}

func TestStyledBarEmpty(t *testing.T) {
	bar := styledBar(decay.Fresh, 0.0)
	stripped := stripANSI(bar)
	for _, ch := range stripped {
		if ch != '░' {
			t.Errorf("expected all empty, got %c", ch)
			break
		}
	}
}

func stripANSI(s string) string {
	var result []byte
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] >= 'a' && s[i] <= 'z' || s[i] >= 'A' && s[i] <= 'Z' {
				inEscape = false
			}
			continue
		}
		result = append(result, s[i])
	}
	return string(result)
}
