# Documentation Contract

<!-- Maintained by dokumentera. Last audit: 2026-04-17 -->

## Conventions

```
doc_root: .           # VISION.md, CHANGELOG.md, TODO.md at root; skill artifacts in .agentera/
style:    terse-technical, no badges, sparse prose matches VISION.md voice
auto_gen: none
```

No versioning block: Go modules version via git tags; no tags exist yet.

## Artifact Mapping

| Artifact | Path | Producers |
|----------|------|-----------|
| VISION.md | VISION.md | visionera |
| CHANGELOG.md | CHANGELOG.md | realisera, dokumentera |
| TODO.md | TODO.md | inspektera, realisera |
| DOCS.md | .agentera/DOCS.md | dokumentera |
| PROGRESS.md | .agentera/PROGRESS.md | realisera |
| DECISIONS.md | .agentera/DECISIONS.md | resonera |
| HEALTH.md | .agentera/HEALTH.md | inspektera |
| PLAN-*.md | .agentera/ (active) or .agentera/archive/ (completed) | planera |

## Index

| Document | Path | Last Updated | Status |
|----------|------|-------------|--------|
| VISION | VISION.md | 2026-04-15 | ■ current |
| CHANGELOG | CHANGELOG.md | 2026-04-17 | ■ current |
| TODO | TODO.md | 2026-04-17 | ■ current |
| PROGRESS | .agentera/PROGRESS.md | 2026-04-17 | ■ current |
| DECISIONS | .agentera/DECISIONS.md | 2026-04-15 | ■ current |
| HEALTH | .agentera/HEALTH.md | 2026-04-17 | ■ current |
| README | README.md | 2026-04-28 | ■ current |
| CLAUDE | CLAUDE.md | — | □ missing |

## Audit log

- **2026-04-17** — first-run survey. Established root-level doc layout, terse-technical style, no auto-generated docs. Synced CHANGELOG.md and PROGRESS.md after drift from uncommitted TUI/scanner feature work. README and CLAUDE.md deferred.
- **2026-04-17** — Audit 2 remediation plan completed (7/7 tasks). Tasks 1-6 shipped feature commits; Task 7 = this checkpoint. CHANGELOG.md + PROGRESS.md + TODO.md all current. Active PLAN.md archived to `.agentera/archive/PLAN-2026-04-17-audit2-remediation.md`.
- **2026-04-17** — Audit 3 remediation plan completed (7/7 tasks). Tasks 1-6 shipped feature commits clearing the Audit 3 warning (`decay.DaysUntilNext`) plus five cosmetic info items (help text, `filteredCache` wrapper, `discoverRepos` shim, `stripANSI`, tui.go linter modernizations); Task 7 = this checkpoint. CHANGELOG.md + PROGRESS.md + TODO.md all current. Active PLAN.md archived to `.agentera/archive/PLAN-2026-04-17-audit3-remediation.md`.
