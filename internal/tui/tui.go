package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
	"github.com/jgabor/gitfetch/internal/core"
	"github.com/jgabor/gitfetch/internal/display"
	gitscanner "github.com/jgabor/gitfetch/internal/git"
	"github.com/mattn/go-runewidth"
)

type mode int

const (
	modeNormal mode = iota
	modeAdding
	modeDiscovering
	modeConfirming
)

type scanProgressMsg struct {
	done      int
	total     int
	result    gitscanner.ScanResult
	remaining []string
}

type discoveredRepo struct {
	path     string
	selected bool
}

type clearErrorMsg struct {
	seq int
}

func startErrorAutoClear(m *model, err error) tea.Cmd {
	m.err = err
	m.errSeq++
	seq := m.errSeq
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return clearErrorMsg{seq: seq}
	})
}

type model struct {
	cfg       *config.Config
	cfgPath   string
	c         *cache.Cache
	cachePath string

	repos []string
	table table.Model

	mode      mode
	statusMsg string
	quitting  bool
	err       error
	errSeq    int

	confirmTarget string

	discovered          []discoveredRepo
	discoverCursor      int
	discoverAllSelected bool

	height      int
	width       int
	inputBuffer string
	scanning    bool

	scanDone    int
	scanTotal   int
	scanResults []gitscanner.ScanResult
}

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	promptStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	checkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	uncheckStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func buildTableModel(repos map[string]cache.RepoEntry, height, termWidth int) table.Model {
	t, _ := NewTable(repos, height, true, termWidth)
	s := TableStyles()
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62"))
	t.SetStyles(s)
	return t
}

func (m *model) rebuildTable() {
	fc := cache.FilterByRepos(m.c, m.repos)
	m.table = buildTableModel(fc, m.height-12, m.width)
	m.fitTable()
}

func NewModel(cfg *config.Config, cfgPath string, c *cache.Cache, cachePath string) model {
	m := model{
		cfg:       cfg,
		cfgPath:   cfgPath,
		c:         c,
		cachePath: cachePath,
		repos:     cfg.Repos,
		height:    24,
		width:     80,
	}
	m.rebuildTable()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.update(msg)
	next := updated.(model)
	next.fitTable()
	return next, cmd
}

func (m *model) fitTable() {
	height := m.height - lipgloss.Height(m.selectedDetails()) - 8
	if m.statusMsg != "" {
		height--
	}
	if m.err != nil {
		height--
	}
	if m.mode == modeAdding || m.mode == modeConfirming {
		height--
	}
	m.table.SetHeight(max(3, height))
}

func (m model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		if m.mode == modeNormal {
			m.rebuildTable()
		}
		return m, nil
	case tea.KeyPressMsg:
		switch m.mode {
		case modeAdding:
			return handleAdding(m, msg)
		case modeDiscovering:
			return handleDiscovering(m, msg)
		case modeConfirming:
			return handleConfirming(m, msg)
		default:
			var tableCmd tea.Cmd
			m.table, tableCmd = m.table.Update(msg)
			keyCmd := handleNormalKeys(&m, msg)
			if m.quitting {
				return m, tea.Quit
			}
			return m, tea.Batch(tableCmd, keyCmd)
		}
	case scanProgressMsg:
		return handleScanProgress(m, msg)
	case clearErrorMsg:
		if m.errSeq == msg.seq {
			m.err = nil
		}
		return m, nil
	}
	return m, nil
}

func handleNormalKeys(m *model, msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
	case "r":
		if m.scanning || len(m.repos) == 0 {
			return nil
		}
		m.scanning = true
		m.statusMsg = "scanning…"
		m.err = nil
		return startScan(*m)
	case "a":
		m.mode = modeAdding
		m.statusMsg = ""
		m.inputBuffer = ""
	case "d", "x":
		selected := m.table.SelectedRow()
		if len(selected) > 0 && selected[0] != "" {
			m.mode = modeConfirming
			m.confirmTarget = selected[0]
			m.statusMsg = ""
		}
	}
	return nil
}

