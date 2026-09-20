# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

The only canonical one-step demo/start path is:

```text
go run ./scripts/demo.go
```

It validates Go 1.22+ and Git before workspace mutation, protects caller-selected workspaces with an exact ownership marker plus dangerous-root, checkout-containment, and symlink-component checks, builds the real CLI and deterministic provider, initializes an isolated Git repository, runs verified orchestration, validates the produced result, and prints `READY` only after success. It is intentionally a deterministic prototype path: BedRock currently has no backend server or web UI and no built-in live-model provider, so no application URL or live-provider readiness is claimed.

## 2026-09-20 independent verification

### Verified baseline

- Reconciled against current rewritten `origin/main`; superseded pre-rewrite hashes remain superseded.
- Current pre-ledger baseline `42da4d4bf6bd901088a9a7b0c08adbfd5cce043f` (`ci: remove redundant demo path safety workflow`) has successful GitHub Actions CI run `35511819998`.
- Linux CI actually executed and passed prerequisite-failure non-mutation, format, `go vet ./...`, `go test ./...`, `go test -race ./...`, canonical one-step start/rerun, foreign-workspace refusal, direct and ancestor symlink-workspace refusal, checkout protection, interrupted-owned-workspace recovery, and CLI failure/rollback smoke.
- Windows CI actually executed and passed the canonical one-step launcher, rerun/idempotency, foreign-workspace refusal, and checkout protection.
- The formerly separate Demo path safety workflow was removed only after its unique ancestor-symlink and partial-recovery assertions were consolidated into primary CI; commit `2c7a1e5e8a4774e31c5a271651acaf1ecb85f6a9` had also completed the old dedicated workflow successfully before removal.
- Windows race detection is not executed because ThreadSanitizer could not initialize in the independently isolated local Windows environment; Linux CI remains the authoritative race gate.
- Current reachable commits map to the `BackendArchitectX` GitHub account/noreply identity; superseded prohibited author history is not treated as active.

### One-step acceptance status

The deterministic prototype has one documented normal launcher, `go run ./scripts/demo.go`. README and CI use the same command. The launcher checks required tools/version before workspace mutation, requires an exact regular-file ownership marker for reuse, refuses dangerous roots, checkout-contained paths, direct symlink roots, and symlinked path components, rebuilds the CLI/provider, recreates only owned demo children, initializes the demo repository, waits for the real BedRock run and verification to finish, checks `result.txt`, and only then prints `READY`.

Current automated evidence covers rerun/idempotency, foreign-workspace refusal, prerequisite-failure non-mutation, invalid/symlink marker rejection, direct and ancestor symlink-workspace refusal, interrupted owned-workspace recovery, Linux success/failure orchestration, and Windows launcher execution. The deterministic default prototype therefore satisfies the current one-step-start acceptance gate. This does not imply live-model, server, web UI, or sandbox readiness.

### Remaining risks / next action

Before any new edit, fetch current `origin/main` because multiple writers may advance it. Launcher path-safety coverage is now consolidated in primary CI; do not recreate the removed duplicate workflow.

Prioritize clean-checkout/release validation and additional failure-path evidence over feature growth. In particular, independently challenge failures after workspace ownership is established but before `READY` (build/init/provider/verification interruption) and prove rerun recovery preserves unrelated state. Keep prerequisite diagnostics useful and fail before mutation whenever the prerequisite can be checked up front.

Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made. Live-provider behavior and broader OS/environment combinations remain outside the deterministic demo proof.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
