package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jgabor/gitfetch/internal/cache"
	"github.com/jgabor/gitfetch/internal/config"
	gitscanner "github.com/jgabor/gitfetch/internal/git"
	"github.com/mattn/go-runewidth"
)

func TestNewModel(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/a", "/b"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg/path", c, "/cache/path")
	if len(m.repos) != 2 {
		t.Errorf("repos = %d, want 2", len(m.repos))
	}
	if m.cfgPath != "/cfg/path" {
		t.Errorf("cfgPath = %q, want /cfg/path", m.cfgPath)
	}
	if m.cachePath != "/cache/path" {
		t.Errorf("cachePath = %q, want /cache/path", m.cachePath)
	}
}

func TestViewEmpty(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if len(view.Content) == 0 {
		t.Error("view should not be empty")
	}
}

func TestViewWithRepos(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/project"}}
	c := cache.New()
	commitDate := time.Now().UTC().Add(-10 * 24 * time.Hour)
	c.Repos["/home/user/project"] = cache.RepoEntry{
		LastCommitDate: &commitDate,
		ScannedAt:      time.Now().UTC(),
	}
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if len(view.Content) == 0 {
		t.Error("view should not be empty")
	}
}

func TestViewShowsHelp(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if len(view.Content) == 0 {
		t.Fatal("view is empty")
	}
}

func TestViewShowsStatus(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.statusMsg = "test status"
	view := m.View()
	if len(view.Content) == 0 {
		t.Error("view should not be empty")
	}
}

func TestRefreshKeyTriggersScan(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/some-repo"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache/cache.json")

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "r"})
	if cmd == nil {
		t.Fatal("expected non-nil cmd after 'r' keypress")
	}
	um := updated.(model)
	if !um.scanning {
		t.Error("expected scanning state active after 'r' keypress")
	}
	if um.statusMsg != "scanning…" {
		t.Errorf("status = %q, want %q", um.statusMsg, "scanning…")
	}
}

func TestRefreshKeyNoopWhenNoRepos(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache/cache.json")

	updated, _ := m.Update(tea.KeyPressMsg{Text: "r"})
	um := updated.(model)
	if um.scanning {
		t.Error("scanning should not start when no repos are tracked")
	}
	if um.statusMsg == "scanning…" {
		t.Error("status should not show scanning when no repos are tracked")
	}
}

func TestRefreshKeyIgnoredWhileScanning(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/p"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache/cache.json")
	m.scanning = true

	_, cmd := m.Update(tea.KeyPressMsg{Text: "r"})
	if cmd != nil {
		t.Error("expected nil cmd while already scanning (deduplicate)")
	}
}

func TestCommitNewRepoSingleValidPath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	newRepo := filepath.Join(dir, "fresh-repo")

	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, cfgPath, c, filepath.Join(dir, "cache.json"))
	m.mode = modeAdding
	m.inputBuffer = newRepo

	m, _ = commitNewRepo(m)
	if m.mode != modeNormal {
		t.Errorf("mode = %v, want modeNormal for valid single path", m.mode)
	}
	if len(m.repos) != 1 || m.repos[0] != newRepo {
		t.Errorf("repos = %v, want [%s]", m.repos, newRepo)
	}
	if !strings.Contains(m.statusMsg, "added:") {
		t.Errorf("statusMsg = %q, want 'added: …'", m.statusMsg)
	}
	// config persisted
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config not persisted: %v", err)
	}
	if len(loaded.Repos) != 1 || loaded.Repos[0] != newRepo {
		t.Errorf("persisted repos = %v, want [%s]", loaded.Repos, newRepo)
	}
}

func TestCommitNewRepoDuplicateStaysInAddMode(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	dup := filepath.Join(dir, "tracked")

	cfg := &config.Config{Repos: []string{dup}}
	c := cache.New()
	m := NewModel(cfg, cfgPath, c, filepath.Join(dir, "cache.json"))
	m.mode = modeAdding
	m.inputBuffer = dup

	m, _ = commitNewRepo(m)
	if m.mode != modeAdding {
		t.Errorf("mode = %v, want modeAdding for duplicate (user can retry)", m.mode)
	}
	if !strings.Contains(m.statusMsg, "already tracked") {
		t.Errorf("statusMsg = %q, want 'already tracked'", m.statusMsg)
	}
	if len(m.repos) != 1 {
		t.Errorf("repos len = %d, want 1 (no new entry)", len(m.repos))
	}
	if m.inputBuffer != "" {
		t.Errorf("inputBuffer = %q, want cleared", m.inputBuffer)
	}
}

