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
- Exact pre-ledger HEAD `7b939d5a30a2e1e86763d553f823b30fe66b7dd2` (`fix(provider): distinguish caller cancellation`) completed GitHub Actions CI run `35522241446` successfully.
- The command-provider cancellation regression now distinguishes caller cancellation from provider timeout while preserving `errors.Is(err, context.Canceled)` semantics. It is exercised by the normal cross-platform unit suite; no timeout behavior was weakened.
- Linux CI actually executed prerequisite-failure non-mutation, format, `go vet ./...`, `go test ./...`, `go test -race ./...`, canonical one-step start/rerun and consolidated launcher safety/recovery tests, and CLI failure/rollback smoke.
- Windows CI actually executed and passed the canonical one-step launcher acceptance.
- Verification-stage recovery is covered after provider output exists: an injected verification failure must not report `READY`, generated `result.txt` is absent after rollback, unrelated caller state survives, and a subsequent canonical rerun reaches `status: VERIFIED` and `READY` with fresh expected output.
- Provider-execution failure and repository-initialization failure recovery are also covered by preceding green descendants.
- Windows race detection is not executed because ThreadSanitizer could not initialize in the independently isolated local Windows environment; Linux CI remains the authoritative race gate.
- Current reachable commits map to the `BackendArchitectX` GitHub account/noreply identity; superseded prohibited author history is not treated as active.

### One-step acceptance status

The deterministic prototype has one documented normal launcher, `go run ./scripts/demo.go`. README and CI use the same command. The launcher checks required tools/version before workspace mutation, requires an exact regular-file ownership marker for reuse, refuses dangerous roots, checkout-contained paths, direct symlink roots, and symlinked path components, rebuilds the CLI/provider, recreates only owned demo children, initializes the demo repository, waits for the real BedRock run and verification to finish, checks `result.txt`, and only then prints `READY`.

Current automated evidence covers rerun/idempotency, foreign-workspace refusal, prerequisite-failure non-mutation, invalid/symlink marker rejection, direct and ancestor symlink-workspace refusal, interrupted owned-workspace recovery, recovery after injected repository-initialization failure, provider-execution failure, and verification failure, Linux success/failure orchestration, rollback, and Windows launcher execution. The deterministic default prototype therefore satisfies the current one-step-start acceptance gate. This does not imply live-model, server, web UI, or sandbox readiness.

### Remaining risks / next action

Before any new edit, fetch current `origin/main` because multiple writers may advance it. Launcher path-safety coverage is consolidated in primary CI; do not recreate the removed duplicate workflow or add scheduler-derived runtime behavior.

The highest-value remaining functional gap is the lack of a built-in self-hostable live-model provider. Independently evaluate the smallest provider-neutral HTTP adapter for OpenAI-compatible local endpoints without coupling orchestration to one vendor. Keep the deterministic zero-secret launcher as the default acceptance path. Any live-provider path must explicitly validate endpoint and model configuration, avoid fabricating secrets, bound request duration and response size, preserve cancellation causes, redact explicitly supplied credentials from diagnostics, and must not claim readiness before the configured provider is actually reachable.

The command-provider caller-cancellation diagnostic is now fixed and regression-covered. Before adding an HTTP provider, keep provider error semantics explicit: caller cancellation, adapter timeout, transport/process failure, invalid response, and verification failure should remain distinguishable without leaking configured secrets.

Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made. Live-provider behavior and broader OS/environment combinations remain outside the deterministic demo proof.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
