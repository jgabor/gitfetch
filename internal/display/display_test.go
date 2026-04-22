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
