# TODO

Sourced from `.agentera/HEALTH.md` Audit 2 · 2026-04-17. Audit 2 remediation plan complete 2026-04-17 (see `.agentera/archive/PLAN-2026-04-17-audit2-remediation.md`).

## ⇶ Critical

_(empty — Audit 2 critical finding resolved)_

## ⇉ Degraded

_(empty — all Audit 2 degraded findings resolved)_

## ⇢ Annoying

- [ ] Decide whether `internal/display` should stay bubbles-aware or split into pure-format + table-widget submodules (Coupling finding, Audit 2 — deferred to `/resonera`)
- [ ] Pre-existing linter hints in `internal/tui/tui.go`: simplify loop with `slices.Contains`, modernize `if` with `max`, replace `WriteString(fmt.Sprintf(...))` with `fmt.Fprintf` (surfaced during Audit 2 plan, out of scope)

## Resolved

- [x] TUI 'r' key refresh — fixed in 3d5fcec (Task 3)
- [x] `ScanRepo` perf regression — fixed in 3e28bb5 (Task 2)
- [x] TUI Update-driven tests — added across 3d5fcec, bd8e631, 44993d1 (Tasks 3, 5, 6); coverage now 58.4%
- [x] Repo-filter loop DRY — fixed in 7d9159e via `cache.FilterByRepos` (Task 1)
- [x] Scanner gap tests (`DiscoverRepos`, `matchFilter`, `listRemotes`) — added in 3e28bb5 (Task 2)
- [x] PROGRESS.md / CHANGELOG.md sync — done across the plan's 6 docs commits
- [x] `decay.Tier.Label()` alias — removed in 010c4d3 (Task 4)
- [x] `commitNewRepo` extraction from `handleAdding` — done in bd8e631 (Task 5)
