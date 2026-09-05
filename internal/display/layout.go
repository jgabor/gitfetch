package display

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jgabor/gitfetch/internal/core"
	"github.com/jgabor/gitfetch/internal/decay"
	"github.com/jgabor/gitfetch/internal/theme"
)

type Column struct {
	Title string
	Width int
}
type Layout struct {
	Columns     []Column
	BarWidth    int
	ActivityMax int
}

// NewLayout shares column sizing and activity scale between static and interactive views.
func NewLayout(rows []core.RepoRow, width int) Layout {
	if width <= 0 {
		width = 80
	}
	l := Layout{}
	if width >= 64 {
		l.BarWidth = 8
	}
	commitW, releaseW, versionW := 7, 7, 7
	for _, r := range rows {
		if r.HasCommit {
			commitW = max(commitW, len(fmt.Sprintf("%dd", r.CommitDays)))
		}
		if r.HasTag {
			releaseW = max(releaseW, len(fmt.Sprintf("%dd", r.TagDays)))
		}
		versionW = min(16, max(versionW, lipgloss.Width(r.Tag)))
		for _, n := range r.WeeklyCommits {
			l.ActivityMax = max(l.ActivityMax, n)
		}
	}
	if l.BarWidth > 0 {
		commitW += l.BarWidth + 1
		releaseW += l.BarWidth + 1
	}
	l.Columns = []Column{{"Repo", 4}, {"Commit", commitW}, {"Release", releaseW}, {"Version", versionW}}
	// Even very small terminals retain a useful repo column.
	for len(l.Columns) > 1 {
		fixed := len(l.Columns) - 1
		for _, c := range l.Columns[1:] {
			fixed += c.Width
		}
		if width-fixed >= 4 {
			break
		}
		l.Columns = l.Columns[:len(l.Columns)-1]
	}
	fixed := len(l.Columns) - 1
	for _, c := range l.Columns[1:] {
		fixed += c.Width
	}
	l.Columns[0].Width = max(1, min(core.RepoColumnWidth(rows), width-fixed))
	return l
}

func (l Layout) Cells(r core.RepoRow) []string {
	tag := r.Tag
	if tag == "" {
		tag = "—"
	}
	commit := ageCell(r.CommitDays, r.HasCommit, 0)
	if l.BarWidth > 0 {
		activity := fmt.Sprintf("%-8s", Sparkline(r.WeeklyCommits, l.ActivityMax))
		if r.HasCommit {
			activity = lipgloss.NewStyle().Foreground(theme.AgeColor(r.CommitDays)).Render(activity)
		}
		commit = activity + " " + strings.Repeat(" ", max(0, 7-lipgloss.Width(commit))) + commit
	}
	cells := []string{r.Name, commit, ageCell(r.TagDays, r.HasTag, l.BarWidth), tag}
	if r.Error != "" {
		cells[1], cells[2], cells[3] = "error", "—", "—"
	}
	return cells[:len(l.Columns)]
}

func ageCell(days int, present bool, barWidth int) string {
	if !present {
		return "—"
	}
	age := fmt.Sprintf("%dd", days)
	style := lipgloss.NewStyle().Foreground(theme.AgeColor(days))
	if barWidth == 0 {
		return style.Render(age)
	}
	return theme.GradientBar(days, decay.OverallProgress(days), barWidth) + " " + style.Render(fmt.Sprintf("%7s", age))
}

// Sparkline uses one linear scale for all repos. A middle dot is a known zero;
// a dash is a week for which no scan data is available.
func Sparkline(counts []int, scale int) string {
	if len(counts) != 8 {
		return "—"
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	var b strings.Builder
	for _, n := range counts {
		switch {
		case n < 0:
			b.WriteRune('—')
		case n == 0:
			b.WriteRune('·')
		default:
			index := min(7, (n*8-1)/max(1, scale))
			b.WriteRune(levels[index])
		}
	}
	return b.String()
}

func (l Layout) Line(cells []string) string {
	parts := make([]string, len(l.Columns))
	for i, c := range l.Columns {
		s := ansi.Truncate(strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(cells[i]), c.Width, "…")
		parts[i] = s + strings.Repeat(" ", max(0, c.Width-lipgloss.Width(s)))
	}
	return strings.TrimRight(strings.Join(parts, " "), " ")
}

func Summary(rows []core.RepoRow, now time.Time) string {
	counts := [4]int{}
	errors, unknown := 0, 0
	var oldest, newest time.Time
	missingScan := false
	for _, r := range rows {
		if r.Error != "" {
			errors++
		} else if !r.HasCommit {
			unknown++
		} else {
			counts[r.CommitTier]++
		}
		if r.ScannedAt.IsZero() {
			missingScan = true
			continue
		}
		if oldest.IsZero() || r.ScannedAt.Before(oldest) {
			oldest = r.ScannedAt
		}
		if newest.IsZero() || r.ScannedAt.After(newest) {
			newest = r.ScannedAt
		}
	}
	s := fmt.Sprintf("%d repos · %d fresh / %d stale / %d decayed / %d dead", len(rows), counts[0], counts[1], counts[2], counts[3])
	if errors > 0 {
		s += fmt.Sprintf(" / %d errors", errors)
	}
	if unknown > 0 {
		s += fmt.Sprintf(" / %d unknown", unknown)
	}
	if oldest.IsZero() {
		return s + " · scan age unknown"
	}
	age := scanAge(now.Sub(newest))
	if scanAge(now.Sub(oldest)) != age {
		age += "–" + scanAge(now.Sub(oldest))
	}
	s += " · scanned " + age + " ago"
	if missingScan {
		s += " (some unknown)"
	}
	return s
}

func scanAge(d time.Duration) string {
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
