# Plan: Audit 3 Remediation

<!-- Level: full | Created: 2026-04-17 | Status: active -->
<!-- Reviewed: 2026-04-17 | Critic issues: 10 found, 6 addressed, 4 dismissed -->

## What
Address all new findings from HEALTH.md Audit 3 (2026-04-17): one warning (unused `decay.DaysUntilNext` dragging package coverage from 92.6% → 77.4%) and four cosmetic info items (TUI help text `k`/`j` typo, dead `filteredCache` wrapper, redundant unexported `discoverRepos` shim, unused `stripANSI` test helper), plus three pre-existing TUI linter hints that were parked in TODO Annoying during the Audit 2 plan (`slices.Contains`, built-in `max`, `fmt.Fprintf`).

## Why
Audit 3's trajectory is ⮉ improving and the findings are small, but leaving them in place lets coverage slip silently, confuses readers (`↓/k`, the case-collision pair), and grows the "annoying" backlog. These are one-sitting cleanups that restore the decay coverage floor and clear the TUI lint surface before any new feature work lands.

Two Audit 3 items are explicitly deferred: `handleDiscovering` complexity (HEALTH: "defer unless the mode grows further") and the `display → bubbles` coupling question (belongs to `/resonera`, already tracked in TODO Annoying). Task 7 surfaces both to TODO.md.

## Constraints
- Each task = one commit, ≤10 LOC of code change for most (Task 6 may exceed modestly).
- No behavior changes visible to the user beyond the corrected help string.
- VISION constraints preserved: cache-first sub-500ms print, no daemon, no network.
- `go test ./...` and `go build ./...` must stay green after every task.
- No versioning bump: DOCS.md has no `versioning` block.

## Scope
**In**: deletion of `decay.DaysUntilNext`; tui.go help-text fix, wrapper inline, linter modernizations; scanner unexported helper inline; test-helper deletion; plan-level freshness checkpoint.
**Out**: `handleDiscovering` decomposition; `display → bubbles` coupling deliberation; external dependency upgrades; new features.
**Deferred**: the two Out items are filed to TODO.md by Task 7.

## Design
Each task is a focused edit in one or two files:

- **Task 1** removes dead code in `internal/decay`; restores coverage by subtracting unreachable statements from the denominator.
- **Task 2** fixes a one-character typo in the normal-mode help render.
- **Task 3** removes an indirection in `internal/tui`: the wrapper and both tests retarget `cache.FilterByRepos` directly.
- **Task 4** inlines a 6-line error-swallowing shim in `internal/git/scanner.go`; `ResolveRepoPaths` becomes a two-line loop that calls `DiscoverRepos` and falls back on error.
- **Task 5** deletes an unused test helper in `internal/display`.
- **Task 6** applies three idiom updates in `internal/tui/tui.go` behind the existing Update-driven tests.
- **Task 7** sweeps plan-level artifacts and archives PLAN.md.

Tasks 1, 4, 5 touch disjoint files from Tasks 2, 3, 6 and are independent. Tasks 2, 3, 6 all touch `internal/tui/tui.go`, so Task 6 depends on 2 and 3 to avoid rebasing.

## Tasks

### Task 1: Delete `decay.DaysUntilNext`
**Depends on**: none
**Status**: ■ complete
**Acceptance**:
▸ GIVEN the `internal/decay` package WHEN `grep -rn DaysUntilNext .` runs after the change THEN there are zero matches anywhere in the repo
▸ GIVEN `go test -cover ./internal/decay/` WHEN it runs THEN reported coverage is ≥92% of statements (Audit 1 baseline was 92.6%)
▸ GIVEN `go build ./...` and `go test ./...` WHEN they run THEN both succeed with no errors

### Task 2: Correct TUI help text so `↓` pairs with `j`
**Depends on**: none
**Status**: □ pending
**Acceptance**:
▸ GIVEN the normal-mode help line emitted by `View()` WHEN rendered THEN the down-arrow `↓` is paired with the letter `j` (not `k`); the up-arrow `↑` remains paired with `k`
▸ GIVEN `go test ./internal/tui/` WHEN it runs THEN all tests pass

### Task 3: Remove the `filteredCache` indirection in tui
**Depends on**: none
**Status**: □ pending
**Acceptance**:
▸ GIVEN the `internal/tui` package WHEN searched THEN no `filteredCache` function is defined
▸ GIVEN all prior callers of `filteredCache` (production + tests) WHEN inspected THEN each calls `cache.FilterByRepos` directly with the `*Cache` and `[]string` in the order `FilterByRepos` requires
▸ GIVEN `go test ./internal/tui/` WHEN it runs THEN all tests pass (existing `filteredCache` test cases retargeted, not deleted)
▸ GIVEN `go build ./...` WHEN it runs THEN it succeeds

