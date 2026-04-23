package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/core"
	"github.com/jgabor/gitfetch/internal/theme"
	"github.com/mattn/go-runewidth"
)

const (
	colRepoMin      = 4
	colVersion      = 10
	colTier         = 8
	defaultBarWidth = 30
	colAge          = 5
)

func truncatePlain(s string, maxLen int) string {
	if runewidth.StringWidth(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		truncated := runewidth.Truncate(s, maxLen-3, "")
		return truncated + "..."
	}
	return runewidth.Truncate(s, maxLen, "")
}

func ComputeBarWidth(termWidth, repoWidth int) int {
	if termWidth <= 0 {
		return defaultBarWidth
	}
	barW := termWidth - repoWidth - colVersion - colTier - colAge - 10
	if barW < 1 {
		barW = 1
	}
	return barW
}

func Columns(repoWidth, barW int) []table.Column {
	return []table.Column{
		{Title: "", Width: 0},
		{Title: "Repo", Width: repoWidth},
		{Title: "Version", Width: colVersion},
		{Title: "Tier", Width: colTier},
		{Title: "Commit", Width: barW},
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

func rowToTableRow(row core.RepoRow, fullPath string, repoWidth, barW int) table.Row {
	if row.Error != "" {
		return table.Row{
			fullPath,
			truncatePlain(row.Name, repoWidth),
			"",
			"error",
			truncatePlain(row.Error, barW),
			"",
		}
	}

	tagCell := truncatePlain(row.Tag, colVersion)
	if row.HasTag {
		tagColor := lipgloss.Color(row.TagTier.Color())
		tagCell = lipgloss.NewStyle().Foreground(tagColor).Render(tagCell)
	}

	return table.Row{
		fullPath,
		truncatePlain(row.Name, repoWidth),
		tagCell,
		row.CommitTier.String(),
		theme.GradientBar(row.CommitDays, row.CommitProgress, barW),
		fmt.Sprintf("%dd", row.CommitDays),
	}
}

func NewTable(repos map[string]cache.RepoEntry, height int, focused bool, termWidth int) (table.Model, []string) {
	rows, paths := core.BuildRows(repos)
	repoWidth := core.RepoColumnWidth(rows)
	barW := ComputeBarWidth(termWidth, repoWidth)
	tableRows := make([]table.Row, 0, len(rows))
	for i, r := range rows {
		tableRows = append(tableRows, rowToTableRow(r, paths[i], repoWidth, barW))
	}

	totalWidth := 0
	cols := Columns(repoWidth, barW)
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
