# Red-team continuation handoff

## 2026-09-20 verifier-contract hardening

### Finding

The engine previously trusted a `Verifier` implementation to return an error whenever any `VerificationResult.ExitCode` was non-zero. A buggy or future verifier could return a non-zero result with `nil` error, causing the engine to label a failed verification `VERIFIED`. This was a misleading-success boundary in the core evidence model.

### Work completed

- Added an engine-level invariant: every returned verification result must have exit code 0 before a run can become `VERIFIED`.
- A non-zero result paired with a nil verifier error is converted into a verification failure and follows the existing retry/rollback path.
- Added a regression verifier that deliberately violates the interface expectation and proves BedRock rejects the false success and restores the original file.

### Evidence

- Pre-change HEAD `b9167488f5bc56405eeffafd797a58e6f2a833ca` had completed successful CI run `35480612617`.
- Source fix commit: `d4cb1946d0217eb7aac08da419224641c7edd85a`.
- Regression test commit: `a7d581245deb9ec26ba4be837884dc4af3a55247`.
- CI run `35480850465` for the source-fix commit was still in progress when inspected. A completed run containing the regression-test commit was not yet visible. Therefore the new work is **UNVERIFIED** in this handoff; no local Go execution is claimed.

### Remaining risks / next action

First inspect CI for current HEAD and repair any format, vet, unit, race, or CLI-smoke failure. If green, the next high-value adversarial target is repository mutation concurrency: dirty paths are captured before provider execution, so external edits that occur after that snapshot may race with provider writes. Also keep the documented trust limitation explicit: verification shell commands and provider subprocesses still execute with local-user permissions and are not an OS sandbox.
