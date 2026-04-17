# Health

## Audit 1 · 2026-04-15

### Architecture alignment
Grade: A

All seven declared packages exist and map exactly to the stated architecture: `cmd/gitfetch` (CLI entry point), `internal/config` (TOML config + XDG paths), `internal/git` (repo scanner via git CLI), `internal/cache` (JSON file cache), `internal/decay` (tier classification logic), `internal/display` (dashboard formatting with lipgloss), `internal/tui` (bubbletea interactive mode). Dependency direction is clean: `cmd` → `{config, cache, display, tui}`; `tui` → `{config, cache, display, decay, git}`; `display` → `{cache, decay}`. Core domain packages (`decay`, `cache`) have zero inbound UI/framework imports. Each package is a single source file with a paired test file — appropriate for the current size. The `decay` package is fully decoupled and could be extracted as a library with no changes.

### Pattern consistency
Grade: A-

Error handling follows a uniform `fmt.Errorf("context: %w", err)` wrapping pattern across all packages. Naming is consistent: PascalCase exports, camelCase internals, `ScanResult`/`RepoEntry` struct field names match JSON/TOML tag conventions. Each package exposes a small, focused API surface (2-5 exported functions/types).

One minor concern: `tui/tui.go:51` uses a package-level mutable var `inputBuffer` instead of storing it on the `model` struct. This breaks the bubbletea model-immutable convention and would cause data races if multiple Program instances existed. Not a bug today (single-process TUI), but inconsistent with the rest of the codebase's clean state management. The `Label()` method on `decay.Tier` is a redundant alias for `String()` with no differentiation.

### Test health
Grade: A-

All tests pass. Coverage by package: `decay` 92.6%, `display` 96.7%, `git` 86.4%, `config` 80.8%, `cache` 78.3%, `tui` 27.6%, `cmd/gitfetch` 0%. Production LOC: 914, test LOC: 1154 — a 1.26:1 test-to-code ratio. Tests cover happy paths, error paths, boundary values (tier thresholds at ±1 day), roundtrip serialization, and edge cases (empty repos, missing files, nil dates). The `display` tests include a hand-rolled ANSI stripper for visual assertions.

Two gaps: `tui` at 27.6% only tests `NewModel` and `View()` rendering — no `Update()` key-handling tests, no `removeRepo`/`handleAdding`/`handleScanDone` coverage. The bubbletea `tea.KeyMsg` interface makes unit testing `Update` slightly awkward but not impossible. `cmd/gitfetch` at 0% is acceptable (thin CLI wiring), though a single integration test validating the cobra command tree would be cheap insurance.

## Overall
Grade: A-

## Audit 2 · 2026-04-17

**Dimensions assessed**: architecture, patterns, coupling, complexity, tests, deps, artifact freshness, security
**Findings**: 1 critical, 6 warnings, 3 info
**Overall trajectory**: ⮋ degrading vs Audit 1 (large uncommitted feature delta without docs/tests catching up)
**Grades**: Architecture [A] | Patterns [B] | Coupling [B] | Complexity [B] | Tests [C] | Deps [A] | Security [A] | Artifact freshness [C]

### Tests: C (⮋ from A-)

#### ⇶ TUI 'r' key no longer refreshes — startScan is dead code (confidence: 95/100)
- **Location**: `internal/tui/tui.go:150-151`, `internal/tui/tui.go:285-291`
- **Evidence**: `handleNormalKeys` case `"r"` sets `m.mode = modeNormal` and nothing else. `startScan` is defined at line 285 but `grep -r startScan` finds only the definition — zero callers. Help text line 399 still advertises "r refresh".
- **Impact**: Core TUI feature silently broken. Users press r, see no error, assume refresh happened; cache goes stale. Tests don't catch it because `Update` is untested.
- **Suggested action**: wire `startScan(m)` into the `r` branch and return the returned `tea.Cmd`. Add an `Update` test that sends `tea.KeyMsg{Runes:[]rune{'r'}}` and asserts a non-nil cmd.

