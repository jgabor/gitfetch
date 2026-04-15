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
