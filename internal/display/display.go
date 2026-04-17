package display

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/decay"
)

const (
	colRepo    = 25
	colVersion = 10
	colTier    = 8
	barWidth   = 15
	colDecay   = barWidth
	colAge     = 5

	barFilled = "#"
	barEmpty  = "-"
)

func plainBar(tier decay.Tier, progress float64) string {
	filled := int(float64(barWidth) * progress)
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}
	return strings.Repeat(barFilled, filled) + strings.Repeat(barEmpty, barWidth-filled)
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

type RepoRow struct {
	Name           string
	Tag            string
	CommitTier     decay.Tier
	CommitDays     int
	CommitProgress float64
	TagTier        decay.Tier
	TagDays        int
	TagProgress    float64
	HasTag         bool
	Error          string
}

func BuildRows(repos map[string]cache.RepoEntry) ([]RepoRow, []string) {
	sortedPaths := make([]string, 0, len(repos))
	for p := range repos {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Slice(sortedPaths, func(i, j int) bool {
		return filepath.Base(sortedPaths[i]) < filepath.Base(sortedPaths[j])
	})

	rows := make([]RepoRow, 0, len(repos))
	paths := make([]string, 0, len(repos))
	for _, name := range sortedPaths {
		entry := repos[name]
		row := RepoRow{Name: filepath.Base(name), Tag: entry.LastTag}
		if entry.Error != "" {
			row.Error = entry.Error
			rows = append(rows, row)
			paths = append(paths, name)
			continue
		}
		if entry.LastCommitDate != nil {
			row.CommitDays = decay.AgeDays(*entry.LastCommitDate)
			row.CommitTier = decay.ClassifyByDays(row.CommitDays)
			row.CommitProgress = decay.TierProgress(row.CommitDays)
		}
		if entry.LastTagDate != nil {
			row.TagDays = decay.AgeDays(*entry.LastTagDate)
			row.TagTier = decay.ClassifyByDays(row.TagDays)
			row.TagProgress = decay.TierProgress(row.TagDays)
			row.HasTag = true
		}
		rows = append(rows, row)
		paths = append(paths, name)
	}
	return rows, paths
}

func Columns() []table.Column {
	return []table.Column{
		{Title: "", Width: 0},
		{Title: "Repo", Width: colRepo},
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

func rowToTableRow(row RepoRow, fullPath string) table.Row {
	if row.Error != "" {
		return table.Row{
			fullPath,
			truncatePlain(row.Name, colRepo),
			"",
			"error",
			truncatePlain(row.Error, colDecay),
			"",
		}
	}

	bar := plainBar(row.CommitTier, row.CommitProgress)

	return table.Row{
		fullPath,
		truncatePlain(row.Name, colRepo),
		truncatePlain(row.Tag, colVersion),
		row.CommitTier.String(),
		bar,
		fmt.Sprintf("%dd", row.CommitDays),
	}
}

func NewTable(repos map[string]cache.RepoEntry, height int, focused bool) (table.Model, []string) {
	rows, paths := BuildRows(repos)
	tableRows := make([]table.Row, 0, len(rows))
	for i, r := range rows {
		tableRows = append(tableRows, rowToTableRow(r, paths[i]))
	}

	totalWidth := 0
	cols := Columns()
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

func FormatDashboard(repos map[string]cache.RepoEntry) string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	b.WriteString(headerStyle.Render("gitfetch -- repo decay tracker"))
	b.WriteString("\n\n")

	if len(repos) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("No repos tracked. Run `gitfetch refresh` to scan."))
		b.WriteString("\n")
		return b.String()
	}

	t, _ := NewTable(repos, len(repos)+2, false)
	styles := TableStyles()
	styles.Selected = styles.Cell
	t.SetStyles(styles)
	b.WriteString(t.View())

	b.WriteString("\n")
	fresh := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("# fresh")
	stale := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("# stale")
	decayed := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("# decayed")
	dead := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("# dead")
	b.WriteString(fmt.Sprintf("Legend: %s  %s  %s  %s", fresh, stale, decayed, dead))
	b.WriteString("\n")

	return b.String()
}
