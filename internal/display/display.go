package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/core"
	"github.com/jgabor/gitfetch/internal/decay"
)

const (
	colRepoMin = 4
	colVersion = 10
	colTier    = 8
	barWidth   = 15
	colDecay   = barWidth
	colAge     = 5

	barFilled = "█"
	barEmpty  = "░"
)

func plainBar(pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := min(int(float64(barWidth)*pct+0.5), barWidth)
	return strings.Repeat(barFilled, filled) + strings.Repeat(barEmpty, barWidth-filled)
}

func colorBar(tier decay.Tier, pct float64) string {
	bar := plainBar(pct)
	filled := min(int(float64(barWidth)*pct+0.5), barWidth)
	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(tier.Color()))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	return filledStyle.Render(bar[:len(barFilled)*filled]) +
		emptyStyle.Render(bar[len(barFilled)*filled:])
}

func truncatePlain(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return s[:maxLen-3] + "..."
	}
	return s[:maxLen]
}

func Columns(repoWidth int) []table.Column {
	return []table.Column{
		{Title: "", Width: 0},
		{Title: "Repo", Width: repoWidth},
		{Title: "Version", Width: colVersion},
		{Title: "Tier", Width: colTier},
		{Title: "Decay", Width: colDecay},
		{Title: "Age", Width: colAge},
	}
}

func TableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("6"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))
	return s
}

func rowToTableRow(row core.RepoRow, fullPath string, repoWidth int) table.Row {
	if row.Error != "" {
		return table.Row{
			fullPath,
			truncatePlain(row.Name, repoWidth),
			"",
			"error",
			truncatePlain(row.Error, colDecay),
			"",
		}
	}

	return table.Row{
		fullPath,
		truncatePlain(row.Name, repoWidth),
		truncatePlain(row.Tag, colVersion),
		row.CommitTier.String(),
		plainBar(row.CommitProgress),
		fmt.Sprintf("%dd", row.CommitDays),
	}
}

func NewTable(repos map[string]cache.RepoEntry, height int, focused bool) (table.Model, []string) {
	rows, paths := core.BuildRows(repos)
	repoWidth := core.RepoColumnWidth(rows)
	tableRows := make([]table.Row, 0, len(rows))
	for i, r := range rows {
		tableRows = append(tableRows, rowToTableRow(r, paths[i], repoWidth))
	}

	totalWidth := 0
	cols := Columns(repoWidth)
	for _, c := range cols {
		totalWidth += c.Width
		if c.Width > 0 {
			totalWidth += 2
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(tableRows),
		table.WithHeight(height),
		table.WithWidth(totalWidth),
		table.WithFocused(focused),
	)

	km := table.DefaultKeyMap()
	km.HalfPageUp = key.NewBinding(key.WithKeys("ctrl+u"))
	km.HalfPageDown = key.NewBinding(key.WithKeys("ctrl+d"))
	km.PageUp = key.NewBinding(key.WithKeys("pgup"))
	km.PageDown = key.NewBinding(key.WithKeys("pgdown"))
	km.GotoTop = key.NewBinding(key.WithKeys("home"))
	km.GotoBottom = key.NewBinding(key.WithKeys("end"))
	t.KeyMap = km

	t.SetStyles(TableStyles())
	return t, paths
}

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
		tbl.Row(
			r.Name,
			r.Tag,
			tier,
			colorBar(r.CommitTier, r.CommitProgress),
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