#### ⇉ TUI coverage regressed to 14.8% while code tripled (confidence: 90/100)
- **Location**: `internal/tui/tui_test.go` (81 lines), `internal/tui/tui.go` (493 lines)
- **Evidence**: prior audit noted tui at 27.6% with `Update` untested; file grew from ~160 to 493 lines (discovery mode, add flow, scan handler, visibleRange) but `tui_test.go` is unchanged — still only NewModel + View smoke tests. Coverage fell to 14.8%.
- **Impact**: the regression above is a direct consequence. `handleAdding`, `handleDiscovering`, `handleScanDone`, `removeRepo`, `visibleRange` all uncovered.
- **Suggested action**: table-driven tests driving `Update` with synthesized `tea.KeyMsg`. Pure helpers (`visibleRange`, `filteredCache`) are trivial unit-test targets — start there.

#### ⇉ Scanner coverage dropped despite new surface area (confidence: 75/100)
- **Location**: `internal/git/scanner.go`
- **Evidence**: `DiscoverRepos`, `ResolveRepoPaths`, `listRemotes`, `matchFilter`, `sanitizeGitError` are new; coverage went 86.4% → 83.8%. `listRemotes` returns its error but no test exercises the error path; `matchFilter` and `DiscoverRepos` have no direct tests.
- **Impact**: silent breakage risk in repo discovery / filter logic.
- **Suggested action**: unit tests for `DiscoverRepos` (single-repo, dir-of-repos, bare dir, unreadable dir) and `matchFilter` (case-insensitivity, substring, empty).

### Patterns: B (⮋ from A-)

#### ⇉ Scanner always fetches full author history even when no filter is set (confidence: 85/100)
- **Location**: `internal/git/scanner.go:35-43`
- **Evidence**: `ScanRepo` unconditionally calls `listAuthors` (which runs `git log --format=%aN` across the whole history) before checking `opts.Author`. The result is stored on `ScanResult.Authors` but `cache.RepoEntry` has no `Authors` field — it's discarded after every refresh.
- **Impact**: violates the "cache-first, sub-500ms" principle. For repos with long histories, refresh now pays an O(commits) cost every time for data that's thrown away. Same story for `listRemotes` (cheaper, but still unnecessary in unfiltered runs).
- **Suggested action**: guard `listAuthors`/`listRemotes` with `if opts.Author != ""` / `if opts.Remote != ""`. Or persist Authors/Remotes to cache if they're intended to be reusable.

#### ⇉ `cmd/gitfetch` and `tui` duplicate the same filtered-cache loop (confidence: 80/100)
- **Location**: `cmd/gitfetch/main.go:35-44`, `internal/tui/tui.go:69-81`
- **Evidence**: identical repoSet construction + map filtering, byte-for-byte equivalent. `tui.filteredCache` is unexported, so main.go can't reuse it.
- **Impact**: DRY violation; two places to fix if the filtering rule evolves (e.g. when cache reconciliation per VISION is implemented).
- **Suggested action**: move `filteredCache` to `internal/cache` (exported `FilterByRepos`) or `internal/display` and call it from both sites.

### Artifact freshness: C (new dimension)

#### ⇉ PROGRESS.md and CHANGELOG.md are stale relative to the working tree (confidence: 85/100)
- **Location**: `.agentera/PROGRESS.md`, `CHANGELOG.md`, working tree diff
- **Evidence**: `git diff HEAD --stat` shows 788+/392- across 13 files (TUI discovery mode, author/remote filter flags, scanner author/remote collection, display table rewrite). PROGRESS.md's newest entry is Cycle 6 — "plan complete, all tasks shipped" — none of the new work is logged. CHANGELOG "Added" list has no entries for `--author`/`--remote` flags, repo discovery checklist, or table-based TUI.
- **Impact**: realisera / resonera lose ground truth about recent evolution; the audit above had to reconstruct it from `git log` + diff. If these changes get committed without updating docs, the drift compounds.
- **Suggested action**: either commit the working tree with a PROGRESS/CHANGELOG sweep in the same commit, or run `/dokumentera` before the next `/realisera` cycle.

### Complexity: B

#### ⇢ `handleAdding` mixes input handling, path validation, discovery dispatch, and persistence (confidence: 60/100)
- **Location**: `internal/tui/tui.go:161-224`
- **Evidence**: 63-line function doing Enter-path validation, `os.Stat`, `DiscoverRepos`, dedup check, `config.Save`, mode transitions. Three failure modes, two success modes, shared `m.inputBuffer` reset in four places.
- **Impact**: hard to unit-test without decomposition (see coverage finding).
- **Suggested action**: extract `commitNewRepo(m *model, path string) tea.Cmd` returning the resolved next mode + status, keep `handleAdding` as a pure keystroke router.

