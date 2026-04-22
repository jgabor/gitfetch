# Experiments

## Experiment 1 · 2026-04-18 10:35

**Hypothesis**: UPX-LZMA decompression at every `exec` dominates the 58 ms startup. A binary built with `-trimpath -ldflags="-s -w"` but **without** UPX compression will be dramatically faster — expected < 15 ms.
**Method**: Diagnostic build with same flags as `mage install` minus UPX, benchmarked with hyperfine (50 runs). If confirmed, gate UPX behind `GITFETCH_COMPRESS=1` in `magefile.go` and re-measure through the locked harness.
**Change**: `magefile.go` `compress()` now no-ops unless `GITFETCH_COMPRESS=1`; comment documents the 60× startup cost of LZMA decompression.
**Metric**: 58.835 ms → 0.958 ms (⮉ 57.877 ms faster, 61.4×)
**Regression**: pass (`go test ./...` all green; `go vet ./...` clean)
**Status**: ■ kept
**Commit**: pending — magefile.go was untracked at session start (active user WIP), so the code change was applied in the working tree and left for the user to bundle with the rest of their in-progress commit. Only the optimera artifacts are committed under this experiment.
**Inspiration**: `file ~/.local/share/go/bin/gitfetch` reporting "no section header" + `mage install` comment warning about UPX `AlreadyPackedException` pointed directly at LZMA decompression overhead. Well-known UPX tradeoff: smaller on disk, slower every exec; LZMA variants can add tens-to-hundreds of ms.
**Conclusion**: The entire 58 ms budget was decompression. Once removed, the Go runtime, cobra init, lipgloss color-profile detection, TOML parse, and JSON parse of 16 cached repos together fit inside ~1 ms. Side effect: stddev collapsed from 0.58 ms (LZMA jitter) to 0.04 ms (deterministic cold exec + page cache). Binary grew 1.32 MB → 4.37 MB; acceptable for a tool whose vision explicitly prioritizes perceived instantaneousness over size.
**Next**: The startup-latency objective is met by a 6× margin. Close the objective unless the user wants a follow-on goal (e.g., drop the cobra+lipgloss import weight to keep a headroom for future features, shrink the uncompressed binary via `-buildmode=pie` off / further `-ldflags` tuning, or add a p99 ceiling).

## Experiment 2 · 2026-04-18 10:38

**Hypothesis**: UPX with its default NRV2B codec (`upx --best --overlay=strip`, no `--lzma`) will be meaningfully faster than LZMA but still well slower than an uncompressed binary — enough to quantify the size-vs-startup tradeoff curve.
**Method**: Diagnostic build identical to exp. 1, then pack with `upx --best --overlay=strip` (no `--lzma` flag). 50-run hyperfine bench.
**Change**: none landed — exploratory measurement only.
**Metric**: baseline-LZMA 58.835 ms → 14.238 ms (⮉ 44.6 ms faster than LZMA; ⮋ 13.3 ms slower than uncompressed)
**Regression**: n/a (no code change)
**Status**: □ discarded — informational; the uncompressed default from exp. 1 stays.
**Commit**: none
**Inspiration**: user question mid-session.
**Conclusion**: UPX's default NRV2B codec decompresses ~4× faster than LZMA but still imposes ~13 ms per exec. Size ladder: LZMA 1.32 MB / NRV2B 1.69 MB / uncompressed 4.37 MB. NRV2B is a reasonable middle ground if binary size ever becomes a hard constraint (e.g., embedded/CI images), but for a `.bashrc` CLI the 13 ms cost outweighs the 2.7 MB saving — uncompressed stays the right default.
**Next**: leave `GITFETCH_COMPRESS=1` as the LZMA opt-in; no need for a separate NRV2B toggle unless size pressure emerges.

## Objective Closed · 2026-04-22

**Final metric**: 1.294 ms mean
**Target**: ≤ 10 ms ✓ · stretch ≤ 5 ms ✓
**Summary**: Two experiments, one kept. UPX-LZMA removal (Experiment 1) was the sole meaningful change. NRV2B characterization (Experiment 2) informed the default but landed no code. Objective archived.

