# TODO

Sourced from `.agentera/HEALTH.md`. Audit 3 remediation plan complete 2026-04-17 (see `.agentera/archive/PLAN-2026-04-17-audit3-remediation.md`).

## ⇶ Critical

_(empty)_

## ⇉ Degraded

_(empty)_

## ⇢ Annoying

- [ ] Consider decomposing `handleDiscovering` (57 lines / 6 branches / 3 config.Save sites) if the discover mode grows new keys — Audit 3 info finding, HEALTH.md notes deferral is appropriate until the mode expands further

### Resonera decisions (2026-04-22)
- [x] `internal/display` → bubbles coupling — resolved in Decision 2 (split to `core/` + `theme/` + `display/` + `tui/`) · see `.agentera/DECISIONS.md`
- [x] Layout and visual hierarchy — resolved in Decision 3 (compact side-by-side bars in single "Decay" column, flat list sorted by commit decay, repo name first) · see `.agentera/DECISIONS.md`

## Resolved

### Audit 2 plan (2026-04-17)
- [x] TUI 'r' key refresh — fixed in 3d5fcec (Task 3)
- [x] `ScanRepo` perf regression — fixed in 3e28bb5 (Task 2)
- [x] TUI Update-driven tests — added across 3d5fcec, bd8e631, 44993d1 (Tasks 3, 5, 6); coverage now 58.4%
- [x] Repo-filter loop DRY — fixed in 7d9159e via `cache.FilterByRepos` (Task 1)
- [x] Scanner gap tests (`DiscoverRepos`, `matchFilter`, `listRemotes`) — added in 3e28bb5 (Task 2)
- [x] PROGRESS.md / CHANGELOG.md sync — done across the plan's 6 docs commits
- [x] `decay.Tier.Label()` alias — removed in 010c4d3 (Task 4)
- [x] `commitNewRepo` extraction from `handleAdding` — done in bd8e631 (Task 5)

### Audit 3 plan (2026-04-17)
- [x] Unused `decay.DaysUntilNext` export — removed in 67c39d9 (Task 1); decay coverage 77.4% → 92.3%
- [x] TUI help text `↓/k` → `↓/j` — fixed in e3b7414 (Task 2)
- [x] `tui.filteredCache` wrapper — inlined into `cache.FilterByRepos` in af13a62 (Task 3)
- [x] Unexported `discoverRepos` shim — inlined into `ResolveRepoPaths` in 5f49c30 (Task 4)
- [x] Unused `stripANSI` test helper — deleted in 444abdc (Task 5)
- [x] Pre-existing tui.go linter hints (`slices.Contains`, built-in `max`, `fmt.Fprintf`) — modernized in c717b5a (Task 6)