func handleAdding(m model, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m, cmd := commitNewRepo(m)
		return m, cmd
	case "esc":
		m.mode = modeNormal
		m.inputBuffer = ""
		m.rebuildTable()
		return m, nil
	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
		return m, nil
	default:
		if len(msg.Text) == 1 {
			m.inputBuffer += msg.Text
		}
		return m, nil
	}
}

func handleConfirming(m model, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		target := m.confirmTarget
		m.mode = modeNormal
		m.confirmTarget = ""
		for i, r := range m.repos {
			if r == target {
				m.repos = append(m.repos[:i], m.repos[i+1:]...)
				break
			}
		}
		m.cfg.Repos = m.repos
		if err := config.Save(m.cfgPath, m.cfg); err != nil {
			return m, startErrorAutoClear(&m, err)
		}
		m.err = nil
		m.statusMsg = fmt.Sprintf("removed: %s", filepath.Base(target))
		m.rebuildTable()
		return m, nil
	case "n", "esc", "q":
		m.mode = modeNormal
		m.confirmTarget = ""
		return m, nil
	default:
		return m, nil
	}
}

func commitNewRepo(m model) (model, tea.Cmd) {
	path := strings.TrimSpace(m.inputBuffer)
	m.inputBuffer = ""
	if path == "" {
		m.mode = modeNormal
		m.rebuildTable()
		return m, nil
	}
	path = config.ExpandPath(path)

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		found, derr := gitscanner.DiscoverRepos(path)
		if derr != nil {
			found = []string{path}
		}
		if len(found) > 1 || (len(found) == 1 && found[0] != path) {
			m.discovered = make([]discoveredRepo, len(found))
			for i, p := range found {
				m.discovered[i] = discoveredRepo{path: p, selected: true}
			}
			m.discoverCursor = 0
			m.discoverAllSelected = true
			m.mode = modeDiscovering
			m.statusMsg = ""
			return m, nil
		}
	}

	if slices.Contains(m.repos, path) {
		m.statusMsg = fmt.Sprintf("already tracked: %s", path)
		return m, nil
	}
	m.repos = append(m.repos, path)
	m.cfg.Repos = m.repos
	if err := config.Save(m.cfgPath, m.cfg); err != nil {
		m.mode = modeNormal
		return m, startErrorAutoClear(&m, err)
	}
	m.mode = modeNormal
	m.err = nil
	m.statusMsg = fmt.Sprintf("added: %s", path)
	m.rebuildTable()
	return m, nil
}

func handleDiscovering(m model, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	total := len(m.discovered)
	switch msg.String() {
	case "up", "k":
		if m.discoverCursor > 0 {
			m.discoverCursor--
		}
	case "down", "j":
		if m.discoverCursor < total-1 {
			m.discoverCursor++
		}
	case "space":
		if total > 0 {
			m.discovered[m.discoverCursor].selected = !m.discovered[m.discoverCursor].selected
		}
	case "a":
		m.discoverAllSelected = !m.discoverAllSelected
		for i := range m.discovered {
			m.discovered[i].selected = m.discoverAllSelected
		}
	case "enter":
		var added []string
		tracked := make(map[string]bool, len(m.repos))
		for _, r := range m.repos {
			tracked[r] = true
		}
		for _, d := range m.discovered {
			if d.selected && !tracked[d.path] {
				m.repos = append(m.repos, d.path)
				added = append(added, filepath.Base(d.path))
			}
		}
		m.cfg.Repos = m.repos
		if err := config.Save(m.cfgPath, m.cfg); err != nil {
			m.mode = modeNormal
			m.discovered = nil
			m.rebuildTable()
			return m, startErrorAutoClear(&m, err)
		}
		m.mode = modeNormal
		m.discovered = nil
		m.err = nil
		if len(added) > 0 {
			m.statusMsg = fmt.Sprintf("added %d repos: %s", len(added), strings.Join(added, ", "))
		} else {
			m.statusMsg = "no repos selected"
		}
		m.rebuildTable()
		return m, nil
	case "esc":
		m.mode = modeNormal
		m.discovered = nil
		m.statusMsg = "cancelled"
		m.rebuildTable()
		return m, nil
	}
	return m, nil
}

