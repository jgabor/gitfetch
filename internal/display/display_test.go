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

func TestFormatDashboardDefaultHasCompactHeaders(t *testing.T) {
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
	if !strings.Contains(out, "Repo") || !strings.Contains(out, "Version") || !strings.Contains(out, "Commit") {
		t.Errorf("default dashboard must show compact table headers: %q", out)
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

func TestFormatDashboardReleaseDecayBar(t *testing.T) {
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
	if barCount != 8 {
		t.Errorf("expected one eight-character release bar, got %d glyphs", barCount)
	}
}

func TestFormatDashboardCommitHasNoDecayBar(t *testing.T) {
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
	if barCount != 0 {
		t.Errorf("commit must not have a decay bar, got %d glyphs", barCount)
	}
}

func TestResponsiveDashboard(t *testing.T) {
	now := time.Now()
	repos := map[string]cache.RepoEntry{
		"/超長い名前-project": {LastCommitDate: &now, LastTagDate: &now, LastTag: "v1.2.3", ScannedAt: now, WeeklyCommits: []int{0, 1, 2, 3, 4, 5, 6, 8}},
		"/unknown":       {},
	}
	for _, width := range []int{1, 20, 40, 64, 80, 120} {
		out := FormatDashboardWidth(repos, false, width)
		for _, line := range strings.Split(out, "\n") {
			if ansi.StringWidth(line) > width {
				t.Errorf("width %d overflow: %q", width, line)
			}
		}
		if width >= 64 {
			if !strings.Contains(out, "Commit") || strings.Contains(out, "Commit 8w") || !strings.Contains(ansi.Strip(out), "·▁▂▃▄▅▆█      0d") {
				t.Errorf("commit column must pair activity with age: %s", out)
			}
			if strings.Count(out, "·▁▂▃▄▅▆█") != 1 || strings.Contains(out, "Activity 8w") {
				t.Errorf("activity must appear only in commit column: %s", out)
			}
		}
		if width < 64 && strings.Contains(out, "░") {
			t.Errorf("narrow output contains bars: %s", out)
		}
	}
	out := ansi.Strip(FormatDashboard(repos, false))
	if !strings.Contains(out, "1 unknown") {
		t.Errorf("missing date counted as fresh: %s", out)
	}
}

func TestSparklineSharedScaleAndMissing(t *testing.T) {
	if got := Sparkline([]int{0, 1, 2, 4, 8, 16, 32, 64}, 64); got != "·▁▁▁▁▂▄█" {
		t.Errorf("got %q", got)
	}
	if got := Sparkline([]int{-1, 0, 1, 2, 3, 4, 5, 8}, 64); got != "—·▁▁▁▁▁▁" {
		t.Errorf("got %q", got)
	}
	if got := Sparkline(nil, 64); got != "—" {
		t.Errorf("legacy cache activity: %q", got)
	}
}

func TestSummaryPartialRefresh(t *testing.T) {
	now := time.Now()
	old := now.Add(-48 * time.Hour)
	repos := map[string]cache.RepoEntry{"/old": {LastCommitDate: &old, ScannedAt: old}, "/new": {LastCommitDate: &now, ScannedAt: now}, "/bad": {Error: "broken"}}
	out := FormatDashboard(repos, false)
	for _, want := range []string{"2 fresh", "1 errors", "scanned <1m–2d ago", "some unknown"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q: %s", want, out)
		}
	}
}
