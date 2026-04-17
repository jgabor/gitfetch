# TODO

Sourced from `.agentera/HEALTH.md` Audit 2 · 2026-04-17.

## ⇶ Critical

- [ ] TUI 'r' key no longer refreshes — wire `startScan` into `handleNormalKeys` (`internal/tui/tui.go:150`) and add an `Update` test that asserts a non-nil `tea.Cmd` is returned for `r`

## ⇉ Degraded

- [ ] Guard `listAuthors` / `listRemotes` behind `opts.Author != ""` / `opts.Remote != ""` in `ScanRepo` (`internal/git/scanner.go:35-49`) — unconditional author history scan regresses cache-first sub-500ms principle
- [ ] Add `Update`-driven tests for TUI: `handleAdding`, `handleDiscovering`, `handleScanDone`, `removeRepo`, `visibleRange`, `filteredCache` (coverage fell from 27.6% to 14.8% while `tui.go` tripled in size)
- [ ] Deduplicate repo-filter loop between `cmd/gitfetch/main.go:35-44` and `internal/tui/tui.go:69-81` — extract a shared `FilterByRepos` helper
- [ ] Add direct tests for `DiscoverRepos`, `matchFilter`, and `listRemotes` error path in `internal/git/scanner.go`
- [ ] Sync `.agentera/PROGRESS.md` + `CHANGELOG.md` with the uncommitted TUI/scanner changes (788+/392- lines undocumented) — or run `/dokumentera`

## ⇢ Annoying

- [ ] Remove redundant `decay.Tier.Label()` alias (`internal/decay/decay.go:37-39`); switch the one caller in `display.rowToTableRow` to `String()`
- [ ] Extract `commitNewRepo` from `handleAdding` (`internal/tui/tui.go:161-224`) to isolate persistence/discovery from keystroke routing
- [ ] Decide whether `internal/display` should stay bubbles-aware or split into pure-format + table-widget submodules (Coupling finding, Audit 2)
