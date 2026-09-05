package tui

import (
	"github.com/charmbracelet/x/ansi"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/core"
	"github.com/jgabor/gitfetch/internal/display"
	"github.com/mattn/go-runewidth"
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

func TableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Cell = s.Cell.Padding(0, 1, 0, 0)
	s.Header = s.Header.Padding(0, 1, 0, 0)
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

func NewTable(repos map[string]cache.RepoEntry, height int, focused bool, termWidth int) (table.Model, []string) {
	rows, paths := core.BuildRows(repos)

	layout := display.NewLayout(rows, termWidth)
	// Bubbles also pads the final column; reserve that cell when needed.
	wanted := len(layout.Columns)
	for _, c := range layout.Columns {
		wanted += c.Width
	}
	if termWidth > 0 && wanted > termWidth {
		layout.Columns[0].Width = max(1, layout.Columns[0].Width-1)
	}
	cols := []table.Column{{Title: "", Width: 0}}
	totalWidth := 0
	for _, c := range layout.Columns {
		cols = append(cols, table.Column{Title: c.Title, Width: c.Width})
		totalWidth += c.Width + 1
	}
	tableRows := make([]table.Row, 0, len(rows))
	for i, r := range rows {
		cells := layout.Cells(r)
		for j := range cells {
			cells[j] = ansi.Truncate(strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(cells[j]), layout.Columns[j].Width, "…")
		}
		tableRows = append(tableRows, append(table.Row{paths[i]}, cells...))
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(tableRows),
		table.WithHeight(max(2, height)),
		table.WithWidth(totalWidth-1),
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
