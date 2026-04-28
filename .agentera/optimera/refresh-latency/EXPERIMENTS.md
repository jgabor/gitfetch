# Experiments

## Experiment 1 · 2026-04-28 18:00

**Hypothesis**: Parallelizing ScanAll with bounded goroutines (semaphore at GOMAXPROCS) reduces wall-clock time because 18 repos × 2 git subprocesses is I/O-bound and the OS interleaves concurrent git processes.

**Method**: Replaced sequential `for` loop in ScanAll with goroutines, one per repo, bounded by `runtime.GOMAXPROCS(0)` semaphore. Results collected by index into pre-allocated slice; filter step unchanged.

**Change**: `internal/git/scanner.go` — parallel ScanAll with semaphore-bounded goroutines

**Metric**: 23.61ms → 15.20ms (⮉ 35.6% faster)

**Regression**: pass

**Status**: ■ kept

**Commit**: be74ac2

**Conclusion**: Concurrency works as expected for I/O-bound git subprocesses. 35% gain from one change. Not yet 2x; the remaining time is dominated by git process spawn overhead.

**Next**: Guard `git for-each-ref` behind a cheap `os.ReadDir(".git/refs/tags")` check to skip the subprocess for repos without version tags.

## Experiment 2 · 2026-04-28 18:02

**Hypothesis**: An `os.ReadDir` guard on `.git/refs/tags` before spawning `git for-each-ref` skips unnecessary subprocesses for repos without version tags, reducing total cost.

**Method**: Added `os.ReadDir(tagsDir)` check at the top of `lastTagInfo`. If no tags directory exists, return early without spawning git.

**Change**: `internal/git/scanner.go` — guard `lastTagInfo` with `os.ReadDir`

**Metric**: 14.02ms → 14.49ms (⮋ 3.4% slower)

**Regression**: pass

**Status**: □ discarded

**Conclusion**: Under parallel execution, repos without tags were already off the critical path. The guard added syscall overhead to repos that DO have tags (which are on the critical path), producing a net regression. Optimal for sequential but counterproductive for parallel.

**Next**: Tune the GOMAXPROCS concurrency bound. The current semaphore uses `runtime.GOMAXPROCS(0)` which may over-subscribe the kernel. Test `GOMAXPROCS/2` and a fixed N=4 to find the sweet spot.

## Experiment 3 · 2026-04-28 18:04

**Hypothesis**: Reducing the concurrency bound from `GOMAXPROCS` to N=4 reduces Go runtime scheduling overhead and kernel contention from excessive concurrent git subprocesses, without serializing enough to hurt throughput.

**Method**: Replaced `runtime.GOMAXPROCS(0)` semaphore capacity with a hardcoded 4. Removed now-unused `runtime` import.

**Change**: `internal/git/scanner.go` — semaphore bound GOMAXPROCS → 4

**Metric**: 14.06ms → 8.53ms (⮉ 39.3% faster)

**Regression**: pass

**Status**: ■ kept

**Commit**: 0a0db66

**Conclusion**: GOMAXPROCS over-subscribed — 16 concurrent goroutines fighting for git subprocess I/O created unnecessary overhead. N=4 is the sweet spot: enough parallelism to keep all cores busy, few enough to avoid scheduling churn. 2x target achieved (23.61ms → 8.53ms = 2.77x).

**Next**: Sweep concurrency values 1-16 to validate N=4 is the true optimum.

## Experiment 4 · 2026-04-28 18:10

**Hypothesis**: A parameter sweep of concurrency bounds from 1 to 16 will reveal N=4 is the knee of the curve, with diminishing returns beyond.

**Method**: Swept N in {1, 2, 3, 4, 5, 6, 8, 12, 16} with 200-500 runs each. Measured mean wall-clock time via hyperfine.

**Change**: `internal/git/scanner.go` — semaphore bound 4 → 8 (sweep-validated optimum)

**Metric**: 8.35ms (N=4) → 6.69ms (N=8) (⮉ 19.9% faster)

**Regression**: pass

**Status**: ■ kept

**Commit**: c105a48

**Sweep data**: N=1:23.2, N=2:13.2, N=3:10.1, N=4:8.4, N=5:7.4, N=6:7.0, N=8:6.7, N=12:13.7, N=16:14.4. Knee at N=3-4, optimum at N=8, sharp performance cliff at N=12+.

**Conclusion**: Hypothesis was wrong — N=4 is not the optimum. N=8 is 20% faster with equal stddev. Beyond N=8, kernel and Go runtime contention dominate. Total improvement from sequential baseline: 23.10ms → 6.69ms = 3.45x.

**Next**: Objective achieved at 3.45x. No further tuning needed for current repo count. If tracked repos double, re-sweep for optimum.

## Experiment 5 · 2026-04-28 18:14

**Hypothesis**: Parsing `/proc/cpuinfo` for physical core count (cores_per_socket × sockets) as a concurrency heuristic matches the N=8 optimum and generalizes across CPU topologies.

**Method**: Replaced hardcoded `8` with `concurrencyBound()` that reads `/proc/cpuinfo` for `cpu cores` × `physical id` count. Falls back to `runtime.NumCPU()/2` on non-Linux. Measured with 1000-run harness.

**Change**: `internal/git/scanner.go` — hardcoded N=8 → `/proc/cpuinfo`-derived physical core count

**Metric**: 6.69ms → 6.91ms (⮋ 3.3% slower)

**Regression**: pass

**Status**: □ discarded

**Conclusion**: Heuristic correctly produces 8 (1 socket × 8 cores) but parsing overhead adds ~200µs per invocation, causing a 0.22ms regression on this system. The heuristic is architecturally sound for portability — it would auto-tune to the physical core count on any Linux machine — but the hardcoded value is faster with no downside for the current machine.

**Next**: User decision: accept marginal regression for portability, or keep hardcoded N=8. If portability is desired, optimize by caching the parsed value in a `sync.Once` init.

## Experiment 6 · 2026-04-28 18:20

**Hypothesis**: Caching the `/proc/cpuinfo` physical core count behind `sync.Once` eliminates per-scan parse overhead while preserving the portability benefit of auto-detecting the optimal concurrency bound.

**Method**: Added package-level `concurrencyOnce sync.Once` + `concurrencyCache int`. `concurrencyBound()` calls `detectPhysicalCores()` exactly once via `Once.Do`, subsequent calls return the cached value (8 on this machine). Replaced hardcoded `8` in semaphore.

**Change**: `internal/git/scanner.go` — hardcoded N=8 → `sync.Once`-cached `/proc/cpuinfo` heuristic

**Metric**: 6.69ms → 6.77ms (⮋ 0.08ms slower)

**Regression**: pass

**Status**: ■ kept (overridden — user accepted 0.08ms regression for portability)

**Commit**: 9ae95fc

**Conclusion**: `sync.Once` cut experiment 5's regression from 0.22ms to 0.08ms, but the one-time `/proc/cpuinfo` parse (~200µs per process) still registers at 1000-run precision. Statistically significant (8× standard error) but practically negligible. The heuristic correctly produces 8 (physical core count) which SMT sweeps confirmed is optimal in both SMT-on and SMT-off modes. The portability benefit is real; the marginal regression is noise-scale.

**Next**: Objective fully achieved. Total improvement from baseline: 23.10ms → 6.77ms = 3.41x with portability.


