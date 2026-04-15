package display

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/decay"
)

const (
	barWidth  = 20
	barFilled = "█"
	barEmpty  = "░"
)

func styledBar(tier decay.Tier, progress float64) string {
	filled := int(float64(barWidth) * progress)
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat(barFilled, filled) + strings.Repeat(barEmpty, barWidth-filled)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(tier.Color())).Render(bar)
}

func tierLabel(tier decay.Tier) string {
	label := fmt.Sprintf("%-7s", tier.Label())
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(tier.Color())).
		Bold(true).
		Render(label)
}

type RepoRow struct {
	Name           string
	CommitTier     decay.Tier
	CommitDays     int
	CommitProgress float64
	TagTier        decay.Tier
	TagDays        int
	TagProgress    float64
	HasTag         bool
	Error          string
}

func BuildRows(repos map[string]cache.RepoEntry) []RepoRow {
	rows := make([]RepoRow, 0, len(repos))
	for name, entry := range repos {
		row := RepoRow{Name: filepath.Base(name)}
		if entry.Error != "" {
			row.Error = entry.Error
			rows = append(rows, row)
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
	}
	return rows
}

func FormatRow(row RepoRow) string {
	if row.Error != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		return fmt.Sprintf("  %-20s  %s", row.Name, errStyle.Render("error: "+row.Error))
	}

	commitLabel := tierLabel(row.CommitTier)
	commitBar := styledBar(row.CommitTier, row.CommitProgress)
	commitAge := fmt.Sprintf("%3dd", row.CommitDays)
	commitLine := fmt.Sprintf("  %-20s  %s %s %s", row.Name, commitLabel, commitBar, commitAge)

	if row.HasTag {
		tagLabel := tierLabel(row.TagTier)
		tagBar := styledBar(row.TagTier, row.TagProgress)
		tagAge := fmt.Sprintf("%3dd", row.TagDays)
		commitLine += fmt.Sprintf("\n%23s%s %s %s", "", tagLabel, tagBar, tagAge)
	}

	return commitLine
}

func FormatDashboard(repos map[string]cache.RepoEntry) string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	b.WriteString(headerStyle.Render("gitfetch — repo decay tracker"))
	b.WriteString("\n\n")

	rows := BuildRows(repos)
	if len(rows) == 0 {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		b.WriteString(hintStyle.Render("No repos tracked. Run `gitfetch refresh` to scan."))
		b.WriteString("\n")
		return b.String()
	}

	for _, row := range rows {
		b.WriteString(FormatRow(row))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	fresh := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("■ fresh")
	stale := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("■ stale")
	decayed := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("■ decayed")
	dead := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("■ dead")
	b.WriteString(fmt.Sprintf("Legend: %s  %s  %s  %s", fresh, stale, decayed, dead))
	b.WriteString("\n")

	return b.String()
}
