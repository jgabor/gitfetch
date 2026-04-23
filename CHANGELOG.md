# Changelog

## [Unreleased]

### Added
- Go module and standard project layout (cmd/gitfetch/, internal/config/)
- TOML config package with XDG path resolution and default config generation
- Cobra CLI with three modes: print (default), refresh, tui
- 12 tests covering config parsing, path resolution, default generation, and error handling
- Git scanner (internal/git) shelling out to git log/tag for commit and v-tag dates
- JSON cache engine (internal/cache) with read/write to XDG_DATA_HOME/gitfetch/cache.json
- `gitfetch refresh` command scans configured repos and writes results to cache
- 17 new tests covering scanner (commits, v-tags, errors, edge cases) and cache (roundtrip, empty, errors)
- Decay tier classification (internal/decay) with four tiers: fresh (<30d), stale (<90d), decayed (<180d), dead (≥180d)
- Lipgloss-based dashboard formatting (internal/display) with color-coded progress bars per tier
- Boundary-tested tier classification at 29/30/31, 89/90/91, 179/180/181 day thresholds
- 40 new tests covering decay classification (24) and display formatting (16)
- `gitfetch` default command reads cache and prints formatted decay dashboard (cache-only, no git calls)
- `gitfetch refresh` now displays the formatted dashboard after scanning and writing cache
- Empty cache prints suggestion message directing user to `gitfetch refresh`
- Bubbletea TUI (`gitfetch tui`) for interactive repo management with navigation, inline refresh, add/remove repos
- Key bindings: j/k navigate, r refresh, a add repo, d/x remove repo, q quit
- `gitfetch refresh --author <substring>` and `--remote <substring>` flags filter which tracked repos are rescanned
- Scanner collects per-repo authors and remotes (used only for filtering; not persisted to cache)
- TUI multi-repo discovery: adding a directory path opens a checklist of detected sub-repos with toggle-all, per-row space/enter selection, and Esc to cancel
- Display layer rewritten on top of `bubbles/table` and shared between the non-interactive dashboard and the TUI
- Tag name cached alongside tag date in `cache.RepoEntry.LastTag`; displayed as the "Version" column
- Config loader expands a leading `~/` in repo paths so configs are portable across hosts
- `refresh` prunes cache entries whose repo was removed from config
- `internal/core/` package with `RepoRow`, `BuildRows`, and `RepoColumnWidth` for shared data between print and TUI
- `internal/theme/` package with `GradientBar` and `TierColor` helpers using `lipgloss` + `go-colorful` for per-character Lab-space interpolation
- Dual-decay layout in print dashboard — commit and tag decay bars rendered side by side in a single "Decay" column

### Changed
- `decay.FreshLimit` tightened from 30 → 10 days to reflect "fresh" meaning actively-worked-on; decay tests and `core.BuildRows` fixtures updated accordingly; spectrum palette segments recompute automatically from the new tier bounds
- TUI display-width correctness: `viewDiscovering` and `truncatePlain` use `runewidth.StringWidth`/`runewidth.Truncate` instead of byte-based `len()` — fixes column alignment and truncation for multibyte repo names
- TUI terminal robustness: cursor renders as static block (no ANSI blink), header uses ASCII-safe dash separator, `visibleRange` edge case already handled correctly
- TUI keybinding cleanup: normal mode help bar lists all scroll keys (ctrl+u, ctrl+d, pgup, pgdown, home, end); discover mode shows only discover-specific bindings
- TUI scan feedback: restructured scan to emit per-repo progress (`scanProgressMsg`) showing counter ("scanning 3/12…") instead of static "scanning…"
- TUI empty states: shows "No scan data yet. Press 'r' to refresh." when repos exist but cache is empty
- TUI error lifecycle: errors auto-clear after 5 seconds (with sequence counter to prevent stale ticks) and clear on any successful action
- `theme.GradientBar` accepts dynamic width parameter; `ComputeBarWidth` adapts bar to terminal width
- TUI confirmation mode: pressing `d`/`x` now shows "Remove 'name'? y/Enter · n/Esc/q" prompt before deleting — no more accidental removals
- Default config no longer seeds `~/projects`; new installs start with an empty repo list (opt in via `gitfetch tui` or `config.toml`)
- `cache.FilterByRepos` centralizes the "filter cache by configured repos" logic; `cmd/gitfetch` and `internal/tui` now share a single source of truth
- Scanner skips author and remote collection when no `--author`/`--remote` filter is set, restoring sub-500ms refresh on repos with long history
- `DiscoverRepos` now returns `([]string, error)` so unreadable directories surface a wrapped error instead of a silent single-path fallback
- Dropped the redundant `decay.Tier.Label()` alias; callers use `Tier.String()` directly
- TUI add-repo flow split: `handleAdding` is a pure keystroke router, `commitNewRepo` owns validation/persistence/discovery dispatch
- Duplicate path entered in add-repo now keeps the user in add mode with a visible status so they can correct and retry
- TUI test coverage rebuilt from 14.8% to 58.4% (visibleRange, filteredCache, removeRepo, handleDiscovering all now exercised via Update)
- Removed unused exported `decay.DaysUntilNext`; decay package coverage restored to 92.3%
- Inlined `tui.filteredCache` wrapper at both call sites (`rebuildTable` and tests use `cache.FilterByRepos` directly)
- Inlined the unexported `discoverRepos` shim in `internal/git/scanner.go` into `ResolveRepoPaths`; removes the `DiscoverRepos`/`discoverRepos` case-only name collision
- Dropped unused `stripANSI` test helper in `internal/display/display_test.go`
- Modernized three tui.go idioms: `slices.Contains` for the duplicate-path check, built-in `max()` for two clamp patterns, `fmt.Fprintf` for a `WriteString(fmt.Sprintf(...))` site — clears the three TODO Annoying linter hints parked during the Audit 2 plan
- `internal/display/` refactored to print-only — no longer imports `bubbles/table`; `FormatDashboard` uses `theme.GradientBar` for both bars
- TUI widget functions (`NewTable`, `TableStyles`, `Columns`, `rowToTableRow`) moved from `internal/display/` to `internal/tui/`

### Fixed
- TUI `r` key now actually triggers a refresh: scans all tracked repos, shows a `scanning…` status, writes the cache on disk, and updates displayed rows
- TUI help text now correctly pairs `↓` with `j` instead of `k` (vim-style navigation)
