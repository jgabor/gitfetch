# Plan: Audit 2 Remediation

<!-- Level: full | Created: 2026-04-17 | Status: active -->
<!-- Reviewed: 2026-04-17 | Critic issues: 9 found, 8 addressed, 1 dismissed -->

## What

Resolve the actionable findings from HEALTH.md Audit 2: fix the silently-broken `r` refresh key, restore the cache-first perf guarantee by skipping author/remote history when no filter is set, rebuild TUI test coverage, DRY the duplicated filter loop, and clean up minor API cruft.

## Why

Audit 2 flagged one critical (broken `r` key), four degraded (perf regression, TUI coverage collapse, DRY violation, scanner test gaps), and three annoying findings. The trajectory is ⮋ degrading vs Audit 1. Leaving this unaddressed compounds drift: the cache-first sub-500ms principle is violated on every refresh, core TUI behavior is silently broken, and the safety net (tests) shrank while surface area tripled. Restoring ground truth now keeps the project's A-/B trajectory intact.

## Constraints

- Cache-first, sub-500ms default command — any change that regresses this violates VISION's north star
- Human-only output; no JSON, no machine consumers
- Shell out to git; no go-git, no daemon, no network
- TOML config, XDG paths, Charm ecosystem (bubbletea/bubbles/lipgloss)
- Standard Go layout: `cmd/gitfetch`, `internal/*`

## Scope

**In**: broken `r` fix, scanner perf guard, scanner gap tests, TUI coverage rebuild, FilterByRepos helper, handleAdding decomposition, decay.Label cleanup.

**Out**: display coupling decision (bubbles-aware vs split) — needs deliberation, belongs to `/resonera`. README.md and CLAUDE.md authoring — separate cycle per DOCS.md.

**Deferred**: persisting `Authors`/`Remotes` to cache (alternative path for the perf finding, not chosen here).

## Design

Three layers touch. `internal/git/scanner.go` gains filter-gated author/remote collection. `internal/cache` gains an exported `FilterByRepos` helper consumed by both `cmd/gitfetch` and `internal/tui`. `internal/tui/tui.go` decomposes `handleAdding` into a keystroke router plus a `commitNewRepo` helper, wires `startScan` into the `r` branch, and grows a proper Update-driven test harness. `internal/decay` sheds the redundant `Label()` alias.

No new packages, no dependency direction changes, no framework additions. The display→bubbles coupling acknowledged in Audit 2 is not addressed here (deferred to resonera).

## Tasks

### Task 1: Extract FilterByRepos helper
**Depends on**: none
**Status**: ■ complete
**Acceptance**:
▸ GIVEN a cache map and a list of configured repo paths WHEN FilterByRepos is called THEN it returns only entries whose path matches a configured repo
▸ GIVEN an empty repo list WHEN FilterByRepos is called THEN the returned map is empty
▸ GIVEN the same inputs WHEN both `cmd/gitfetch` and `internal/tui` produce their filtered view THEN they return identical maps (verified by a shared test or by calling a single exported helper)
▸ Proportionality: 1 pass + 1 fail test per behavior

### Task 2: Scanner perf guard + gap tests
**Depends on**: none
**Status**: ■ complete
**Acceptance**:
▸ GIVEN `ScanRepo` is invoked with `opts.Author == ""` WHEN scanning completes THEN `ScanResult.Authors` is nil or empty
▸ GIVEN `ScanRepo` is invoked with `opts.Remote == ""` WHEN scanning completes THEN `ScanResult.Remotes` is nil or empty
▸ GIVEN `ScanRepo` is invoked with `opts.Author` set WHEN scanning completes THEN `ScanResult.Authors` contains the repo's commit authors
▸ GIVEN a directory containing zero, one, or multiple nested `.git` repos WHEN `DiscoverRepos` runs THEN the returned slice matches the expected set
▸ GIVEN an unreadable directory WHEN `DiscoverRepos` runs THEN it returns an error wrapped with context
▸ GIVEN a case-mixed query and haystack WHEN `matchFilter` runs THEN match succeeds (case-insensitive substring)
▸ GIVEN a repo whose remote subprocess exits non-zero WHEN `listRemotes` runs THEN it returns an error wrapped with context
▸ Proportionality: 1 pass + 1 fail per unit

