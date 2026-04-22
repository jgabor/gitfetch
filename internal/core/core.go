package core

import (
	"path/filepath"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/decay"
)

const colRepoMin = 4

func RepoColumnWidth(rows []RepoRow) int {
	w := colRepoMin
	for _, r := range rows {
		if n := lipgloss.Width(r.Name); n > w {
			w = n
		}
	}
	return w
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
	type rowPath struct {
		row  RepoRow
		path string
	}

	all := make([]rowPath, 0, len(repos))
	for name, entry := range repos {
		row := RepoRow{Name: filepath.Base(name), Tag: entry.LastTag}
		if entry.Error != "" {
			row.Error = entry.Error
		} else {
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
		}
		all = append(all, rowPath{row: row, path: name})
	}

	sort.Slice(all, func(i, j int) bool {
		a, b := all[i].row, all[j].row
		aErr, bErr := a.Error != "", b.Error != ""
		if aErr != bErr {
			return !aErr
		}
		if aErr && bErr {
			return a.Name < b.Name
		}
		if a.CommitTier != b.CommitTier {
			return a.CommitTier > b.CommitTier
		}
		if a.CommitDays != b.CommitDays {
			return a.CommitDays > b.CommitDays
		}
		return a.Name < b.Name
	})

	rows := make([]RepoRow, len(all))
	paths := make([]string, len(all))
	for i, rp := range all {
		rows[i] = rp.row
		paths[i] = rp.path
	}
	return rows, paths
}
