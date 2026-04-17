package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
	"github.com/jgabor/gitfetch/internal/display"
	gitscanner "github.com/jgabor/gitfetch/internal/git"
)

type mode int

const (
	modeNormal mode = iota
	modeAdding
	modeDiscovering
)

type scanDoneMsg struct {
	results []gitscanner.ScanResult
}

type discoveredRepo struct {
	path     string
	selected bool
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

	discovered          []discoveredRepo
	discoverCursor      int
	discoverAllSelected bool

	height      int
	width       int
	inputBuffer string
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

func filteredCache(cfgRepos []string, c *cache.Cache) map[string]cache.RepoEntry {
	return cache.FilterByRepos(c, cfgRepos)
}

func buildTableModel(repos map[string]cache.RepoEntry, height int) table.Model {
	t, _ := display.NewTable(repos, height, true)
	s := display.TableStyles()
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62"))
	t.SetStyles(s)
	return t
}

func (m *model) rebuildTable() {
	fc := filteredCache(m.repos, m.c)
	m.table = buildTableModel(fc, m.height-6)
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		if m.mode == modeNormal {
			m.rebuildTable()
		}
		return m, nil
	case tea.KeyMsg:
		switch m.mode {
		case modeAdding:
			return handleAdding(m, msg)
		case modeDiscovering:
			return handleDiscovering(m, msg)
		default:
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			handleNormalKeys(&m, msg)
			if m.quitting {
				return m, tea.Quit
			}
			return m, cmd
		}
	case scanDoneMsg:
		return handleScanDone(m, msg)
	}
	return m, nil
}

func handleNormalKeys(m *model, msg tea.KeyMsg) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
	case "r":
		m.mode = modeNormal
	case "a":
		m.mode = modeAdding
		m.statusMsg = ""
		m.inputBuffer = ""
	case "d", "x":
		removeRepo(m)
	}
}

func handleAdding(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
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

		for _, existing := range m.repos {
			if existing == path {
				m.mode = modeNormal
				m.statusMsg = fmt.Sprintf("already tracked: %s", path)
				m.rebuildTable()
				return m, nil
			}
		}
		m.repos = append(m.repos, path)
		m.cfg.Repos = m.repos
		if err := config.Save(m.cfgPath, m.cfg); err != nil {
			m.err = err
			m.mode = modeNormal
			return m, nil
		}
		m.mode = modeNormal
		m.statusMsg = fmt.Sprintf("added: %s", path)
		m.rebuildTable()
		return m, nil
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
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
		return m, nil
	}
}

func handleDiscovering(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case " ":
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
			m.err = err
			m.mode = modeNormal
			m.discovered = nil
			m.rebuildTable()
			return m, nil
		}
		m.mode = modeNormal
		m.discovered = nil
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
		return scanDoneMsg{results: gitscanner.ScanAll(repos, gitscanner.ScanOptions{})}
	}
}

func handleScanDone(m model, msg scanDoneMsg) (tea.Model, tea.Cmd) {
	for _, r := range msg.results {
		m.c.Repos[r.RepoPath] = cache.RepoEntry{
			LastCommitDate: r.LastCommitDate,
			LastTagDate:    r.LastTagDate,
			LastTag:        r.LastTag,
			Error:          r.Error,
			ScannedAt:      r.ScannedAt,
		}
	}
	if err := cache.Save(m.cachePath, m.c); err != nil {
		m.err = err
		return m, nil
	}
	ok, fail := 0, 0
	for _, r := range msg.results {
		if r.Error != "" {
			fail++
		} else {
			ok++
		}
	}
	m.statusMsg = fmt.Sprintf("refreshed %d repos: %d ok, %d failed", len(msg.results), ok, fail)
	m.rebuildTable()
	return m, nil
}

func removeRepo(m *model) {
	selected := m.table.SelectedRow()
	if len(selected) == 0 {
		return
	}
	removed := selected[0]
	if removed == "" {
		return
	}
	for i, r := range m.repos {
		if r == removed {
			m.repos = append(m.repos[:i], m.repos[i+1:]...)
			break
		}
	}
	m.cfg.Repos = m.repos
	if err := config.Save(m.cfgPath, m.cfg); err != nil {
		m.err = err
		return
	}
	m.statusMsg = fmt.Sprintf("removed: %s", filepath.Base(removed))
	m.rebuildTable()
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
		start = total - availableLines
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render("gitfetch — repo decay tracker"))
	b.WriteString("\n\n")

	if m.mode == modeDiscovering {
		return viewDiscovering(m, &b)
	}

	if len(m.repos) == 0 {
		b.WriteString(helpStyle.Render("No repos tracked. Press 'a' to add a path or directory."))
		b.WriteString("\n")
	} else {
		b.WriteString(m.table.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if m.mode == modeAdding {
		b.WriteString(promptStyle.Render("Add path (directory or repo): "))
		b.WriteString(m.inputBuffer)
		b.WriteString(lipgloss.NewStyle().Blink(true).Render("█"))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Enter to confirm · Directories auto-discover repos · Esc to cancel"))
		b.WriteString("\n")
	} else {
		b.WriteString(helpStyle.Render("↑/k up · ↓/k down · r refresh · a add · d/x remove · q quit"))
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

	return b.String()
}

func viewDiscovering(m model, b *strings.Builder) string {
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
	available := m.height - chromeLines
	if available < 3 {
		available = 3
	}

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
		maxPathWidth := m.width - len(name) - 8
		pathStr := d.path
		if maxPathWidth > 10 && len(pathStr) > maxPathWidth {
			pathStr = "..." + pathStr[len(pathStr)-maxPathWidth+3:]
		}

		b.WriteString(cursor)
		b.WriteString(style.Render(fmt.Sprintf("%s ", check)))
		b.WriteString(fmt.Sprintf("%-20s", name))
		b.WriteString(helpStyle.Render(fmt.Sprintf("  %s", pathStr)))
		b.WriteString("\n")
	}

	if end < total {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  ... %d more below", total-end)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Space toggle · Enter confirm · a toggle all · Esc cancel"))
	b.WriteString("\n")

	if m.statusMsg != "" {
		b.WriteString(statusStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}

	return b.String()
}

func Run(cfg *config.Config, cfgPath string, c *cache.Cache, cachePath string) error {
	m := NewModel(cfg, cfgPath, c, cachePath)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
