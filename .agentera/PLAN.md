# Plan: TUI UX Polish

<!-- Level: full | Created: 2026-04-23 | Status: active -->
<!-- Reviewed: 2026-04-23 | Critic issues: 5 found, 2 addressed, 3 dismissed -->

## What

Fix 17 UI/UX issues in the TUI: byte-vs-display-width bugs that break on non-ASCII repo names, missing confirmation before destructive actions, overloaded keybindings across modes, absent scan progress feedback, persistent error states, terminal compatibility gaps, and a hardcoded-width gradient bar.

## Why

The TUI is the primary human interface. Byte-width truncation breaks rendering for multibyte repo names. Immediate deletion on `d`/`x` risks accidental data loss. No scan feedback makes the UI feel frozen. Error messages never clear. These erode trust and usability.

## Constraints

- Must not break existing test suite
- Must preserve module boundaries (core/theme/display/tui)
- Must not add new dependencies (runewidth is already an indirect dep via lipgloss)
- Must stay within bubbletea/lipgloss/bubbles v2 APIs

## Scope

**In**: All 17 identified issues grouped into 7 task clusters
**Out**: Light/dark theme adaptation (flagged for TODO.md), full undo stack, async partial scan results
**Deferred**: Light/dark terminal adaptation (design system question, not a bug fix)

## Design

Seven task clusters touch four modules. Byte-width fixes in `tui` and `table` use the existing `runewidth` library already on the dependency graph. The confirmation mode extends the existing mode machine (`normal → adding / discovering`) with a `confirming` state that captures the pending action and target repo. Keybinding cleanup centralizes the help string into a single source keyed by mode. Scan feedback passes a counter through the bubbletea `tea.Msg` pipeline. Error lifecycle uses a `tea.Tick` timer to auto-clear. Terminal robustness moves `AltScreen` to `Init()`, replaces the em-dash with a safe ASCII separator, and handles the blink cursor with a visible fallback. The gradient bar adapts to terminal width using the bubbletea window-size message.

## Tasks

### Task 1: Display-width correctness
**Depends on**: none
**Status**: ■ complete
**Acceptance**:
▸ GIVEN a repo name containing multibyte characters WHEN the discover list renders THEN the name column aligns correctly without overflow or wrapping
▸ GIVEN a repo name exceeding the discover column width WHEN the list renders THEN truncation uses display width not byte length
▸ GIVEN a repo name with multibyte characters WHEN truncatePlain runs in table.go THEN truncation respects display width
▸ Test proportionality: 1 pass + 1 fail per truncation function (discover path formatting, truncatePlain)

### Task 2: Terminal robustness
**Depends on**: none
**Status**: ■ complete
**Acceptance**:
▸ GIVEN the TUI starts WHEN Init() runs THEN AltScreen is set in the init command, not in View()
▸ GIVEN a terminal that disables ANSI blink WHEN add-path mode renders THEN the cursor is visible as a static block rather than invisible
▸ GIVEN a non-UTF8 locale WHEN the header renders THEN it uses an ASCII-safe separator instead of em-dash
▸ GIVEN availableLines is 0 in discover mode WHEN visibleRange runs THEN it returns an empty range without panicking
▸ Test proportionality: 1 pass + 1 fail per fix (AltScreen placement, cursor fallback)

### Task 3: Keybinding cleanup and help expansion
**Depends on**: Task 1
**Status**: ■ complete
**Acceptance**:
▸ GIVEN normal mode WHEN the help bar renders THEN it lists all active keybindings including scroll keys (ctrl+u, ctrl+d, pgup, pgdown, home, end)
▸ GIVEN discover mode WHEN the user presses a key used in normal mode with different meaning THEN only the discover-mode action fires
▸ GIVEN discover mode WHEN the help bar renders THEN it shows discover-specific bindings distinct from normal mode
▸ GIVEN discover mode WHEN the user presses a normal-mode-only key (e.g., r for refresh) THEN no normal-mode action fires
▸ Test proportionality: 1 pass + 1 fail per mode (pass: correct help content; fail: wrong-mode key ignored)

### Task 4: Confirmation mode for destructive actions
**Depends on**: Task 3
**Status**: ■ complete
**Acceptance**:
▸ GIVEN normal mode with a selected repo WHEN the user presses d or x THEN the UI enters a confirming state showing the repo path and pending action, without deleting the repo
▸ GIVEN confirming mode WHEN the user presses y or enter THEN the repo is removed and the mode returns to normal
▸ GIVEN confirming mode WHEN the user presses n, escape, or q THEN the mode returns to normal with no change
▸ GIVEN confirming mode WHEN the user presses any other key THEN the UI stays in confirming mode with no side effect
▸ Test proportionality: 1 pass + 1 fail per confirming transition (confirm, cancel, reject invalid key)

### Task 5: Scan feedback and empty states
**Depends on**: Task 1
**Status**: ■ complete
**Acceptance**:
▸ GIVEN a scan is running WHEN each repo completes THEN the UI shows a counter (e.g., "scanning 3/12…") instead of static "scanning…"
▸ GIVEN the scan command WHEN it runs THEN it emits per-repo progress messages rather than blocking until all repos are scanned
▸ GIVEN repos exist in config but cache is empty WHEN the table renders THEN the UI shows a message prompting the user to press r to refresh
▸ GIVEN no repos are tracked WHEN the table renders THEN the existing "No repos tracked" message appears
▸ Test proportionality: 1 pass per state (scanning with counter, empty cache, empty config)

### Task 6: Error lifecycle and visual adaptiveness
**Depends on**: Task 2
**Status**: ■ complete
**Acceptance**:
▸ GIVEN an error is displayed WHEN the next successful action completes THEN the error message clears
▸ GIVEN an error is displayed WHEN 5 seconds elapse without user action THEN the error message auto-clears
▸ GIVEN a terminal wider than 30 characters WHEN the gradient bar renders THEN it fills available width up to the content area
▸ GIVEN a terminal narrower than 30 characters WHEN the gradient bar renders THEN it shrinks to fit without horizontal overflow
▸ Test proportionality: 1 pass + 1 fail per behavior (error clear on action, error auto-clear, bar width adaptation)

### Task 7: Plan-level freshness checkpoint
**Depends on**: Tasks 1, 2, 3, 4, 5, 6
**Status**: ■ complete
**Acceptance**:
▸ GIVEN all prior tasks are complete WHEN the checkpoint runs THEN CHANGELOG.md has an entry summarizing the TUI UX polish work
▸ GIVEN all prior tasks are complete WHEN the checkpoint runs THEN PROGRESS.md reflects the completed plan
▸ GIVEN all prior tasks are complete WHEN the checkpoint runs THEN TODO.md flags light/dark terminal adaptation as a future item

## Overall Acceptance
▸ GIVEN a repo with multibyte name WHEN rendered in any TUI mode THEN all text aligns and truncates by display width
▸ GIVEN the user presses d on a repo WHEN the confirmation prompt appears THEN the repo is not deleted until explicit confirmation
▸ GIVEN a scan runs with N repos WHEN progress updates THEN the counter reflects current progress
▸ GIVEN an error occurs WHEN 5 seconds pass or the user acts THEN the error clears
▸ GIVEN any terminal width WHEN the UI renders THEN no horizontal overflow occurs

## Surprises
