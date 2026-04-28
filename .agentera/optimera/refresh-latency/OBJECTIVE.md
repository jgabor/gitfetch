# Reduce gitfetch refresh wall-clock latency

## Objective
Reduce mean wall-clock time of `gitfetch refresh` from 23.4ms to under 12ms — a 2x reduction — measured via hyperfine on the user's 18-repo config. The metric is the mean time reported by hyperfine (`--warmup 1 --min-runs 5`).

## Why This Matters
`gitfetch refresh` is the cold path: it runs N git subprocesses per invocation. At 23ms for 18 repos it's already fast, but that's the floor when everything is cached by the OS. On a cold filesystem, or with more repos, the sequential scan bottleneck dominates. The real target is throughput scaling: make refresh time grow sub-linearly with repo count by parallelizing the scan loop.

## Measurement
Wall-clock time via hyperfine. The harness runs `hyperfine --json --warmup 1 --min-runs 5 './gitfetch refresh'`, extracts the mean, and reports it in milliseconds. Lower is better.

## Constraints
- `go test ./...` and `go vet ./...` must pass
- `ScanResult` struct content must be unchanged (same fields, same semantics)
- One failing repo must not abort the batch (error isolation)
- Cache output must be byte-identical to sequential execution for the same repo set
- No new external dependencies without proven value

## Scope
- `internal/git/scanner.go`: scan loop (ScanAll), ScanRepo, individual git command calls
- `cmd/gitfetch/main.go`: refresh command wiring (if needed for concurrency control)
- No changes to cache, display, TUI, config, or decay packages

## Baseline
23.4ms mean (hyperfine, 18 repos, 95 runs, stddev 1.7ms). Measured 2026-04-28.
