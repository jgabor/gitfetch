package display

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/jgabor/gitfetch/internal/cache"
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
	if strings.Contains(out, "Repo") && strings.Contains(out, "Version") && strings.Contains(out, "Commit") {
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

func TestFormatDashboardDualDecayBars(t *testing.T) {
	commitDate := makeDate(10)
	tagDate := makeDate(5)
	repos := map[string]cache.RepoEntry{
		"/r1": {
			LastCommitDate: &commitDate,
			LastTagDate:    &tagDate,
			LastTag:        "v1.0.0",
			ScannedAt:      time.Now().UTC(),
		},
	}
	out := FormatDashboard(repos, false)
	stripped := ansi.Strip(out)
	barCount := strings.Count(stripped, "█") + strings.Count(stripped, "░")
	if barCount != 30 {
		t.Errorf("expected exactly 30 bar glyphs for commit bar, got %d", barCount)
	}
}

func TestFormatDashboardSingleDecayBar(t *testing.T) {
	commitDate := makeDate(10)
	repos := map[string]cache.RepoEntry{
		"/r1": {
			LastCommitDate: &commitDate,
			ScannedAt:      time.Now().UTC(),
		},
	}
	out := FormatDashboard(repos, false)
	stripped := ansi.Strip(out)
	barCount := strings.Count(stripped, "█") + strings.Count(stripped, "░")
	if barCount != 30 {
		t.Errorf("expected exactly 30 bar glyphs for single commit bar, got %d", barCount)
	}
}


