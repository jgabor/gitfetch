package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/core"
	"github.com/jgabor/gitfetch/internal/theme"
)

const (
	colRepoMin = 4
	colVersion = 10
	colTier    = 8
	barWidth   = 15
	colDecay   = barWidth
	colAge     = 5
)

func plainBar(pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := min(int(float64(barWidth)*pct+0.5), barWidth)
	return strings.Repeat(theme.BarFilled, filled) + strings.Repeat(theme.BarEmpty, barWidth-filled)
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
		theme.GradientBar(row.CommitTier, row.CommitProgress),
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
