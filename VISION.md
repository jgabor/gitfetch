# gitfetch

## North Star

gitfetch makes repo neglect visible at a glance — the moment a terminal opens, you know which projects need attention. It brings the instant, glanceable philosophy of neofetch to repository health: zero friction, zero waiting, just a color-coded snapshot of your code garden.

## Who It's For

- **The poly-repo developer** — juggling dozens of checkouts across work, side projects, and open source. Needs to see at a glance which repos have gone stale.
- **The maintainer** — tracking whether personal projects are getting attention or quietly rotting.
- **The terminal dweller** — lives in the shell, wants tools that integrate into existing workflows (`.bashrc`, `.zshrc)`, not browser dashboards.

## Principles

- **Cache-first, always** — the default command reads a local cache and exits. No git subprocesses, no network, no delay. Sub-500ms guarantee for `.bashrc` use.
- **Three modes, one tool** — `gitfetch` prints from cache, `gitfetch refresh` scans and updates cache, `gitfetch tui` manages repos interactively. Each mode does one thing well.
- **Human-only output** — no JSON, no API, no machine consumers. The terminal is the interface. Commit activity sparklines and release age bars pair visual signals with exact ages.
- **Two-column decay and activity** — commit age and release age shown side by side. A repo can be actively committed but unreleased, or released but unmaintained. Both dimensions matter. Cached weekly activity adds development history without slowing the default command.
- **XDG conventions** — config in `XDG_CONFIG_HOME/gitfetch/`, cache in `XDG_DATA_HOME/gitfetch/`. Standard paths, no surprises.

## Direction

gitfetch is a mature personal tool with its core architecture settled. Future evolution focuses on:

- **Configurable thresholds** — letting users tune tier boundaries (fresh/stale/decayed/dead) per repo or globally via TOML config.
- **Smarter tag patterns** — supporting configurable tag patterns beyond `v*` for projects using different release conventions.
- **Cache reconciliation** — gracefully handling config changes (added/removed repos) without requiring a full rescan.

The architecture is intentionally simple: shell out to git, cache results, display tiers. No daemon, no database, no network layer. This simplicity is the feature.

## Identity

gitfetch is opinionated but unobtrusive. It doesn't nag — it shows. It doesn't run in the background — it prints and exits. It trusts the terminal as the right place for developer tools. Sparse, colorful, honest.
