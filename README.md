# gitfetch

Tracks which of your repos you've been neglecting. Like neofetch, but for commit and release decay.

## Install

Go 1.26.2+, or build with Mage.

```
git clone https://github.com/jgabor/gitfetch
cd gitfetch
mage install          # builds and copies to $GOBIN
```

Plain Go:

```
go install github.com/jgabor/gitfetch/cmd/gitfetch@latest
```

## Usage

```
gitfetch                           # dashboard from cache (instant)
gitfetch refresh                   # rescan all tracked repos
gitfetch refresh --author "jgabor" --remote "github"
gitfetch tui                       # interactive repo manager
gitfetch --verbose                 # show legend and full scan errors
```

| Command            |                                                                                                                                                       |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `gitfetch`         | Reads cache, prints dashboard, exits. No git subprocesses, sub-500ms, `.bashrc`-safe. Auto-refreshes every N invocations when `refresh_every` is set. |
| `gitfetch refresh` | Scans every tracked repo concurrently (bounded by physical cores). Supports `--author` and `--remote` filter flags. Writes results to cache.          |
| `gitfetch tui`     | Bubbletea TUI with vim keys. Add/remove repos, manual refresh, multi-repo discovery on directory paths.                                               |

### Refresh filters

| Flag             | Description                                                                 |
| ---------------- | --------------------------------------------------------------------------- |
| `--author <str>` | Only scan repos whose author list contains the substring (case-insensitive) |
| `--remote <str>` | Only scan repos whose remote URL contains the substring (case-insensitive)  |

### TUI keybindings

| Key               | Action                                                                                                                                                                        |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `↑` `↓` / `j` `k` | Move cursor (vim-style)                                                                                                                                                       |
| `r`               | Refresh (scan all tracked repos)                                                                                                                                              |
| `a`               | Add a repo path — if the path is a directory with multiple `.git` subdirectories, a checklist appears to toggle repos with `space` (or `a` to toggle all, `enter` to confirm) |
| `d` / `x`         | Remove selected repo (with confirmation)                                                                                                                                      |
| `q` / `ctrl+c`    | Quit                                                                                                                                                                          |

## Configuration

`$XDG_CONFIG_HOME/gitfetch/config.toml`, auto-created on first run:

```toml
repos = []
refresh_every = 0    # auto-rescan every N invocations (0 = never)

[display]
color = "auto"       # auto, always, never
width = 0            # max dashboard width (0 = auto-detect terminal)
```

Populate `repos` via `gitfetch tui` (`a` to add) or by editing the file. Paths support `~/` expansion. Removed repos are pruned from cache automatically.

## Decay dashboard

Each repo occupies one row with commit activity and age, release age, and version.
An eight-character release bar shows age on a 180-day scale. Longer bars mean older
releases; exact commit and release ages remain visible without color. A summary shows commit-tier counts,
errors, unknown dates, and the range of cache scan ages.

At 64 columns and above, the Commit column pairs an activity sparkline with the
exact commit age. Its color matches commit freshness. The sparkline shows commits in eight completed
UTC weeks (Monday to Monday), oldest first. All repos share a linear count scale.
`·` means zero commits; `—` means unavailable data. Counts cover commits reachable
from HEAD and use committer dates. The current incomplete week is excluded.
Older cached weeks align to the current window, with unscanned weeks marked
unavailable. Run `gitfetch refresh` once to populate activity in an older cache.
`--verbose` shows the scale maximum and full scan errors.

Below 64 columns, the sparkline and release bar are omitted while ages remain. Very narrow layouts progressively omit version, release, and commit
columns to preserve the repo name. Long names and versions are truncated to fit.
When stdout is not a terminal, the default width is 80; `display.width` caps it.

The TUI uses the same columns and scale. Details beneath the selected row show
the full path, version, timestamps, displayed activity period, and scan error.

| Tier    | Threshold   | Bar    |
| ------- | ----------- | ------ |
| fresh   | < 10 days   | Green  |
| stale   | < 90 days   | Yellow |
| decayed | < 180 days  | Orange |
| dead    | >= 180 days | Red    |

Gradient bars use per-character Lab-color-space interpolation (`go-colorful`, `lipgloss`). It's overkill for a terminal tool and I'm fine with that.

## Internals

No daemon. No database. No network. Scanner calls `git log -1 --format=%ct`, an eight-week `git log --since-as-filter` for cached activity counts, and `git for-each-ref --sort=-creatordate --format="%(refname:short) %(creatordate:unix)" --count=1 refs/tags/v*`, writes JSON to `$XDG_DATA_HOME/gitfetch/cache.json`. Everything else reads cache. Always instant.

## License

MIT
