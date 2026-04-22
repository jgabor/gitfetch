# Plan: Split display package and add dual-decay layout

<!-- Level: full | Created: 2026-04-22 | Status: active -->
<!-- Reviewed: 2026-04-22 | Critic issues: 4 found, 4 addressed, 0 dismissed -->

## What

Split `internal/display` into focused packages and add dual-decay visualization. `core/` holds shared data (`RepoRow`, `BuildRows`). `theme/` holds visual styling (gradient bars, tier colors). `display/` becomes print-only. `tui/` owns its interactive table widget. Both print and TUI render commit and tag decay as compact side-by-side bars in a single "Decay" column.

## Why

Audit 2 flagged `display → bubbles` coupling: the print path imports `bubbles/table` unnecessarily. VISION promises "two-column decay" but only commit decay is shown today. This plan fixes the architecture and fulfills the vision.

## Constraints

- Startup must stay ~1 ms (`go-colorful` is already indirect; direct import adds zero weight).
- All existing tests pass.
- No CLI surface changes.
- `cmd/gitfetch/` import paths may update but behavior stays identical.

## Scope

**In**: `internal/display/`, `internal/tui/`, new `internal/core/`, new `internal/theme/`
**Out**: `internal/git/`, `internal/cache/`, `internal/config/`, `internal/decay/`
**Deferred**: Advanced gradient effects beyond basic color interpolation.

## Design

- `internal/core/` — Pure data (`RepoRow`, `BuildRows`, `repoColumnWidth`). Zero UI imports.
- `internal/theme/` — Visual styling (`GradientBar`, `TierColor`). Uses `lipgloss` + `go-colorful`.
- `internal/display/` — Print-only. Imports `core` + `theme`. Renders compact two-bar layout.
- `internal/tui/` — Owns interactive table widget. Imports `core` + `theme` + `bubbles/table`.

## Tasks

### Task 1: Extract `core/` package
**Depends on**: none
**Status**: □ pending
**Acceptance**:
▸ GIVEN `RepoRow`, `BuildRows`, `repoColumnWidth` WHEN moved to `internal/core/` THEN `display/` and `tui/` import them from `core/`
▸ GIVEN `go test ./internal/core/` WHEN run THEN `BuildRows` tests pass (1 pass + 1 fail per unit)
▸ GIVEN `go test ./...` WHEN run THEN all packages pass
▸ GIVEN `display_test.go` tests for `NewTable`/`TableStyles` WHEN examined THEN they are scheduled for migration to `tui/` in Task 4

### Task 2: Create `theme/` package
**Depends on**: Task 1
**Status**: □ pending
**Acceptance**:
▸ GIVEN `theme.GradientBar(tier, pct)` WHEN called with any `decay.Tier` THEN it returns a lipgloss-styled string
▸ GIVEN `theme.GradientBar` WHEN called with 0% or 100% progress THEN edge cases render correctly
▸ GIVEN `go test ./internal/theme/` WHEN run THEN coverage meets 1 pass + 1 fail per tier + edge cases

### Task 3: Refactor `display/` for print rendering
**Depends on**: Task 2
**Status**: □ pending
**Acceptance**:
▸ GIVEN `FormatDashboard` WHEN called with repos having both commit and tag data THEN both bars appear side by side in a single "Decay" column
▸ GIVEN `go list -f '{{.Imports}}' ./internal/display/` WHEN run THEN `bubbles/table` does not appear in imports
▸ GIVEN `go test ./internal/display/` WHEN run THEN test parity is maintained

### Task 4: Move TUI widget to `tui/`
**Depends on**: Task 2
**Status**: □ pending
**Acceptance**:
▸ GIVEN `tui.NewTable` WHEN called THEN it builds a `bubbles/table.Model` using `core/` and `theme/`
▸ GIVEN `go test ./internal/tui/` WHEN run THEN TUI tests pass
▸ GIVEN `go run ./cmd/gitfetch tui` WHEN launched THEN the interactive table renders correctly
▸ GIVEN `theme/` gradient bars WHEN rendered inside `bubbles/table` cells THEN ANSI sequences display correctly or a fallback plain path exists

### Task 5: Plan-level freshness checkpoint
**Depends on**: Task 3, Task 4
**Status**: □ pending
**Acceptance**:
▸ GIVEN plan completion WHEN documented THEN CHANGELOG.md lists the dual-decay layout and package split
▸ GIVEN plan completion WHEN documented THEN PROGRESS.md captures the aggregate outcome
▸ GIVEN plan completion WHEN archived THEN PLAN.md moves to `.agentera/archive/`

## Overall Acceptance

▸ GIVEN the default `gitfetch` command WHEN it prints the dashboard THEN both commit and tag decay bars are visible side by side
▸ GIVEN `go test ./...` WHEN run THEN all packages pass
▸ GIVEN `go build ./...` WHEN run THEN `internal/display/` does not import `bubbles/table`
▸ GIVEN the startup-latency harness WHEN run after all changes THEN mean execution time stays ≤ 2 ms

## Surprises

_(empty)_
