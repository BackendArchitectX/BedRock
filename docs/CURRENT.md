# Current Engineering Handoff

## 2026-09-20 pre-apply mutation guard verification

### Current state

BedRock remains a Go 1.22 local-first orchestration prototype on `main`. The pre-apply mutation guard is implemented: after provider execution, the engine re-reads Git dirty paths before applying provider changes so edits made while the provider is running are protected.

### Work completed

- Re-inspected current history, current engine/test code, and the latest GitHub Actions result instead of carrying forward prior PASS claims.
- Latest CI run `35486767580` on `172d514c2d6ecc8a83292813b705b53bad303023` failed only in `TestEngineRejectsTargetChangedDuringProviderExecution`.
- The runtime behavior was correct: `ChangeSet.Apply` rejected the provider overwrite with `refusing to overwrite pre-existing dirty path "result.txt"` and preserved the external edit. The test expected an older/different wording (`refusing to overwrite dirty path`) and therefore failed.
- Updated only the assertion to match the stable error emitted by the existing protection path. No runtime behavior was weakened or changed.

### Tests actually executed / evidence

- GitHub Actions run `35486767580`: Format PASS; Vet PASS; `go test ./...` FAIL solely at `TestEngineRejectsTargetChangedDuringProviderExecution` because of the stale error substring. Race and CLI smoke stages were skipped after the unit-test failure.
- Commit `96a6285a4919599d2499539b5fc885d04179448f` updates the regression assertion.
- No local Go execution is claimed; this run used the connected GitHub repository and CI evidence.

### Verification status

**UNVERIFIED on current HEAD** until GitHub Actions completes on a commit containing the corrected assertion. The preceding failure demonstrates that the new mutation guard itself rejected the overwrite and preserved the external edit, but race and CLI stages still need to execute successfully on the corrected HEAD.

### Remaining risks / next action

First inspect CI on current HEAD. If green, independently challenge the pre-apply guard for non-Git directories and for provider changes spanning both BedRock-owned and externally modified paths. Keep the fail-closed rule: never overwrite user work merely to complete an autonomous run. Provider and verification subprocesses still run with local user permissions; no sandbox-security claim should be made.
