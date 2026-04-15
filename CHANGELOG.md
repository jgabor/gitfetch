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
