# Latest engineering handoff

## 2026-09-20 — repair-attempt mutation guard

### Current state

BedRock remains a Go 1.22 local-first orchestration prototype on `main`. Before this pass, CI run `35487275560` on `064a9b202ec8218b0259396cef1b714e6b26eb0f` completed successfully.

### Work completed

- Re-inspected current `main`, recent commits, CI, `docs/ENGINEERING_LEDGER.md`, runtime change tracking, engine retry behavior, and existing concurrency tests.
- Challenged the new provider-execution mutation guard and found a second-attempt data-loss hole: after a failed verification, the engine removed every previously changed path from the fresh dirty-path set without checking whether the file still contained BedRock's last write. A user edit between repair attempts could therefore be overwritten by the retry.
- Added `ChangeSet.ExcludeOwnWrites`, which removes a dirty path from protection only when its current bytes still exactly match BedRock's last recorded write. Missing or externally changed files remain protected.
- Replaced the engine's unconditional deletion of previously changed paths with that ownership check.
- Added `TestEnginePreservesTargetChangedBetweenRepairAttempts`, which performs a real temporary Git workflow, causes the first provider edit to fail verification, mutates the target externally during the second provider call, and requires BedRock to reject the overwrite and preserve the external contents.

### Verification actually observed

- Baseline CI run `35487275560` on `064a9b202ec8218b0259396cef1b714e6b26eb0f`: **PASS**.
- CI run `35487745993` on the first source commit failed only at `gofmt` because `changes.go` lacked a trailing newline; vet/tests/race/smokes were skipped. The exact formatter log was inspected and the newline was fixed in `809c5642b52d119787023e074e6cb56c774caafc`.
- The new regression test file also received a terminating newline in `086bb048ec84ef01f716c9337d26fa7b2c3bb875`.
- CI for the final source + regression-test HEAD was not yet visible/completed when this handoff was written. Therefore the new behavior is **UNVERIFIED** until a run containing `086bb048...` (or a descendant containing it) completes successfully.
- No local Go execution is claimed; this pass used the connected GitHub repository/API.

### Remaining risks / next action

First inspect CI on the current HEAD and repair any actual format/vet/unit/race/CLI-smoke failure. If green, challenge the remaining time-of-check/time-of-use window between the final dirty-path check and `os.WriteFile`: the current guard detects mutations made during provider execution and between repair attempts, but another process could still race after the check and before the write. Do not add a complex locking subsystem unless a simple safe write strategy can materially reduce that window. Provider and verification subprocesses still run with local-user permissions; no sandbox claim should be made.
