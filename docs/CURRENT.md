# Current Engineering Handoff

## 2026-09-20 verifier cancellation CI repair

### Current state

BedRock remains a Go 1.22 local-first orchestration prototype on `main`. Runtime/provider/verifier cancellation support is present, but the latest completed CI exposed a flaky cancellation regression test rather than a demonstrated runtime defect.

### Work completed

- Re-inspected current `main`, recent commits, current verifier source/tests, durable ledger, and actual GitHub Actions logs instead of carrying forward prior claims.
- CI run `35491836217` on `ce3a984eb1f643f544f8d931bf7bd9b39a6a2688` passed Format and Vet, then failed only `TestShellVerifierPreservesCancellationCause` after 10.11 seconds; race and CLI smoke stages were skipped.
- Root cause: the test launched `sh -c "sleep 30"`. Canceling Go's `exec.CommandContext` kills the shell process, but a spawned descendant can retain the inherited stdout/stderr pipes; `Cmd.Run` can therefore wait for pipe EOF even though the command process was canceled. The test was asserting process-tree behavior BedRock does not currently implement.
- Replaced that flaky descendant-process timing test with a deterministic already-canceled-context regression. It still verifies that `ShellVerifier` preserves `context.Canceled` without pretending BedRock has process-group sandbox/termination semantics.

### Tests actually executed / evidence

- GitHub Actions run `35491836217`: Format PASS; Vet PASS; `go test ./...` FAIL only at `TestShellVerifierPreservesCancellationCause`; race and CLI smoke stages skipped.
- Repair commit: `68635b1a590bde2a461aca09395ff4c089a25cd8` (`fix(test): remove shell descendant cancellation flake`).
- CI run `35492245417` for the repair was queued at final inspection. No local Go execution is claimed.

### Verification status

**UNVERIFIED on repair HEAD** until run `35492245417` completes. Do not report unit/race/CLI PASS for this commit yet.

### Remaining risks / next action

First inspect run `35492245417` and repair any actual failure. Separately, treat descendant-process termination as an explicit known limitation: provider/verifier commands may spawn children that outlive cancellation or hold pipes open. If this becomes a required safety guarantee, design cross-platform process-tree termination deliberately rather than encoding it accidentally in a unit test. Provider and verification subprocesses still execute with local user permissions; no sandbox-security claim should be made.
