## Cycle 1 · 2026-04-15

**What**: project scaffold — Go module, config package (TOML/XDG), cobra CLI routing, 12 tests
**Commit**: c85213c feat: scaffold project with config package and CLI routing
**Inspiration**: pelletier/go-toml/v2 for TOML, spf13/cobra for CLI
**Discovered**: none
**Verified**: `go build ./...` OK, `go vet ./...` OK, `go test ./...` 12/12 PASS
**Next**: Task 2 — git scanner and cache engine
**Context**: bootstrap only · no scope creep into scanning/cache/display/tui · TOML config, XDG paths, cobra subcommands
