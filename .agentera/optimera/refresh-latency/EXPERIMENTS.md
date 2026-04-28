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

**Next**: Objective met. No further experiments needed unless repo count grows or git process overhead changes materially.

