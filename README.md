# gitfetch

Tracks which of your repos you've been neglecting. Like neofetch, but for commit and release decay.

## Install

Go 1.26+, Mage, or plain `go build`.

```
git clone https://github.com/jgabor/gitfetch
cd gitfetch
mage install
```

## Usage

```
gitfetch                  # dashboard from cache (instant)
gitfetch refresh          # rescan, update cache, print
gitfetch refresh --author "jgabor" --remote "github"
gitfetch tui              # interactive repo manager
gitfetch --verbose        # show headers and color legend
```

| Command | |
|---|---|
| `gitfetch` | Reads cache, prints dashboard, exits. No git subprocesses, sub-500ms, `.bashrc`-safe. |
| `gitfetch refresh` | Runs `git log` and `git for-each-ref` on every tracked repo. Writes results to cache. |
| `gitfetch tui` | Bubbletea TUI with vim keys. Add/remove repos, trigger rescans, multi-repo discover. |

## Configuration

`$XDG_CONFIG_HOME/gitfetch/config.toml`, auto-created on first run:

```toml
repos = []
refresh_every = 0    # auto-rescan every N invocations (0 = never)

[display]
color = "auto"       # auto, always, never
width = 0            # max width (0 = auto-detect terminal)
```

Populate `repos` via `gitfetch tui` (`a` to add) or by editing the file. Paths support `~/`.

## Decay dashboard

One bar per repo showing commit freshness. Tag freshness colors the Version column text — green, yellow, orange, or red depending on how long since the last release.

| Tier | Threshold | Bar |
|------|-----------|-----|
| fresh | < 10 days | Green |
| stale | < 90 days | Yellow |
| decayed | < 180 days | Orange |
| dead | >= 180 days | Red |

Gradient bars use per-character Lab-color-space interpolation (`go-colorful`, `lipgloss`). It's overkill for a terminal tool and I'm fine with that.

## Internals

No daemon. No database. No network. Scanner calls `git log --format=%ci -1` and `git for-each-ref --sort=-creatordate`, writes JSON to `$XDG_DATA_HOME/gitfetch/cache.json`. Everything else reads cache. Always instant.

## License

MIT
