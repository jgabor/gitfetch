package display

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	ltable "charm.land/lipgloss/v2/table"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/core"
	"github.com/jgabor/gitfetch/internal/decay"
	"github.com/jgabor/gitfetch/internal/theme"
)

func FormatDashboard(repos map[string]cache.RepoEntry, verbose bool) string {
	var b strings.Builder

	if verbose {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
		b.WriteString(headerStyle.Render("gitfetch -- repo decay tracker"))
		b.WriteString("\n\n")
	}

	if len(repos) == 0 {
		if verbose {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("No repos tracked. Run `gitfetch refresh` to scan."))
			b.WriteString("\n")
		}
		return b.String()
	}

	rows, _ := core.BuildRows(repos)
	headerCellStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Padding(0, 1)
	cellStyle := lipgloss.NewStyle().Padding(0, 1)

	tbl := ltable.New().
		Border(lipgloss.NormalBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		BorderRow(false).
		BorderHeader(verbose).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == ltable.HeaderRow {
				return headerCellStyle
			}
			return cellStyle
		})

	if verbose {
		tbl.Headers("Repo", "Version", "Tier", "Decay", "Age")
	}

	for _, r := range rows {
		if r.Error != "" {
			tbl.Row(r.Name, "", lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("error"), r.Error, "")
			continue
		}
		tier := lipgloss.NewStyle().
			Foreground(lipgloss.Color(r.CommitTier.Color())).
			Render(r.CommitTier.String())
		decayCell := theme.GradientBar(r.CommitDays, r.CommitProgress)
		if r.HasTag {
			decayCell = decayCell + " " + theme.GradientBar(r.TagDays, r.TagProgress)
		}
		tbl.Row(
			r.Name,
			r.Tag,
			tier,
			decayCell,
			fmt.Sprintf("%dd", r.CommitDays),
		)
	}

	b.WriteString(tbl.Render())
	b.WriteString("\n")

	if verbose {
		b.WriteString("\n")
		swatch := func(t decay.Tier) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Color())).Render("█ " + t.String())
		}
		fmt.Fprintf(&b, "Legend: %s  %s  %s  %s",
			swatch(decay.Fresh), swatch(decay.Stale), swatch(decay.Decayed), swatch(decay.Dead))
		b.WriteString("\n")
	}

	return b.String()
}
