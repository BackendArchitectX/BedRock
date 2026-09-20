# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

The only canonical one-step demo/start path is:

```text
go run ./scripts/demo.go
```

It validates Go 1.22+ and Git before workspace mutation, protects caller-selected workspaces with an exact ownership marker plus dangerous-root/checkout containment and symlink-root checks, builds the real CLI and deterministic provider, initializes an isolated Git repository, runs verified orchestration, validates the produced result, and prints `READY` only after success. It is intentionally a deterministic prototype path: BedRock currently has no backend server or web UI and no built-in live-model provider, so no application URL or live-provider readiness is claimed.

## 2026-09-20 independent verification

### Verified baseline

- Reconciled against current rewritten `origin/main`; superseded pre-rewrite hashes remain superseded.
- GitHub Actions run `35504493734` for `f9fb8434697adeebf0063ba6316cea00328fc4e6` completed successfully.
- Linux CI actually executed and passed prerequisite-failure safety, format, vet, unit tests, `go test -race ./...`, canonical one-step start/rerun, foreign-workspace/checkout protection, and CLI failure/rollback smoke.
- Windows CI actually executed and passed the canonical one-step launcher acceptance, rerun, foreign-workspace refusal, and checkout-protection path.
- The launcher rejects a symlink `BEDROCK_DEMO_DIR` root before marker access or child cleanup. The preceding regression also rejects a symlink `.bedrock-demo-owned` marker and verifies that its external target and workspace are not mutated.
- Windows race detection is not executed because ThreadSanitizer could not initialize in the independently isolated local Windows environment; Linux CI remains the authoritative race gate.
- Current commit author/committer map to the `BackendArchitectX` GitHub account; no superseded prohibited author history is treated as active.

### One-step acceptance status

The deterministic prototype has one documented normal launcher, `go run ./scripts/demo.go`. The launcher checks required tools/version before creating or cleaning its workspace, requires an exact regular-file ownership marker for reuse, refuses filesystem-root, symlink-root, and checkout-containing workspace paths, rebuilds the CLI/provider, recreates only owned demo children, initializes the demo repository, waits for the real BedRock run and verification to finish, checks `result.txt`, and only then prints `READY`. README and CI use the same command. Rerun/idempotency, foreign-workspace refusal, prerequisite-failure non-mutation, invalid/symlink marker rejection, Linux success/failure orchestration, and Windows launcher execution are covered by current CI.

### Remaining risks / next action

Before any new edit, fetch current `origin/main` because multiple writers may advance it. The symlink-root implementation itself is green in CI, but the exact `BEDROCK_DEMO_DIR` symlink-root refusal does not yet have its own explicit regression in the workflow; add that focused non-mutation regression before further launcher hardening. Then prioritize clean-checkout/release validation, failed dependency/tool diagnostics, dependency health, and documentation accuracy over feature growth.

Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made. Live-provider behavior and broader OS/environment combinations remain outside the deterministic demo proof.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