func TestCommitNewRepoEmptyPathReturnsToNormal(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, filepath.Join(dir, "c.toml"), c, filepath.Join(dir, "cache.json"))
	m.mode = modeAdding
	m.inputBuffer = "   "

	m, _ = commitNewRepo(m)
	if m.mode != modeNormal {
		t.Errorf("mode = %v, want modeNormal for empty trimmed input", m.mode)
	}
	if len(m.repos) != 0 {
		t.Errorf("repos should stay empty, got %v", m.repos)
	}
}

func TestCommitNewRepoMultiRepoDirectoryEntersDiscovery(t *testing.T) {
	parent := t.TempDir()
	for _, name := range []string{"a", "b"} {
		sub := filepath.Join(parent, name)
		if err := os.MkdirAll(filepath.Join(sub, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, filepath.Join(parent, "c.toml"), c, filepath.Join(parent, "cache.json"))
	m.mode = modeAdding
	m.inputBuffer = parent

	m, _ = commitNewRepo(m)
	if m.mode != modeDiscovering {
		t.Errorf("mode = %v, want modeDiscovering for multi-repo directory", m.mode)
	}
	if len(m.discovered) != 2 {
		t.Errorf("discovered len = %d, want 2", len(m.discovered))
	}
	if len(m.repos) != 0 {
		t.Errorf("repos should stay empty until discovery confirms, got %v", m.repos)
	}
}

func TestScanProgressWritesCacheAndClearsScanning(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")

	cfg := &config.Config{Repos: []string{"/home/user/x"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, cachePath)
	m.scanning = true

	now := time.Now().UTC()
	msg := scanProgressMsg{
		done:      1,
		total:     1,
		result:    gitscanner.ScanResult{RepoPath: "/home/user/x", LastCommitDate: &now, ScannedAt: now},
		remaining: nil,
	}
	updated, _ := m.Update(msg)
	um := updated.(model)
	if um.scanning {
		t.Error("scanning should be false after final scanProgressMsg")
	}
	loaded, err := cache.Load(cachePath)
	if err != nil {
		t.Fatalf("cache not persisted: %v", err)
	}
	if _, ok := loaded.Repos["/home/user/x"]; !ok {
		t.Error("scan result not written to cache on disk")
	}
}

func TestViewQuitting(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.quitting = true
	view := m.View()
	if view.Content != "" {
		t.Errorf("quitting view should be empty, got %q", view.Content)
	}
}

func TestVisibleRangeBounds(t *testing.T) {
	cases := []struct {
		name               string
		cursor             int
		total              int
		available          int
		wantStart, wantEnd int
	}{
		{"zero total", 0, 0, 10, 0, 0},
		{"zero available", 0, 5, 0, 0, 0},
		{"single item", 0, 1, 10, 0, 1},
		{"total equals available", 3, 5, 5, 0, 5},
		{"cursor near start", 1, 20, 10, 0, 10},
		{"cursor mid", 10, 20, 10, 5, 15},
		{"cursor near end", 19, 20, 10, 10, 20},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end := visibleRange(tc.cursor, tc.total, tc.available)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Errorf("visibleRange(%d, %d, %d) = (%d,%d), want (%d,%d)",
					tc.cursor, tc.total, tc.available, start, end, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func makeDate(daysAgo int) time.Time {
	return time.Now().UTC().Add(-time.Duration(daysAgo) * 24 * time.Hour)
}

func TestNewTableCreatesColumns(t *testing.T) {
	commitDate := makeDate(10)
	repos := map[string]cache.RepoEntry{
		"/r1": {LastCommitDate: &commitDate, ScannedAt: time.Now().UTC()},
	}
	tm, _ := NewTable(repos, 10, false, 80)
	rows := tm.Rows()
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) == 0 || rows[0][0] != "/r1" {
		t.Errorf("expected hidden path column to contain /r1, got %v", rows[0])
	}
}

func TestFilteredCacheMatches(t *testing.T) {
	c := cache.New()
	c.Repos["/a"] = cache.RepoEntry{}
	c.Repos["/b"] = cache.RepoEntry{}
	got := cache.FilterByRepos(c, []string{"/a"})
	if len(got) != 1 {
		t.Errorf("expected 1 entry, got %d", len(got))
	}
	if _, ok := got["/a"]; !ok {
		t.Error("expected /a present")
	}
}

func TestFilteredCacheExcludesUnconfigured(t *testing.T) {
	c := cache.New()
	c.Repos["/a"] = cache.RepoEntry{}
	c.Repos["/rogue"] = cache.RepoEntry{}
	got := cache.FilterByRepos(c, []string{"/a"})
	if _, ok := got["/rogue"]; ok {
		t.Error("/rogue should not appear; FilterByRepos must exclude unconfigured paths")
	}
}

func TestRemoveRepoEntersConfirmingState(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	cfg := &config.Config{Repos: []string{"/first", "/second"}}
	c := cache.New()
	now := time.Now().UTC()
	c.Repos["/first"] = cache.RepoEntry{LastCommitDate: &now, ScannedAt: now}
	c.Repos["/second"] = cache.RepoEntry{LastCommitDate: &now, ScannedAt: now}
	m := NewModel(cfg, cfgPath, c, filepath.Join(dir, "cache.json"))

	updated, _ := m.Update(tea.KeyPressMsg{Text: "d"})
	um := updated.(model)

	if um.mode != modeConfirming {
		t.Fatalf("mode = %v, want modeConfirming after pressing d", um.mode)
	}
	if len(um.repos) != 2 {
		t.Fatalf("repos len = %d, want 2 (not deleted yet)", len(um.repos))
	}
	if um.confirmTarget != "/first" {
		t.Errorf("confirmTarget = %q, want /first", um.confirmTarget)
	}
}

func TestConfirmingYesRemovesRepo(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	cfg := &config.Config{Repos: []string{"/first", "/second"}}
	c := cache.New()
	now := time.Now().UTC()
	c.Repos["/first"] = cache.RepoEntry{LastCommitDate: &now, ScannedAt: now}
	c.Repos["/second"] = cache.RepoEntry{LastCommitDate: &now, ScannedAt: now}
	m := NewModel(cfg, cfgPath, c, filepath.Join(dir, "cache.json"))
	m.mode = modeConfirming
	m.confirmTarget = "/first"

	updated, _ := m.Update(tea.KeyPressMsg{Text: "y"})
	um := updated.(model)

	if um.mode != modeNormal {
		t.Errorf("mode = %v, want modeNormal after confirming", um.mode)
	}
	if len(um.repos) != 1 {
		t.Fatalf("repos len = %d, want 1 after confirm", len(um.repos))
	}
	if um.repos[0] != "/second" {
		t.Errorf("remaining repo = %q, want /second", um.repos[0])
	}
	if um.confirmTarget != "" {
		t.Errorf("confirmTarget = %q, want empty after confirm", um.confirmTarget)
	}
	if !strings.Contains(um.statusMsg, "removed:") {
		t.Errorf("statusMsg = %q, want 'removed: …'", um.statusMsg)
	}
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config not persisted: %v", err)
	}
	if len(loaded.Repos) != 1 {
		t.Errorf("persisted repos len = %d, want 1", len(loaded.Repos))
	}
}

func TestConfirmingEnterRemovesRepo(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	cfg := &config.Config{Repos: []string{"/first", "/second"}}
	c := cache.New()
	m := NewModel(cfg, cfgPath, c, filepath.Join(dir, "cache.json"))
	m.mode = modeConfirming
	m.confirmTarget = "/first"

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	um := updated.(model)

	if um.mode != modeNormal {
		t.Errorf("mode = %v, want modeNormal after enter confirm", um.mode)
	}
	if len(um.repos) != 1 {
		t.Fatalf("repos len = %d, want 1 after enter confirm", len(um.repos))
	}
}

func TestConfirmingCancelReturnsToNormal(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/first", "/second"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeConfirming
	m.confirmTarget = "/first"

	for _, key := range []string{"n", "esc", "q"} {
		t.Run(key, func(t *testing.T) {
			m2 := m
			var km tea.KeyPressMsg
			if key == "esc" {
				km = tea.KeyPressMsg{Code: tea.KeyEscape}
			} else {
				km = tea.KeyPressMsg{Text: key}
			}
			updated, _ := m2.Update(km)
			um := updated.(model)
			if um.mode != modeNormal {
				t.Errorf("mode = %v after pressing %q, want modeNormal", um.mode, key)
			}
			if len(um.repos) != 2 {
				t.Errorf("repos len = %d after pressing %q, want 2 (unchanged)", len(um.repos), key)
			}
			if um.confirmTarget != "" {
				t.Errorf("confirmTarget = %q after pressing %q, want empty", um.confirmTarget, key)
			}
		})
	}
}

func TestConfirmingIgnoresOtherKeys(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/first", "/second"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeConfirming
	m.confirmTarget = "/first"

	updated, _ := m.Update(tea.KeyPressMsg{Text: "a"})
	um := updated.(model)
	if um.mode != modeConfirming {
		t.Error("pressing invalid key should stay in confirming mode")
	}
	if len(um.repos) != 2 {
		t.Errorf("repos len = %d, want 2 (unchanged on invalid key)", len(um.repos))
	}
	if um.confirmTarget != "/first" {
		t.Errorf("confirmTarget = %q, want /first (unchanged on invalid key)", um.confirmTarget)
	}
}

func TestConfirmingViewShowsPrompt(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/my-repo"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeConfirming
	m.confirmTarget = "/home/user/my-repo"

	view := m.View()
	if !strings.Contains(view.Content, "Remove 'my-repo'?") {
		t.Errorf("confirming view should show 'Remove 'my-repo'?', got: %s", view.Content)
	}
	if !strings.Contains(view.Content, "y/Enter") {
		t.Error("confirming view should show y/Enter help")
	}
}

func TestRemoveRepoNoopWhenNoRows(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")

	updated, _ := m.Update(tea.KeyPressMsg{Text: "x"})
	um := updated.(model)
	if um.mode != modeNormal {
		t.Errorf("mode = %v, want modeNormal when no rows to remove", um.mode)
	}
	if len(um.repos) != 0 {
		t.Errorf("repos should stay empty, got %v", um.repos)
	}
}

func TestHandleDiscoveringToggleAndToggleAll(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{
		{path: "/one", selected: true},
		{path: "/two", selected: true},
	}
	m.discoverAllSelected = true

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	um := updated.(model)
	if um.discovered[0].selected {
		t.Error("space should toggle the cursor entry off")
	}
	if !um.discovered[1].selected {
		t.Error("space must not affect non-cursor entries")
	}

	updated2, _ := um.Update(tea.KeyPressMsg{Text: "a"})
	um2 := updated2.(model)
	for i, d := range um2.discovered {
		if d.selected {
			t.Errorf("discovered[%d].selected = true; 'a' should toggle all off", i)
		}
	}
}

func TestHandleDiscoveringEnterCommits(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, cfgPath, c, filepath.Join(dir, "cache.json"))
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{
		{path: "/picked", selected: true},
		{path: "/skipped", selected: false},
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	um := updated.(model)

	if um.mode != modeNormal {
		t.Errorf("mode = %v, want modeNormal after enter commit", um.mode)
	}
	if len(um.repos) != 1 || um.repos[0] != "/picked" {
		t.Errorf("repos = %v, want [/picked]", um.repos)
	}
	if um.discovered != nil {
		t.Errorf("discovered should be cleared, got %v", um.discovered)
	}
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config not persisted: %v", err)
	}
	if len(loaded.Repos) != 1 || loaded.Repos[0] != "/picked" {
		t.Errorf("persisted repos = %v, want [/picked]", loaded.Repos)
	}
}

func TestHandleDiscoveringEscCancels(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{{path: "/x", selected: true}}

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	um := updated.(model)
	if um.mode != modeNormal {
		t.Errorf("mode = %v, want modeNormal after esc", um.mode)
	}
	if len(um.repos) != 0 {
		t.Errorf("repos should stay empty when discovery cancelled; got %v", um.repos)
	}
}

func TestTruncatePlainASCIIWithinLimit(t *testing.T) {
	got := truncatePlain("hello", 10)
	if got != "hello" {
		t.Errorf("truncatePlain(%q, 10) = %q, want %q", "hello", got, "hello")
	}
}

func TestTruncatePlainCJKTruncatesByDisplayWidth(t *testing.T) {
	input := "日本語テスト名前"
	maxLen := 8
	got := truncatePlain(input, maxLen)
	if runewidth.StringWidth(got) > maxLen {
		t.Errorf("truncatePlain display width = %d, want <= %d (got %q)", runewidth.StringWidth(got), maxLen, got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("truncatePlain should end with ..., got %q", got)
	}
}

func TestTruncatePlainEmojiWithinLimit(t *testing.T) {
	input := "🎉🚀"
	got := truncatePlain(input, 10)
	if got != input {
		t.Errorf("truncatePlain(%q, 10) = %q, want %q (display width %d)", input, got, input, runewidth.StringWidth(input))
	}
}

func TestTruncatePlainEmojiOverflows(t *testing.T) {
	input := "🎉🎉🎉🎉🎉🎉"
	maxLen := 6
	got := truncatePlain(input, maxLen)
	if runewidth.StringWidth(got) > maxLen {
		t.Errorf("truncatePlain display width = %d, want <= %d (got %q)", runewidth.StringWidth(got), maxLen, got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("truncatePlain should end with ..., got %q", got)
	}
}

func TestViewDiscoveringCJKRepoNameAlignment(t *testing.T) {
	dir := t.TempDir()
	cjkDir := filepath.Join(dir, "日本語プロジェクト")
	if err := os.MkdirAll(cjkDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, filepath.Join(dir, "cfg.toml"), c, filepath.Join(dir, "cache.json"))
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{{path: cjkDir, selected: true}}
	m.discoverCursor = 0
	m.width = 80
	m.height = 24

	view := m.View()
	content := view.Content
	lines := strings.Split(content, "\n")

	var found bool
	for _, line := range lines {
		if strings.Contains(line, "日本語プロジェクト") {
			found = true
			visible := lipgloss.Width(line)
			if visible > m.width {
				t.Errorf("line display width %d exceeds terminal width %d: %q", visible, m.width, line)
			}
		}
	}
	if !found {
		t.Fatal("expected CJK repo name to appear in view output")
	}
}

func TestViewAltScreenEnabled(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if !view.AltScreen {
		t.Error("AltScreen should be true on normal mode view")
	}
}

func TestViewInitReturnsNilInV2(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil; bubbletea v2 uses View.AltScreen property, not Init commands")
	}
}

func TestAddPathCursorContainsStaticBlock(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeAdding
	view := m.View()
	if !strings.Contains(view.Content, "█") {
		t.Error("add-path mode should render a static block cursor")
	}
	if !view.AltScreen {
		t.Error("AltScreen should be true even in add-path mode")
	}
}

func TestAddPathCursorNoBlinkANSI(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeAdding
	view := m.View()
	if strings.Contains(view.Content, "\x1b[5m") {
		t.Error("add-path cursor should not contain ANSI blink escape (\\x1b[5m)")
	}
}

func TestHeaderNoEmDash(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if strings.Contains(view.Content, "—") {
		t.Error("header should use ASCII-safe separator, not em-dash (U+2014)")
	}
	if !strings.Contains(view.Content, "gitfetch - repo decay tracker") {
		t.Error("header should contain 'gitfetch - repo decay tracker'")
	}
}

func TestNormalModeHelpShowsScrollKeys(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	for _, key := range []string{"ctrl+u", "ctrl+d", "pgup", "pgdown", "home", "end", "r refresh", "a add", "d/x remove", "q quit"} {
		if !strings.Contains(view.Content, key) {
			t.Errorf("normal mode help should contain %q", key)
		}
	}
}

func TestDiscoverModeHelpShowsDiscoverBindings(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{{path: "/x", selected: true}}
	view := m.View()
	for _, key := range []string{"Space toggle", "a toggle all", "Enter confirm", "Esc cancel"} {
		if !strings.Contains(view.Content, key) {
			t.Errorf("discover mode help should contain %q", key)
		}
	}
	if strings.Contains(view.Content, "r refresh") {
		t.Error("discover mode help should not contain normal-mode 'r refresh' binding")
	}
	if strings.Contains(view.Content, "d/x remove") {
		t.Error("discover mode help should not contain normal-mode 'd/x remove' binding")
	}
}

func TestDiscoverModeIgnoresRefreshKey(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/repo"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{{path: "/x", selected: true}}

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "r"})
	um := updated.(model)
	if um.scanning {
		t.Error("discover mode should ignore 'r' — scanning should not start")
	}
	if cmd != nil {
		t.Error("discover mode should ignore 'r' — cmd should be nil")
	}
}

func TestDiscoverModeIgnoresRemoveKey(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/repo"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{{path: "/x", selected: true}}
	before := len(m.repos)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "d"})
	um := updated.(model)
	if len(um.repos) != before {
		t.Errorf("discover mode should ignore 'd' — repos changed from %d to %d", before, len(um.repos))
	}
}

