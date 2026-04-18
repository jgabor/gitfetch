# gitfetch startup latency

## Objective

Reduce the mean wall-clock runtime of the default `gitfetch` command (cache-backed dashboard print, no refresh, no TUI) measured by hyperfine over ≥50 runs.

- **Metric**: mean execution time (ms), lower is better
- **Baseline**: 58.7 ms ± 0.3 ms (installed binary, UPX-LZMA packed, 16 repos in cache)
- **Target**: ≤ 10 ms mean
- **Stretch**: ≤ 5 ms mean

pfetch at 1.1 ms is the aspirational reference but is a POSIX shell script with no runtime; a statically-linked Go binary with a cobra CLI and lipgloss renderer cannot realistically reach that floor. 10 ms is fast enough for any `.bashrc` use and brings gitfetch into the same perceptual class.

## Why This Matters

VISION.md guarantees "sub-500ms for `.bashrc` use"; 58 ms already meets that, but every millisecond shows up in every new shell. Fetch-style tools live or die by perceived instantaneousness: anything that visibly delays prompt appearance erodes the "neofetch for git repos" promise. The gap to pfetch is large enough to suggest removable overhead — most likely binary decompression at exec, heavy cobra/lipgloss init-time work, and JSON/TOML parsing of small files.

## Measurement

The harness:

1. Builds the binary exactly as `mage install` does (including UPX compression when UPX is available), placing it at a known path.
2. Runs `hyperfine` with `--warmup 3 --runs 50` against the fresh binary, using the real user config and cache.
3. Parses hyperfine JSON export and emits `{metric: <mean_ms>, direction: "lower", unit: "ms", detail: "...", breakdown: [...]}`.

An experiment can opt out of UPX compression (via `GITFETCH_HARNESS_NO_UPX=1`) when testing whether packing dominates the cost. The harness always records which variant ran in `detail`.

## Constraints

- All existing tests and `go vet ./...` must pass (regression gate).
- No changes to output format — the default dashboard's rendered bytes must be byte-identical to the baseline for the same cache + config + terminal width, except when the hypothesis explicitly targets rendering work and the user sees the diff in review.
- Public CLI surface (flags, subcommands, exit codes) stays unchanged.
- XDG paths, file formats (config.toml, cache.json), and cache schema stay unchanged.
- No caching layers that defer scan work onto the default path — the default path must remain pure-read from the on-disk cache.
- No network access, no git subprocesses, no daemon.
- Follow existing code patterns; keep changes proportional to the hypothesis.

## Scope

- `cmd/gitfetch/main.go` — CLI wiring, arg parsing.
- `internal/config/` — config loading (TOML parse).
- `internal/cache/` — cache loading (JSON parse).
- `internal/display/` — dashboard formatting.
- `internal/decay/` — tier classification.
- `magefile.go` — build flags, compression toggle.
- `go.mod` — dependency swaps when justified by a hypothesis.

Out of scope unless a hypothesis explicitly justifies it: `internal/tui/`, `internal/git/`.
