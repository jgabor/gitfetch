package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
	"github.com/jgabor/gitfetch/internal/decay"
	"github.com/jgabor/gitfetch/internal/display"
	gitscanner "github.com/jgabor/gitfetch/internal/git"
)

type mode int

const (
	modeNormal mode = iota
	modeAdding
)

type scanDoneMsg struct {
	results []gitscanner.ScanResult
}

type model struct {
	cfg       *config.Config
	cfgPath   string
	c         *cache.Cache
	cachePath string

	repos  []string
	cursor int

	mode      mode
	statusMsg string
	quitting  bool

	err error
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	inputBuffer string
)

func NewModel(cfg *config.Config, cfgPath string, c *cache.Cache, cachePath string) model {
	return model{
		cfg:       cfg,
		cfgPath:   cfgPath,
		c:         c,
		cachePath: cachePath,
		repos:     cfg.Repos,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case modeAdding:
			return handleAdding(m, msg)
		default:
			return handleNormal(m, msg)
		}
	case scanDoneMsg:
		return handleScanDone(m, msg)
	}
	return m, nil
}

func handleNormal(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.repos)-1 {
			m.cursor++
		}
	case "r":
		return m, tea.Batch(startScan(m), tea.EnterAltScreen)
	case "a":
		m.mode = modeAdding
		m.statusMsg = ""
		inputBuffer = ""
		return m, nil
	case "d", "x":
		return removeRepo(m)
	}
	return m, nil
}

func handleAdding(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		path := strings.TrimSpace(inputBuffer)
		inputBuffer = ""
		if path == "" {
			m.mode = modeNormal
			return m, nil
		}
		for _, existing := range m.repos {
			if existing == path {
				m.mode = modeNormal
				m.statusMsg = fmt.Sprintf("already tracked: %s", path)
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
		m.cursor = len(m.repos) - 1
		m.mode = modeNormal
		m.statusMsg = fmt.Sprintf("added: %s", path)
		return m, nil
	case "esc":
		m.mode = modeNormal
		inputBuffer = ""
		return m, nil
	case "backspace":
		if len(inputBuffer) > 0 {
			inputBuffer = inputBuffer[:len(inputBuffer)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			inputBuffer += msg.String()
		}
		return m, nil
	}
}

func startScan(m model) tea.Cmd {
	repos := make([]string, len(m.repos))
	copy(repos, m.repos)
	return func() tea.Msg {
		return scanDoneMsg{results: gitscanner.ScanAll(repos)}
	}
}

func handleScanDone(m model, msg scanDoneMsg) (tea.Model, tea.Cmd) {
	for _, r := range msg.results {
		m.c.Repos[r.RepoPath] = cache.RepoEntry{
			LastCommitDate: r.LastCommitDate,
			LastTagDate:    r.LastTagDate,
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
	return m, nil
}

func removeRepo(m model) (tea.Model, tea.Cmd) {
	if len(m.repos) == 0 {
		return m, nil
	}
	removed := m.repos[m.cursor]
	m.repos = append(m.repos[:m.cursor], m.repos[m.cursor+1:]...)
	m.cfg.Repos = m.repos
	if err := config.Save(m.cfgPath, m.cfg); err != nil {
		m.err = err
		return m, nil
	}
	if m.cursor >= len(m.repos) && m.cursor > 0 {
		m.cursor--
	}
	m.statusMsg = fmt.Sprintf("removed: %s", removed)
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(headerStyle.Render("gitfetch — repo decay tracker"))
	b.WriteString("\n\n")

	if len(m.repos) == 0 {
		b.WriteString(helpStyle.Render("No repos tracked. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		for i, repoPath := range m.repos {
			entry, exists := m.c.Repos[repoPath]
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("▸ ")
			}

			if !exists {
				b.WriteString(fmt.Sprintf("%s%-30s  %s\n", cursor, filepath.Base(repoPath),
					helpStyle.Render("not scanned yet")))
				continue
			}

			row := display.BuildRows(map[string]cache.RepoEntry{repoPath: entry})
			if len(row) == 0 {
				continue
			}

			formatted := display.FormatRow(row[0])
			lines := strings.Split(formatted, "\n")
			for li, line := range lines {
				if li == 0 {
					b.WriteString(cursor)
				} else {
					b.WriteString("  ")
				}
				b.WriteString(line)
				b.WriteString("\n")
			}

			if exists && entry.Error == "" {
				days := decay.AgeDays(*entry.LastCommitDate)
				tier := decay.ClassifyByDays(days)
				b.WriteString(fmt.Sprintf("    %s\n", helpStyle.Render(
					fmt.Sprintf("path: %s  tier: %s", repoPath, tier.Label()))))
			} else if exists && entry.Error != "" {
				b.WriteString(fmt.Sprintf("    %s\n", helpStyle.Render(
					fmt.Sprintf("path: %s", repoPath))))
			}
		}
	}

	b.WriteString("\n")

	if m.mode == modeAdding {
		b.WriteString(promptStyle.Render("Add repo path: "))
		b.WriteString(inputBuffer)
		b.WriteString(lipgloss.NewStyle().Blink(true).Render("█"))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Enter to confirm · Esc to cancel"))
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

func Run(cfg *config.Config, cfgPath string, c *cache.Cache, cachePath string) error {
	m := NewModel(cfg, cfgPath, c, cachePath)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
