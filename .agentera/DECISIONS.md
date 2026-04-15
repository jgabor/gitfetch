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
