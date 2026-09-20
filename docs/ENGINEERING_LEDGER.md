# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

The only canonical one-step demo/start path is:

```text
go run ./scripts/demo.go
```

It validates Go 1.22+ and Git before workspace mutation, protects caller-selected workspaces with an exact ownership marker plus dangerous-root/checkout containment checks, builds the real CLI and deterministic provider, initializes an isolated Git repository, runs verified orchestration, validates the produced result, and prints `READY` only after success. It is intentionally a deterministic prototype path: BedRock currently has no backend server or web UI and no built-in live-model provider, so no application URL or live-provider readiness is claimed.

## 2026-09-20 independent verification

### Verified baseline

- Reconciled against current rewritten `origin/main`; superseded pre-rewrite hashes remain superseded.
- GitHub Actions run `35503624556` for `e9b161ccb9545d015b557a09e45b19ca7931b27f` completed successfully.
- Linux CI actually executed and passed prerequisite-failure safety, format, vet, unit tests, `go test -race ./...`, canonical one-step start, and CLI failure/rollback smoke.
- Windows CI actually executed and passed the canonical one-step launcher acceptance step.
- The current independent regression rejects a symlink `.bedrock-demo-owned` marker and verifies that its external target and workspace are not mutated.
- Windows race detection is not executed because ThreadSanitizer could not initialize in the independently isolated local Windows environment; Linux CI remains the authoritative race gate.
- Current commit author/committer map to the `BackendArchitectX` GitHub account; no superseded prohibited author history is treated as active.

### One-step acceptance status

The deterministic prototype has one documented normal launcher, `go run ./scripts/demo.go`. The launcher checks required tools/version before creating or cleaning its workspace, requires an exact regular-file ownership marker for reuse, refuses filesystem-root and checkout-containing workspace paths, rebuilds the CLI/provider, recreates only owned demo children, initializes the demo repository, waits for the real BedRock run and verification to finish, checks `result.txt`, and only then prints `READY`. README and CI use the same command. Rerun/idempotency, foreign-workspace refusal, prerequisite-failure non-mutation, invalid/symlink marker rejection, Linux success/failure orchestration, and Windows launcher execution are covered by current CI.

### Remaining risks / next action

Before any new edit, fetch current `origin/main` because multiple writers may advance it. A worthwhile next independent safety check is physical-path handling for a caller-supplied `BEDROCK_DEMO_DIR` whose workspace path itself is a symlink: current containment checks are lexical, while marker access and child cleanup traverse path components. Determine the intended policy, add a regression first, and either reject a symlink workspace root or canonicalize it before containment/cleanup. Also continue checking failed dependency/tool diagnostics and clean-checkout behavior rather than adding product runtime concepts merely to mirror external scheduling.

Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made. Live-provider behavior and broader OS/environment combinations remain outside the deterministic demo proof.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