### Task 4: Inline the unexported `discoverRepos` shim in scanner
**Depends on**: none
**Status**: □ pending
**Acceptance**:
▸ GIVEN `internal/git/scanner.go` WHEN read THEN the unexported `discoverRepos` helper (the error-swallowing wrapper around `DiscoverRepos`) no longer exists; `ResolveRepoPaths` calls `DiscoverRepos` directly and falls back to `[]string{path}` on error inline
▸ GIVEN the exported `DiscoverRepos` WHEN inspected THEN its signature and behavior are unchanged
▸ GIVEN `go test ./internal/git/` WHEN it runs THEN all tests pass including `ResolveRepoPaths` cases (unreadable-dir fallback still works)
▸ GIVEN `go build ./...` WHEN it runs THEN it succeeds

### Task 5: Delete unused `stripANSI` test helper
**Depends on**: none
**Status**: □ pending
**Acceptance**:
▸ GIVEN `staticcheck ./internal/display/` WHEN it runs THEN there is no `U1000: func stripANSI is unused` report
▸ GIVEN `go test ./internal/display/` WHEN it runs THEN all tests pass

### Task 6: Modernize tui.go linter hints
Apply three idiom updates in `internal/tui/tui.go`:
(a) `slices.Contains` for the linear loop in the duplicate-tracker / duplicate-check path
(b) the built-in `max()` in place of `if a < b { a = b }` patterns (two sites)
(c) `fmt.Fprintf(&b, ...)` in place of `b.WriteString(fmt.Sprintf(...))` sites

**Depends on**: Task 2, Task 3
**Status**: □ pending
**Acceptance**:
▸ GIVEN `staticcheck ./internal/tui/` WHEN it runs THEN none of the three pre-existing hints (slices.Contains-for-loop, minmax, WriteString+Sprintf) are reported; no new hints introduced
▸ GIVEN `go vet ./...` WHEN it runs THEN it reports no issues
▸ GIVEN `go test ./internal/tui/` WHEN it runs THEN all tests pass (Update-driven tests exercise the changed paths)
▸ GIVEN `go build ./...` WHEN it runs THEN it succeeds

### Task 7: Plan-level freshness checkpoint
**Depends on**: Task 1, Task 2, Task 3, Task 4, Task 5, Task 6
**Status**: □ pending
**Acceptance**:
▸ GIVEN CHANGELOG.md WHEN read THEN it has a dated entry summarizing the Audit 3 remediation (Tasks 1-6) with commit hashes
▸ GIVEN `.agentera/PROGRESS.md` WHEN read THEN it has a plan-level cycle entry (in addition to the per-task entries) referencing the six feature commits
▸ GIVEN TODO.md WHEN read THEN (1) the three resolved linter hints are moved from Annoying to Resolved, (2) a new Annoying entry exists for `handleDiscovering` complexity with a note that HEALTH Audit 3 already tracks the deferral rationale, (3) the existing `display → bubbles` entry is preserved
▸ GIVEN `.agentera/DOCS.md` WHEN read THEN the Audit log has a new entry for 2026-04-17 Audit 3 remediation
▸ GIVEN the plan completes WHEN finalization commit lands THEN `.agentera/PLAN.md` is archived to `.agentera/archive/PLAN-2026-04-17-audit3-remediation.md` and the active `.agentera/PLAN.md` is deleted

## Overall Acceptance
▸ GIVEN the repo after all tasks WHEN `go test ./...` runs THEN all packages pass; decay coverage ≥92%, display/cache/git coverage not worse than Audit 3 baseline, tui coverage still ≥58%
▸ GIVEN `staticcheck ./...` WHEN it runs THEN the four Audit 3 / TODO-Annoying hints (stripANSI, slices.Contains, minmax, WriteString+Sprintf) are absent
▸ GIVEN `go run ./cmd/gitfetch` WHEN run against the existing cache THEN the rendered dashboard is byte-identical to before the plan (no user-visible behavior change except the corrected help line inside the TUI)
▸ GIVEN HEALTH.md Audit 3 findings WHEN cross-checked after Task 7 THEN every item is resolved, deferred-to-TODO, or explicitly deferred-per-HEALTH

## Surprises
<!-- Populated by realisera during execution when reality diverges from plan -->
