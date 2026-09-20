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
- GitHub Actions run `35505816276` for `9c34b6a1703412a8aed7465862ca99dd146a8f7d` completed successfully.
- Linux CI actually executed and passed prerequisite-failure safety, format, vet, unit tests, `go test -race ./...`, canonical one-step start/rerun, foreign-workspace protection, explicit symlink-workspace-root refusal, checkout protection, and CLI failure/rollback smoke.
- The symlink-workspace regression uses a target carrying a valid BedRock marker and sentinel, invokes the canonical launcher through a symlink root, requires non-zero exit, verifies the sentinel remains unchanged, and requires the refusal diagnostic.
- Windows CI actually executed and passed the canonical one-step launcher acceptance, rerun, foreign-workspace refusal, and checkout-protection path.
- The launcher rejects a direct symlink `BEDROCK_DEMO_DIR` root before marker access or child cleanup. The preceding regression also rejects a symlink `.bedrock-demo-owned` marker and verifies that its external target and workspace are not mutated.
- Windows race detection is not executed because ThreadSanitizer could not initialize in the independently isolated local Windows environment; Linux CI remains the authoritative race gate.
- Current commit author/committer map to the `BackendArchitectX` GitHub account; no superseded prohibited author history is treated as active.

### One-step acceptance status

The deterministic prototype has one documented normal launcher, `go run ./scripts/demo.go`. The launcher checks required tools/version before creating or cleaning its workspace, requires an exact regular-file ownership marker for reuse, refuses filesystem-root, direct symlink-root, and checkout-containing workspace paths, rebuilds the CLI/provider, recreates only owned demo children, initializes the demo repository, waits for the real BedRock run and verification to finish, checks `result.txt`, and only then prints `READY`. README and CI use the same command. Rerun/idempotency, foreign-workspace refusal, prerequisite-failure non-mutation, invalid/symlink marker rejection, direct symlink-workspace refusal, Linux success/failure orchestration, and Windows launcher execution are covered by current CI.

### Remaining risks / next action

Before any new edit, fetch current `origin/main` because multiple writers may advance it. A direct symlink workspace root is now independently regression-tested and green. The next launcher safety case is an ordinary-looking `BEDROCK_DEMO_DIR` beneath a symlinked ancestor: `filepath.Abs` and `os.Lstat(work)` do not canonicalize or reject ancestor symlinks, so containment and cleanup remain lexical for that case. Add a non-mutation regression for a symlinked parent and then either reject symlink ancestors or compare canonical physical paths before any marker access or cleanup. Re-run Linux and Windows canonical acceptance after the fix.

After that, prioritize clean-checkout/release validation, failed dependency/tool diagnostics, dependency health, and documentation accuracy over feature growth.

Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made. Live-provider behavior and broader OS/environment combinations remain outside the deterministic demo proof.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