func startScan(m model) tea.Cmd {
	repos := make([]string, len(m.repos))
	copy(repos, m.repos)
	return func() tea.Msg {
		resolved := gitscanner.ResolveRepoPaths(repos)
		if len(resolved) == 0 {
			return scanProgressMsg{done: 0, total: 0}
		}
		result := gitscanner.ScanRepo(resolved[0], gitscanner.ScanOptions{})
		return scanProgressMsg{
			done:      1,
			total:     len(resolved),
			result:    result,
			remaining: resolved[1:],
		}
	}
}

func scanNextRepo(done, total int, remaining []string) tea.Cmd {
	if len(remaining) == 0 {
		return nil
	}
	repo := remaining[0]
	rest := make([]string, len(remaining)-1)
	copy(rest, remaining[1:])
	return func() tea.Msg {
		result := gitscanner.ScanRepo(repo, gitscanner.ScanOptions{})
		return scanProgressMsg{
			done:      done + 1,
			total:     total,
			result:    result,
			remaining: rest,
		}
	}
}

func handleScanProgress(m model, msg scanProgressMsg) (tea.Model, tea.Cmd) {
	if msg.total == 0 {
		m.scanning = false
		m.statusMsg = "no repos to scan"
		return m, nil
	}
	m.scanDone = msg.done
	m.scanTotal = msg.total
	m.scanResults = append(m.scanResults, msg.result)
	m.statusMsg = fmt.Sprintf("scanning %d/%d…", msg.done, msg.total)
	if len(msg.remaining) == 0 {
		return finalizeScan(m)
	}
	return m, scanNextRepo(msg.done, msg.total, msg.remaining)
}

func finalizeScan(m model) (tea.Model, tea.Cmd) {
	m.scanning = false
	for _, r := range m.scanResults {
		m.c.Repos[r.RepoPath] = cache.RepoEntry{
			LastCommitDate: r.LastCommitDate,
			WeeklyCommits:  r.WeeklyCommits,
			LastTagDate:    r.LastTagDate,
			LastTag:        r.LastTag,
			Error:          r.Error,
			ScannedAt:      r.ScannedAt,
		}
	}
	if err := cache.Save(m.cachePath, m.c); err != nil {
		m.scanResults = nil
		m.scanDone = 0
		m.scanTotal = 0
		return m, startErrorAutoClear(&m, err)
	}
	ok, fail := 0, 0
	for _, r := range m.scanResults {
		if r.Error != "" {
			fail++
		} else {
			ok++
		}
	}
	m.statusMsg = fmt.Sprintf("refreshed %d repos: %d ok, %d failed", len(m.scanResults), ok, fail)
	m.scanResults = nil
	m.scanDone = 0
	m.scanTotal = 0
	m.err = nil
	m.rebuildTable()
	return m, nil
}

