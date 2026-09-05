package core

import (
	"path/filepath"
	"sort"
	"time"

	"charm.land/lipgloss/v2"
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
	HasCommit      bool
	WeeklyCommits  []int
	ScannedAt      time.Time
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
		row := RepoRow{Name: filepath.Base(name), Tag: entry.LastTag, ScannedAt: entry.ScannedAt, WeeklyCommits: alignedActivity(entry, time.Now())}
		if entry.Error != "" {
			row.Error = entry.Error
		} else {
			if entry.LastCommitDate != nil && !entry.LastCommitDate.IsZero() {
				row.HasCommit = true
				row.CommitDays = decay.AgeDays(*entry.LastCommitDate)
				row.CommitTier = decay.ClassifyByDays(row.CommitDays)
				row.CommitProgress = decay.OverallProgress(row.CommitDays)
			}
			if entry.LastTagDate != nil && !entry.LastTagDate.IsZero() {
				row.TagDays = decay.AgeDays(*entry.LastTagDate)
				row.TagTier = decay.ClassifyByDays(row.TagDays)
				row.TagProgress = decay.OverallProgress(row.TagDays)
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

// alignedActivity puts cached counts on the current eight completed UTC weeks.
// Negative counts mean that a week has not been scanned, not zero activity.
func alignedActivity(entry cache.RepoEntry, now time.Time) []int {
	if len(entry.WeeklyCommits) != 8 || entry.ScannedAt.IsZero() {
		return nil
	}
	shift := int(cache.WeekStart(now).Sub(cache.WeekStart(entry.ScannedAt)).Hours() / (7 * 24))
	out := make([]int, 8)
	for i := range out {
		source := i + shift
		out[i] = -1
		if source >= 0 && source < 8 {
			out[i] = entry.WeeklyCommits[source]
		}
	}
	return out
}
