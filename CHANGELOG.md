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

### Changed
- Default config no longer seeds `~/projects`; new installs start with an empty repo list (opt in via `gitfetch tui` or `config.toml`)
- `cache.FilterByRepos` centralizes the "filter cache by configured repos" logic; `cmd/gitfetch` and `internal/tui` now share a single source of truth
- Scanner skips author and remote collection when no `--author`/`--remote` filter is set, restoring sub-500ms refresh on repos with long history
- `DiscoverRepos` now returns `([]string, error)` so unreadable directories surface a wrapped error instead of a silent single-path fallback
- Dropped the redundant `decay.Tier.Label()` alias; callers use `Tier.String()` directly
- TUI add-repo flow split: `handleAdding` is a pure keystroke router, `commitNewRepo` owns validation/persistence/discovery dispatch
- Duplicate path entered in add-repo now keeps the user in add mode with a visible status so they can correct and retry

### Fixed
- TUI `r` key now actually triggers a refresh: scans all tracked repos, shows a `scanning…` status, writes the cache on disk, and updates displayed rows