### Task 3: Fix broken `r` refresh
**Depends on**: none
**Status**: ■ complete
**Acceptance**:
▸ GIVEN a TUI in `modeNormal` WHEN the user presses `r` THEN `handleNormalKeys` returns a non-nil `tea.Cmd`
▸ GIVEN that `tea.Cmd` WHEN invoked THEN the TUI transitions into a scanning state visible to the user (mode change, status line, or equivalent observable effect)
▸ GIVEN the scan completes WHEN its result message is delivered to `Update` THEN the cache on disk reflects the new data and displayed rows update
▸ GIVEN the test suite WHEN invoked THEN at least one Update-driven test sends a `tea.KeyMsg{Runes:[]rune{'r'}}` and asserts both the non-nil cmd and the state transition
▸ Proportionality: 1 pass + 1 fail

### Task 4: Remove decay.Tier.Label() alias
**Depends on**: none
**Status**: □ pending
**Acceptance**:
▸ GIVEN `internal/decay` WHEN inspected THEN `Tier.Label` is absent
▸ GIVEN `display.rowToTableRow` WHEN rendering THEN it uses `Tier.String()`
▸ GIVEN `go build ./...` and `go test ./...` WHEN run THEN both pass
▸ GIVEN any new test introduced later in this plan WHEN inspected THEN it does not call `Tier.Label`

### Task 5: Extract commitNewRepo from handleAdding
**Depends on**: Task 3, Task 4
**Status**: □ pending
**Acceptance**:
▸ GIVEN `handleAdding` WHEN inspected THEN it only routes keystrokes; persistence, path validation, and discovery dispatch live in a `commitNewRepo` helper
▸ GIVEN `commitNewRepo` WHEN called with a valid single-repo path THEN the config is saved and the TUI returns to `modeNormal` with the new repo tracked
▸ GIVEN `commitNewRepo` WHEN called with a directory containing multiple repos THEN the TUI transitions into discovery/selection mode
▸ GIVEN `commitNewRepo` WHEN called with an invalid or duplicate path THEN the TUI stays in add mode with a visible error and the input buffer behavior matches existing semantics
▸ GIVEN `go test ./...` WHEN run THEN it passes
▸ Proportionality: 1 pass + 1 fail per new unit

### Task 6: Rebuild TUI coverage
**Depends on**: Task 1, Task 3, Task 5
**Status**: □ pending
**Acceptance**:
▸ GIVEN the TUI package WHEN `go test -cover ./internal/tui/` runs THEN coverage is ≥ 40% (baseline 14.8%)
▸ GIVEN Update-driven tests WHEN run THEN they cover `handleAdding` happy + error paths, `handleDiscovering` (toggle + commit), `handleScanDone`, `removeRepo`
▸ GIVEN unit tests for pure helpers WHEN run THEN they cover `visibleRange` bounds (zero, one, viewport-sized, larger-than-viewport) and `filteredCache` behavior
▸ Proportionality: 1 pass + 1 fail per new unit; `visibleRange` may expand edge cases (rationale: viewport math with multiple boundary conditions)

### Task 7: Plan-level freshness checkpoint
**Depends on**: Task 1, Task 2, Task 3, Task 4, Task 5, Task 6
**Status**: □ pending
**Acceptance**:
▸ GIVEN the plan's completion WHEN `CHANGELOG.md` is inspected THEN Added/Changed/Fixed entries reflect every task in this plan
▸ GIVEN the plan's completion WHEN `.agentera/PROGRESS.md` is inspected THEN a cycle entry summarizes the plan's aggregate outcome at the plan level
▸ GIVEN the plan's completion WHEN `TODO.md` is inspected THEN resolved items are removed or marked complete and the Audit 2 source header is updated
▸ GIVEN the plan's completion WHEN `.agentera/DOCS.md` is inspected THEN its Index reflects the current Last Updated dates for modified artifacts
▸ GIVEN the plan WHEN archived THEN it lives at `.agentera/archive/PLAN-2026-04-17-audit2-remediation.md` and `.agentera/PLAN.md` is removed

## Overall Acceptance

▸ GIVEN all tasks complete WHEN `go test ./...` runs THEN every package passes
▸ GIVEN all tasks complete WHEN `gitfetch refresh` runs without `--author` or `--remote` filters on a repo with long history THEN `ScanResult.Authors` and `ScanResult.Remotes` are empty (behavioral proxy for "no wasted history scan")
▸ GIVEN all tasks complete WHEN the TUI opens and the user presses `r` THEN a refresh is triggered and the resulting cache on disk reflects the scan
▸ GIVEN all tasks complete WHEN `gitfetch` (default command) runs against a populated cache THEN it completes in under 500ms (VISION north-star verification; measured by hand if no perf harness exists)
▸ GIVEN all tasks complete WHEN `go test -cover ./internal/tui/` runs THEN coverage is ≥ 40%

## Surprises

<!-- Empty; populated by realisera during execution when reality diverges from plan -->