func TestViewDiscoveringLongCJKPathTruncation(t *testing.T) {
	dir := t.TempDir()
	deepPath := dir
	for i := 0; i < 20; i++ {
		deepPath = filepath.Join(deepPath, "日本語ディレクトリ")
	}

	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.mode = modeDiscovering
	m.discovered = []discoveredRepo{{path: deepPath, selected: true}}
	m.discoverCursor = 0
	m.width = 40
	m.height = 24

	view := m.View()
	content := view.Content
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "日本語ディレクトリ") {
			visible := lipgloss.Width(line)
			if visible > m.width {
				t.Errorf("repo line display width %d exceeds terminal width %d: %q", visible, m.width, line)
			}
		}
	}
}

func TestScanProgressCounterShown(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/a", "/b", "/c"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.scanning = true

	now := time.Now().UTC()
	msg := scanProgressMsg{
		done:  2,
		total: 3,
		result: gitscanner.ScanResult{
			RepoPath:       "/b",
			LastCommitDate: &now,
			ScannedAt:      now,
		},
		remaining: []string{"/c"},
	}
	updated, _ := m.Update(msg)
	um := updated.(model)
	if !um.scanning {
		t.Error("scanning should still be true with remaining repos")
	}
	if um.statusMsg != "scanning 2/3…" {
		t.Errorf("status = %q, want %q", um.statusMsg, "scanning 2/3…")
	}
	if um.scanDone != 2 || um.scanTotal != 3 {
		t.Errorf("progress = %d/%d, want 2/3", um.scanDone, um.scanTotal)
	}
	if len(um.scanResults) != 1 {
		t.Errorf("scanResults len = %d, want 1", len(um.scanResults))
	}
}

