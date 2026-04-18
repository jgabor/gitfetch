package display

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
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

func TestFormatDashboardEmptyVerbose(t *testing.T) {
	out := FormatDashboard(map[string]cache.RepoEntry{}, true)
	if !strings.Contains(out, "No repos tracked") {
		t.Errorf("verbose empty dashboard missing hint: %q", out)
	}
}

func TestFormatDashboardEmptyQuiet(t *testing.T) {
	out := FormatDashboard(map[string]cache.RepoEntry{}, false)
	if strings.Contains(out, "No repos tracked") {
		t.Errorf("quiet empty dashboard should omit hint: %q", out)
	}
	if strings.Contains(out, "gitfetch") {
		t.Errorf("quiet dashboard should omit app header: %q", out)
	}
}

func TestFormatDashboardHeader(t *testing.T) {
	out := FormatDashboard(map[string]cache.RepoEntry{}, true)
	if !strings.Contains(out, "gitfetch") {
		t.Errorf("verbose dashboard missing header: %q", out)
	}
}

func TestFormatDashboardLegendVerbose(t *testing.T) {
	commitDate := makeDate(10)
	repos := map[string]cache.RepoEntry{
		"/r1": {LastCommitDate: &commitDate, ScannedAt: time.Now().UTC()},
	}
	out := FormatDashboard(repos, true)
	if !strings.Contains(out, "Legend:") {
		t.Errorf("verbose dashboard missing legend: %q", out)
	}
}

func TestFormatDashboardQuietOmitsChrome(t *testing.T) {
	commitDate := makeDate(10)
	repos := map[string]cache.RepoEntry{
		"/home/user/r1": {LastCommitDate: &commitDate, ScannedAt: time.Now().UTC()},
	}
	out := FormatDashboard(repos, false)
	if strings.Contains(out, "Legend:") {
		t.Errorf("quiet dashboard must omit legend: %q", out)
	}
	if strings.Contains(out, "gitfetch --") {
		t.Errorf("quiet dashboard must omit app header: %q", out)
	}
	if strings.Contains(out, "Repo") && strings.Contains(out, "Version") && strings.Contains(out, "Decay") {
		t.Errorf("quiet dashboard must omit table headers: %q", out)
	}
	if !strings.Contains(out, "r1") {
		t.Errorf("quiet dashboard must still show repo row: %q", out)
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
	out := FormatDashboard(repos, true)
	if !strings.Contains(out, "fresh-project") {
		t.Errorf("missing fresh-project: %q", out)
	}
	if !strings.Contains(out, "stale-project") {
		t.Errorf("missing stale-project: %q", out)
	}
}

func TestRenderBarWidth(t *testing.T) {
	bar := colorBar(decay.Fresh, 0.5)
	if got := ansi.StringWidth(bar); got != barWidth {
		t.Errorf("bar width = %d cells, want %d", got, barWidth)
	}
}

func TestRenderBarFull(t *testing.T) {
	bar := colorBar(decay.Dead, 1.0)
	plain := ansi.Strip(bar)
	if !strings.Contains(plain, "█") {
		t.Errorf("expected filled bar to contain '█', got %q", plain)
	}
	if strings.Contains(plain, "░") {
		t.Errorf("full bar must not contain empty glyph '░', got %q", plain)
	}
}

func TestRenderBarEmpty(t *testing.T) {
	bar := colorBar(decay.Fresh, 0.0)
	plain := ansi.Strip(bar)
	if !strings.Contains(plain, "░") {
		t.Errorf("expected empty bar to contain '░', got %q", plain)
	}
	if strings.Contains(plain, "█") {
		t.Errorf("empty bar must not contain filled glyph '█', got %q", plain)
	}
}

func TestNewTableCreatesColumns(t *testing.T) {
	commitDate := makeDate(10)
	repos := map[string]cache.RepoEntry{
		"/r1": {LastCommitDate: &commitDate, ScannedAt: time.Now().UTC()},
	}
	tm, _ := NewTable(repos, 10, false)
	cols := Columns(colRepoMin)
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
	got := repoColumnWidth(rows)
	if got != len("a-very-long-repository-name") {
		t.Errorf("repoColumnWidth = %d, want %d", got, len("a-very-long-repository-name"))
	}
}

func TestRepoColumnWidthRespectsMin(t *testing.T) {
	if got := repoColumnWidth(nil); got != colRepoMin {
		t.Errorf("repoColumnWidth(nil) = %d, want %d", got, colRepoMin)
	}
	if got := repoColumnWidth([]RepoRow{{Name: "x"}}); got != colRepoMin {
		t.Errorf("repoColumnWidth(short) = %d, want %d", got, colRepoMin)
	}
}