### Patterns (minor, repeat)

#### ⇢ `decay.Tier.Label()` is still a redundant alias for `String()` (confidence: 90/100)
- **Location**: `internal/decay/decay.go:37-39`
- **Evidence**: unchanged since Audit 1. Both return identical strings. `Label` is called in `display.rowToTableRow` but could use `String()`.
- **Impact**: cosmetic / API surface cruft.
- **Suggested action**: delete `Label()`, switch the single caller to `String()`.

### Coupling: B

#### ⇢ `internal/display` now depends on `bubbles/table` and `bubbles/key` (confidence: 55/100)
- **Evidence**: display used to be a pure formatter (lipgloss only). It now constructs `table.Model` and rebinds `KeyMap`. `cmd/gitfetch` pulls bubbles transitively through display even for the non-interactive default command.
- **Impact**: muddies the layering sketched in the Audit 1 architecture grade — `display` was described as a leaf. Not actually harmful (bubbles is already in the tree via tui), but the conceptual boundary shifted without being acknowledged.
- **Suggested action**: optional — either accept the shift (update VISION "architecture" sketch if it exists) or split into `display/format` (pure) and `display/table` (bubbles-aware).

### Security: A
Regex scan for `password|secret|token|api_key|AKIA|ghp_|xox[bp]-|sk-|BEGIN.*PRIVATE` found no matches in source. No `eval` / dynamic `exec` constructions. `exec.Command` uses fixed argv forms, no shell strings.

> This is a lightweight surface scan. For comprehensive security analysis, use dedicated tools: semgrep, Snyk, govulncheck, or similar static analysis and vulnerability scanning tools appropriate to your stack.

### Deps: A
`go.mod` is tidy. All direct deps pinned to concrete versions. No outdated-dep scan run (no offline cache), but versions are recent (bubbletea 1.3.10, lipgloss 1.1.0, cobra 1.10.2).

### Architecture: A
Package graph unchanged; no new packages. Dependency direction still clean: `cmd → {config, cache, display, tui, git}`, `tui → {cache, config, display, git}`, `display → {cache, decay, bubbles}`. The display→bubbles widening is the one drift (see Coupling above).

### Trends vs Audit 1
- **Improved**: `inputBuffer` is now a field on `model` (line 55), resolving the Audit 1 package-level-var concern.
- **Degraded**: TUI test coverage (27.6%→14.8%), scanner coverage (86.4%→83.8%), decay coverage (92.6%→78.1%). Refresh performance regressed via unconditional `listAuthors`. A real feature ('r' key) went silently dead.
- **New findings**: broken `r`, dead `startScan`, duplicated filter loop, stale PROGRESS/CHANGELOG, wasted author scan.
- **Resolved**: the `inputBuffer` package-level var.

### Patterns Observed
- **Module structure**: one `*.go` + one `*_test.go` per package holds for everything except `tui.go` — the one file that's outgrowing the convention (493 lines, 16 functions).
- **Error handling**: uniform `fmt.Errorf("ctx: %w", err)`; `sanitizeGitError` is the one exception and is thoughtful — it prefers stdout, then stderr, then the raw exit error.
- **Testing approach**: table-driven where thresholds exist (decay), integration-style with `exec.Command` against real git (scanner), snapshot-ish ANSI comparison (display). TUI tests are shape-only ("view is non-empty").
- **Dependency patterns**: direct deps pinned; indirects let go-mod resolve. No vendor dir. No lockfile pressure.
- **Shelling out**: scanner always exec's git; there is no go-git dependency and no plan to add one per VISION ("shell out to git, cache results, display tiers").

A clean, well-structured first release. Architecture is textbook Go project layout with clear package boundaries and no circular dependencies. Error handling and naming are consistent throughout. Test coverage is strong for domain logic (decay, display, git, config, cache) but the TUI layer is undertested — the only actionable gap. The `inputBuffer` package-level var is a minor code smell worth fixing before the TUI grows more state. No critical findings; project is in excellent shape for its first iteration.