func TestEmptyCacheShowsRefreshPrompt(t *testing.T) {
	cfg := &config.Config{Repos: []string{"/home/user/project"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if !strings.Contains(view.Content, "No scan data yet") {
		t.Error("view should show 'No scan data yet' when repos exist but cache is empty")
	}
	if !strings.Contains(view.Content, "Press 'r' to refresh") {
		t.Error("view should prompt user to press 'r' to refresh")
	}
	if strings.Contains(view.Content, "No repos tracked") {
		t.Error("view should not show 'No repos tracked' when repos exist in config")
	}
}

func TestEmptyConfigShowsNoReposTracked(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	view := m.View()
	if !strings.Contains(view.Content, "No repos tracked") {
		t.Error("view should show 'No repos tracked' when no repos configured")
	}
	if strings.Contains(view.Content, "No scan data yet") {
		t.Error("view should not show 'No scan data yet' when no repos are tracked")
	}
}

func TestErrorClearsOnSuccessfulScan(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")
	cfg := &config.Config{Repos: []string{"/home/user/x"}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, cachePath)
	m.err = fmt.Errorf("previous error")
	m.errSeq = 1

	updated, _ := m.Update(tea.KeyPressMsg{Text: "r"})
	um := updated.(model)
	if um.err != nil {
		t.Errorf("error should clear on scan start, got %v", um.err)
	}
}

func TestErrorPersistsWithoutAction(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.err = fmt.Errorf("persistent error")
	view := m.View()
	if !strings.Contains(view.Content, "error: persistent error") {
		t.Error("error should persist in view until cleared")
	}
}

func TestErrorAutoClearOnMatchingSeq(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.err = fmt.Errorf("auto-clear error")
	m.errSeq = 5

	updated, _ := m.Update(clearErrorMsg{seq: 5})
	um := updated.(model)
	if um.err != nil {
		t.Errorf("error should auto-clear when seq matches, got %v", um.err)
	}
}

func TestErrorNoAutoClearOnSeqMismatch(t *testing.T) {
	cfg := &config.Config{Repos: []string{}}
	c := cache.New()
	m := NewModel(cfg, "/cfg", c, "/cache")
	m.err = fmt.Errorf("newer error")
	m.errSeq = 10

	updated, _ := m.Update(clearErrorMsg{seq: 5})
	um := updated.(model)
	if um.err == nil {
		t.Error("error should NOT clear when seq does not match (stale tick)")
	}
}

func TestErrorClearsOnSuccessfulRemove(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	cfg := &config.Config{Repos: []string{"/first", "/second"}}
	c := cache.New()
	now := time.Now().UTC()
	c.Repos["/first"] = cache.RepoEntry{LastCommitDate: &now, ScannedAt: now}
	c.Repos["/second"] = cache.RepoEntry{LastCommitDate: &now, ScannedAt: now}
	m := NewModel(cfg, cfgPath, c, filepath.Join(dir, "cache.json"))
	m.err = fmt.Errorf("previous error")
	m.errSeq = 3
	m.mode = modeConfirming
	m.confirmTarget = "/first"

	updated, _ := m.Update(tea.KeyPressMsg{Text: "y"})
	um := updated.(model)
	if um.err != nil {
		t.Errorf("error should clear on successful remove, got %v", um.err)
	}
}

func TestTableResponsiveLayout(t *testing.T) {
	now := time.Now()
	repos := map[string]cache.RepoEntry{"/長い名前-project": {LastCommitDate: &now, LastTagDate: &now, LastTag: "v1.2.3", ScannedAt: now, WeeklyCommits: []int{0, 1, 2, 3, 4, 5, 6, 8}}}
	for _, width := range []int{20, 40, 64, 80, 120} {
		tm, _ := NewTable(repos, 10, true, width)
		out := tm.View()
		for _, line := range strings.Split(out, "\n") {
			if lipgloss.Width(line) > width {
				t.Errorf("width %d overflow: %q", width, line)
			}
		}
		if width >= 64 {
			if !strings.Contains(out, "Commit") || strings.Contains(out, "Commit 8w") || !strings.Contains(tm.SelectedRow()[2], "·▁▂▃▄▅▆█") {
				t.Errorf("commit column missing activity: %s", out)
			}
			if len(tm.SelectedRow()) != 5 {
				t.Error("expected hidden path plus repo, commit, release, version")
			}
		}
		if width < 64 && strings.Contains(out, "░") {
			t.Errorf("narrow table should omit bars: %s", out)
		}
		if tm.SelectedRow()[0] != "/長い名前-project" {
			t.Fatal("hidden path was truncated")
		}
	}
}

func TestSelectedDetailsPreserveValues(t *testing.T) {
	now := time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)
	path := "/long/path/project"
	c := cache.New()
	c.Repos[path] = cache.RepoEntry{LastCommitDate: &now, LastTagDate: &now, LastTag: "v1.2.3", ScannedAt: time.Now().AddDate(0, 0, -14), WeeklyCommits: []int{0, 1, 2, 3, 4, 5, 6, 7}, Error: "a full scan error with useful context"}
	m := NewModel(&config.Config{Repos: []string{path}}, "", c, "")
	m.width = 120
	m.rebuildTable()
	out := m.selectedDetails()
	for _, want := range []string{path, "2026-09-05T10:00:00Z", "v1.2.3", cache.WeekStart(time.Now()).AddDate(0, 0, -56).Format("2006-01-02") + " to " + cache.WeekStart(time.Now()).Format("2006-01-02"), "a full scan error with useful context"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q: %s", want, out)
		}
	}
}

func TestPageNavigationUsesVisibleHeight(t *testing.T) {
	c := cache.New()
	repos := []string{}
	now := time.Now()
	for i := 0; i < 50; i++ {
		path := fmt.Sprintf("/long/path/that/needs/to/wrap/in/a/narrow/terminal/repo-%02d", i)
		repos = append(repos, path)
		c.Repos[path] = cache.RepoEntry{LastCommitDate: &now, ScannedAt: now, WeeklyCommits: []int{0, 1, 2, 3, 4, 5, 6, 7}}
	}
	m := NewModel(&config.Config{Repos: repos}, "", c, "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	m = updated.(model)
	// Table height includes its two-line header; page navigation uses body rows.
	visibleRows := lipgloss.Height(m.table.View()) - 2
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	m = updated.(model)
	if m.table.Cursor() != visibleRows {
		t.Errorf("page jumped %d rows; visible body has %d", m.table.Cursor(), visibleRows)
	}
}
