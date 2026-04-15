# Plan: Bootstrap gitfetch — repo decay tracker

<!-- Level: full · Created: 2026-04-15 · Status: active -->
<!-- Reviewed: 2026-04-15 | Critic issues: 5 found, 3 addressed, 2 dismissed -->

## What

Build gitfetch, a cache-first Go TUI that scans git repos for commit and release staleness, classifies them into decay tiers (fresh/stale/decayed/dead), and displays a color-coded dashboard. Three modes: instant print from cache (default), scan-and-update (refresh), and interactive repo management (tui).

## Why

Personal tool for tracking repo neglect across many checkouts. The "neofetch for repos" framing: runs on terminal open and instantly shows which projects need attention. Decision 1 defines the full product; this plan decomposes it into buildable tasks.

## Constraints

- Must print from cache and exit in under 500ms (for .bashrc use)
- Human-only output: no JSON, no API, no machine consumers
- TOML config, JSON cache, XDG directory conventions
- Charm stack (bubbletea, lipgloss) for TUI and formatting
- Shell out to git for scanning (no go-git dependency)
- Standard Go project layout: cmd/gitfetch/, internal/
- VISION.md does not exist yet; plan proceeds from Decision 1 which is firm

## Scope

**In**: Config loading, git scanning, cache I/O, decay calculation, terminal display, print/refresh/tui commands
**Out**: Remote repo scanning (local checkouts only), CI/CD integration, multi-user support
**Deferred**: Custom decay tier thresholds in config, configurable tag patterns, cache reconciliation on config changes

## Design

Six packages under `internal/`:
- **config**: TOML parsing, XDG path resolution, default generation
- **git**: Shell-based repo scanner extracting last commit date and last v*-tag date
- **cache**: JSON read/write to XDG_DATA_HOME
- **decay**: Tier classification with hardcoded thresholds (fresh <30d, stale <90d, decayed <180d, dead ≥180d)
- **display**: Lipgloss-based formatting with two-column progress bars and tier colors
- **tui**: Bubbletea model for interactive repo management

Data flow: config → scanner → cache → decay tiers → display. Default command reads cache end-to-end. Refresh runs scanner first. TUI wraps everything interactively.

## Tasks

### Task 1: Project scaffold and config
**Depends on**: none
**Status**: ■ complete
**Acceptance**:
▸ GIVEN a fresh install with no config WHEN gitfetch runs THEN it creates a default config at XDG_CONFIG_HOME/gitfetch/config.toml with sane defaults
▸ GIVEN a valid config file with repo paths WHEN gitfetch loads THEN it parses all repo paths and display preferences
▸ GIVEN an invalid TOML config WHEN gitfetch loads THEN it prints a clear error message and exits non-zero
▸ Test proportionality: 1 pass + 1 fail per testable unit (config parsing, path resolution, default generation)

### Task 2: Git scanner and cache engine
**Depends on**: Task 1
**Status**: □ pending
**Acceptance**:
▸ GIVEN a configured repo with recent commits WHEN the scanner runs THEN the cache contains the last commit date for that repo
▸ GIVEN a configured repo with v-prefixed tags WHEN the scanner runs THEN the cache contains the most recent tag date
▸ GIVEN a configured repo with no tags WHEN the scanner runs THEN the cache records no release date (commit age only)
▸ GIVEN a populated cache file WHEN gitfetch reads it THEN all repo data deserializes correctly
▸ GIVEN a configured path that is not a git repo WHEN the scanner encounters it THEN the cache records an error for that repo without crashing
▸ GIVEN a configured path that does not exist WHEN the scanner encounters it THEN the cache records an error for that repo without crashing
▸ Test proportionality: 1 pass + 1 fail per testable unit; edge case expansion for scanner (non-git dir, missing path, empty repo) — multi-branch error handling

### Task 3: Decay calculation and display formatting
**Depends on**: Task 1
**Status**: □ pending
**Acceptance**:
▸ GIVEN a repo with a commit 10 days ago WHEN decay is calculated THEN it is classified as "fresh"
▸ GIVEN a repo with a commit 60 days ago WHEN decay is calculated THEN it is classified as "stale"
▸ GIVEN a repo with a commit 120 days ago WHEN decay is calculated THEN it is classified as "decayed"
▸ GIVEN a repo with a commit 200 days ago WHEN decay is calculated THEN it is classified as "dead"
▸ GIVEN multiple repos at varying decay levels WHEN the dashboard renders THEN each repo shows two tier-labeled color-coded progress bars (commit age, release age)
▸ Test proportionality: edge case expansion for tier boundary logic (29/30/31d, 89/90/91d, 179/180/181d) — 3+ conditional branches warrant boundary tests

### Task 4: Print and refresh commands
**Depends on**: Task 2, Task 3
**Status**: □ pending
**Acceptance**:
▸ GIVEN a populated cache WHEN gitfetch runs with no arguments THEN it prints the formatted decay dashboard and exits
▸ GIVEN an empty cache WHEN gitfetch runs THEN it prints a message suggesting gitfetch refresh and exits
▸ GIVEN configured repos WHEN gitfetch refresh runs THEN all repos are scanned, cache is updated, and the dashboard prints
▸ GIVEN gitfetch is invoked from a shell rc file with a populated cache WHEN the terminal opens THEN the dashboard renders in under 500ms

### Task 5: TUI for repo management
**Depends on**: Task 4
**Status**: □ pending
**Acceptance**:
▸ GIVEN gitfetch tui is launched WHEN the interface loads THEN it shows a checklist of configured repos with current decay status
▸ GIVEN the TUI is showing repos WHEN the user triggers an inline refresh THEN scan data updates and the display refreshes in place
▸ GIVEN the TUI is showing repos WHEN the user adds a new repo path THEN the config updates and the repo appears in the display
▸ GIVEN the TUI is showing repos WHEN the user removes a repo THEN the config updates and the repo disappears from the display

### Task 6: Plan-level freshness checkpoint
**Depends on**: Task 1, Task 2, Task 3, Task 4, Task 5
**Status**: □ pending
**Acceptance**:
▸ GIVEN this plan's work has shipped WHEN CHANGELOG.md is checked THEN it has an [Unreleased] section with Added entries summarizing each task's user-visible impact
▸ GIVEN this plan is complete WHEN PROGRESS.md is checked THEN it has at least one cycle entry whose What field summarizes the plan
▸ GIVEN this plan is complete WHEN the project root is inspected THEN VISION.md exists reflecting the product direction from Decision 1

## Overall Acceptance
▸ GIVEN gitfetch is installed and configured WHEN a terminal opens via .bashrc THEN the user sees an instant color-coded decay dashboard for all tracked repos
▸ GIVEN the user runs gitfetch refresh WHEN scanning completes THEN the dashboard reflects current commit and release ages
▸ GIVEN the user runs gitfetch tui WHEN the interface opens THEN they can add repos, remove repos, and trigger inline refreshes

## Surprises
[Populated by realisera during execution when reality diverges from plan]
