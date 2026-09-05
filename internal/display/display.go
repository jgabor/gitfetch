package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/core"
)

func FormatDashboard(repos map[string]cache.RepoEntry, verbose bool) string {
	return FormatDashboardWidth(repos, verbose, 80)
}

func FormatDashboardWidth(repos map[string]cache.RepoEntry, verbose bool, width int) string {
	if width <= 0 {
		width = 80
	}
	var b strings.Builder
	line := func(s string) { b.WriteString(ansi.Wrap(s, width, "")); b.WriteByte('\n') }
	if verbose {
		line("gitfetch · repo decay tracker")
	}
	if len(repos) == 0 {
		if verbose {
			line("No repos tracked. Run `gitfetch refresh` to scan.")
		}
		return b.String()
	}
	rows, _ := core.BuildRows(repos)
	l := NewLayout(rows, width)
	line(Summary(rows, time.Now()))
	headers := make([]string, len(l.Columns))
	for i, c := range l.Columns {
		headers[i] = c.Title
	}
	line(l.Line(headers))
	for _, r := range rows {
		line(l.Line(l.Cells(r)))
	}
	if verbose {
		line("Legend: longer release bars = older · fresh <10d / stale <90d / decayed <180d / dead ≥180d")
		line(fmt.Sprintf("Commit activity: 8 completed UTC weeks, oldest first · █ = %d commits/week · · zero / — unavailable", l.ActivityMax))
		for _, r := range rows {
			if r.Error != "" {
				line(r.Name + ": " + r.Error)
			}
		}
	}
	return b.String()
}
