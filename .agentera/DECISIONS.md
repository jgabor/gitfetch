# Decisions

## Decision 1 · 2026-04-15

**Question**: product design of gitfetch, a repo decay tracker
**Context**: greenfield Go TUI project. scans git repos, measures decay, displays as dashboard. auto-runs on terminal open.
**Alternatives**:
- [single interactive session] : all-in-one TUI, no separation of concerns
- [background daemon + viewer] : over-engineered for a personal tool
**Choice**: configure + run model with cached display
**Reasoning**: the "neofetch for repos" framing settled it. human-only tool that must be instant on terminal open demands a cache-first architecture. `gitfetch` prints from cache and exits (for .bashrc). `gitfetch refresh` scans git and updates cache. `gitfetch tui` opens bubbletea checklist for curating tracked repos with inline refresh capability. decay shown as two columns (commit age, release age) with tiered color-coded progress bars: fresh/stale/decayed/dead. config in XDG_CONFIG_HOME/gitfetch/, cache in XDG_DATA_HOME/gitfetch/. standard Go layout: cmd/gitfetch/, internal/. TOML config. Charm ecosystem (bubbletea, lipgloss). no JSON output needed, human-only.
**Confidence**: firm
**Feeds into**: VISION.md

## Decision 2 · 2026-04-22

**Question**: how to structure `internal/display` so it can grow visually rich without coupling the print path to the TUI path or regressing startup
**Context**: `internal/display` serves both `FormatDashboard` (default `gitfetch` print) and `NewTable`/`TableStyles` (TUI interactive widget). This creates a `display → bubbles` coupling flagged in Audit 2. VISION demands "human-only output, color-coded progress bars" and sub-500ms startup (now ~1 ms). User wants gradients and richer visuals for both print and TUI.
**Alternatives**:
- [monolithic display] : keep `display` as-is; both print and TUI logic coexist. Simple, but coupling persists and TUI-specific imports leak into the print path.
- [split display + extract core data] : create `internal/core/` for `RepoRow`/`BuildRows`, `internal/theme/` for shared visual styles (lipgloss + go-colorful gradients), keep `display/` for print only, move TUI widget to `tui/`. Clean separation, shared visual language, zero new binary weight since `go-colorful` is already an indirect dependency.
- [shared theme, each mode owns rendering] : similar to above but `display/` and `tui/` each own their own table widget without a shared `theme/` package. Less consistency, more duplication.
**Choice**: split into `core/` + `theme/` + `display/` + `tui/`
**Reasoning**: `core/` isolates shared data so neither rendering package owns it. `theme/` guarantees visual consistency between print and TUI using `go-colorful` (already in tree, zero startup cost). Moving the TUI widget to `tui/` breaks the `display → bubbles` coupling and keeps the print path lean. This satisfies both the performance constraint and the visual ambition.
**Confidence**: provisional — gradient rendering inside `bubbles/table` cells needs validation, but the structure is sound
**Feeds into**: TODO.md

## Decision 3 · 2026-04-22

**Question**: how to improve the layout of the printed table and TUI for better visual scannability
**Context**: current layout shows only commit decay; VISION promises "two-column decay" (commit age + release age). user wants the overall structure and visual hierarchy improved.
**Alternatives**:
- [full columns for each decay dimension] : separate Commit Tier, Commit Bar, Commit Age, Tag Tier, Tag Bar, Tag Age columns. Maximum clarity, but table gets very wide.
- [compact inline text] : replace bars with colored text like "45d / 120d tag". Very compact but loses the intuitive progress-bar visual.
- [compact side-by-side bars] : single "Decay" column containing two bars (commit + tag) side by side or stacked. Balances clarity and width.
**Choice**: compact side-by-side bars in a single "Decay" column
**Reasoning**: flat list sorted by commit decay (most decayed first) keeps the current mental model. repo name first anchors the row. a single "Decay" column with two bars (commit above, tag below, or side by side) fulfills the VISION's two-column promise without excessive width. `theme/` package generates both bars so print and TUI stay visually consistent.
**Confidence**: firm — straightforward rendering change using existing `RepoRow` data
**Feeds into**: TODO.md
