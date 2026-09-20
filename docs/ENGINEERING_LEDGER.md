# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes either a provider-neutral command adapter or the built-in OpenAI-compatible HTTP adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

The only canonical one-step demo/start path is:

```text
go run ./scripts/demo.go
```

It validates Go 1.22+ and Git before workspace mutation, protects caller-selected workspaces with an exact ownership marker plus dangerous-root, checkout-containment, and symlink-component checks, builds the real CLI and deterministic provider, initializes an isolated Git repository, runs verified orchestration, validates the produced result, and prints `READY` only after success. It intentionally remains a deterministic zero-secret acceptance path; BedRock currently has no backend server or web UI, so no application URL is claimed.

## 2026-09-20 current verification

### Verified baseline

- Reconciled against current rewritten `origin/main`; superseded pre-rewrite hashes remain superseded.
- Exact pre-change HEAD `239320e725afd7418866687d5205584fdde54a36` (`docs: align live provider guidance`) completed GitHub Actions CI run `35525905900` successfully.
- Linux CI executes prerequisite-failure non-mutation, format, `go vet ./...`, `go test ./...`, `go test -race ./...`, canonical one-step start/rerun, consolidated launcher safety/recovery tests, and CLI failure/rollback smoke.
- Windows CI executes the canonical one-step launcher acceptance. Local Windows race detection remains not executed because ThreadSanitizer cannot initialize in the independently isolated environment; Linux CI is the authoritative race gate.
- Launcher recovery is covered after repository initialization, provider execution, and verification failures while preserving unrelated caller state.
- The built-in OpenAI-compatible HTTP provider now has focused coverage for successful execution, bearer authentication, credential redaction, caller cancellation, unsafe configuration, response-size limits, and rejection of trailing HTTP envelope data.
- README documents both the external command adapter and built-in HTTP provider while preserving the deterministic zero-secret canonical demo.

### One-step acceptance status

The deterministic prototype has one documented normal launcher, `go run ./scripts/demo.go`. README and CI use the same command. The launcher checks required tools/version before workspace mutation, requires an exact regular-file ownership marker for reuse, refuses dangerous roots, checkout-contained paths, direct symlink roots, and symlinked path components, rebuilds the CLI/provider, recreates only owned demo children, initializes the demo repository, waits for the real BedRock run and verification to finish, checks `result.txt`, and only then prints `READY`.

Current automated evidence covers rerun/idempotency, foreign-workspace refusal, prerequisite-failure non-mutation, invalid/symlink marker rejection, direct and ancestor symlink-workspace refusal, interrupted owned-workspace recovery, recovery after injected repository-initialization failure, provider-execution failure, and verification failure, Linux success/failure orchestration, rollback, and Windows launcher execution. The deterministic default prototype therefore satisfies the current one-step-start acceptance gate. This does not imply live-model, server, web UI, or sandbox readiness.

### Current in-flight change

After the green `239320e` baseline, `e1c499c` makes provider configuration fail closed instead of silently ignoring command-provider-only `--provider-arg` / `--provider-env` options when `--provider-endpoint` is selected. `b9b5bef` adds regression coverage for both mixed-option cases. CI for the descendant must complete before these changes are called verified.

### Remaining risks / next action

Before any new edit, fetch current `origin/main` because multiple writers may advance it. Launcher path-safety coverage is consolidated in primary CI; do not recreate the removed duplicate workflow or add scheduler-derived runtime behavior.

First inspect CI for `b9b5bef` or its latest descendant and repair any genuine regression. If green, the next provider work should focus on live-endpoint operability evidence rather than expanding the wire protocol speculatively: exercise the built-in HTTP provider through the actual CLI against a controlled OpenAI-compatible HTTP server, including authenticated configuration and an unreachable/malformed endpoint, and prove BedRock does not report successful orchestration when provider readiness fails.

Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made. The built-in HTTP adapter implements a deliberately narrow OpenAI-compatible chat-completions contract; compatibility with a specific self-hosted model/runtime is not implied until exercised against it. No live external model is part of the deterministic one-step acceptance path.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