func visibleRange(cursor, total, availableLines int) (start, end int) {
	if total == 0 || availableLines <= 0 {
		return 0, 0
	}
	if total <= availableLines {
		return 0, total
	}
	half := availableLines / 2
	start = cursor - half
	end = start + availableLines
	if start < 0 {
		start = 0
		end = availableLines
	}
	if end > total {
		end = total
		start = max(total-availableLines, 0)
	}
	return start, end
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render("gitfetch - repo decay tracker"))
	b.WriteString("\n\n")

	if m.mode == modeDiscovering {
		return viewDiscovering(m, &b)
	}

	if len(m.repos) == 0 {
		b.WriteString(helpStyle.Render("No repos tracked. Press 'a' to add a path or directory."))
		b.WriteString("\n")
	} else if len(m.table.Rows()) == 0 {
		b.WriteString(helpStyle.Render("No scan data yet. Press 'r' to refresh."))
		b.WriteString("\n")
	} else {
		rows, _ := core.BuildRows(cache.FilterByRepos(m.c, m.repos))
		b.WriteString(ansi.Truncate(display.Summary(rows, time.Now()), max(1, m.width), "…"))
		b.WriteByte('\n')
		details := m.selectedDetails()
		b.WriteString(m.table.View())
		b.WriteByte('\n')
		b.WriteString(details)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if m.mode == modeAdding {
		b.WriteString(promptStyle.Render("Add path (directory or repo): "))
		b.WriteString(m.inputBuffer)
		b.WriteString(cursorStyle.Render("█"))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Enter to confirm · Directories auto-discover repos · Esc to cancel"))
		b.WriteString("\n")
	} else if m.mode == modeConfirming {
		b.WriteString(promptStyle.Render(fmt.Sprintf("Remove '%s'?", filepath.Base(m.confirmTarget))))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("y/Enter to confirm · n/Esc/q to cancel"))
		b.WriteString("\n")
	} else {
		b.WriteString(helpStyle.Render("↑/k up · ↓/j down · ctrl+u half-up · ctrl+d half-down · pgup/pgdown page · home/end jump · r refresh · a add · d/x remove · q quit"))
		b.WriteString("\n")
	}

	if m.statusMsg != "" {
		b.WriteString(statusStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(errStyle.Render(fmt.Sprintf("error: %v", m.err)))
		b.WriteString("\n")
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func viewDiscovering(m model, b *strings.Builder) tea.View {
	selected := 0
	for _, d := range m.discovered {
		if d.selected {
			selected++
		}
	}

	b.WriteString(promptStyle.Render(fmt.Sprintf("Discovered %d repos · %d/%d selected", len(m.discovered), selected, len(m.discovered))))
	b.WriteString("\n")

	total := len(m.discovered)
	chromeLines := 5
	available := max(m.height-chromeLines, 3)

	b.WriteString("\n")

	start, end := visibleRange(m.discoverCursor, total, available)

	if start > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ... %d more above", start)))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		d := m.discovered[i]
		cursor := "  "
		if i == m.discoverCursor {
			cursor = cursorStyle.Render("▸ ")
		}

		check := "☐"
		style := uncheckStyle
		if d.selected {
			check = "☑"
			style = checkStyle
		}

		name := filepath.Base(d.path)
		nameWidth := runewidth.StringWidth(name)
		maxPathWidth := m.width - nameWidth - 8
		pathStr := d.path
		if maxPathWidth > 10 && runewidth.StringWidth(pathStr) > maxPathWidth {
			pathStr = truncatePlain(pathStr, maxPathWidth)
		}

		b.WriteString(cursor)
		b.WriteString(style.Render(fmt.Sprintf("%s ", check)))
		padWidth := 20 - nameWidth
		if padWidth > 0 {
			b.WriteString(name + strings.Repeat(" ", padWidth))
		} else {
			b.WriteString(truncatePlain(name, 20))
		}
		b.WriteString(helpStyle.Render(fmt.Sprintf("  %s", pathStr)))
		b.WriteString("\n")
	}

	if end < total {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ... %d more below", total-end)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/k up · ↓/j down · Space toggle · a toggle all · Enter confirm · Esc cancel"))
	b.WriteString("\n")

	if m.statusMsg != "" {
		b.WriteString(statusStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func Run(cfg *config.Config, cfgPath string, c *cache.Cache, cachePath string) error {
	m := NewModel(cfg, cfgPath, c, cachePath)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func (m model) selectedDetails() string {
	selected := m.table.SelectedRow()
	if len(selected) == 0 {
		return ""
	}
	path := selected[0]
	entry, ok := m.c.Repos[path]
	if !ok {
		return ""
	}
	date := func(t *time.Time) string {
		if t == nil || t.IsZero() {
			return "—"
		}
		return t.UTC().Format(time.RFC3339)
	}
	text := path + "\nCommit: " + date(entry.LastCommitDate) + " · Release: " + date(entry.LastTagDate)
	if entry.LastTag != "" {
		text += " · " + entry.LastTag
	}
	text += "\nScanned: " + date(&entry.ScannedAt)
	if len(entry.WeeklyCommits) == 8 {
		end := cache.WeekStart(time.Now())
		text += " · Activity: " + end.AddDate(0, 0, -56).Format("2006-01-02") + " to " + end.Format("2006-01-02") + " (exclusive, UTC)"
	} else {
		text += " · Activity unavailable; refresh to collect"
	}
	if entry.Error != "" {
		text += "\nError: " + entry.Error
	}
	return ansi.Wrap(text, max(1, m.width), "")
}
