# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

The only canonical one-step demo/start path is:

```text
go run ./scripts/demo.go
```

It validates Go 1.22+ and Git, protects caller-selected workspaces with an ownership marker and dangerous-root/checkout containment checks, builds the real CLI plus deterministic provider, initializes an isolated Git repository, runs verified orchestration, validates the produced result, and prints `READY` only after success. It is intentionally a deterministic prototype path: BedRock currently has no backend server or web UI and no built-in live-model provider, so no application URL or live-provider readiness is claimed.

## 2026-09-20 release integration

### Verified baseline

- Reconciled against current rewritten `origin/main`; superseded pre-rewrite hashes remain superseded.
- GitHub Actions run `35501930929` for `29a79cb7a57e8ac3dd42f0cc7222824960a92043` completed successfully.
- Linux CI passed format, vet, unit tests, `go test -race ./...`, canonical one-step start, rerun/idempotency and failure/rollback smoke.
- Windows CI passed the exact canonical one-step launcher, result/readiness assertions, rerun behavior and foreign-workspace protection.
- Windows race detection is not executed because the user independently reproduced a ThreadSanitizer startup/address-space failure outside BedRock; Linux CI remains the authoritative race gate.

### Integration delta

- Removed obsolete `scripts/demo.sh`. It duplicated the canonical Go launcher, was no longer referenced by README/CI, was POSIX-only, and had weaker workspace-marker validation than `scripts/demo.go`. Keeping it created an unnecessary second launcher implementation and a future security/documentation drift risk.
- The canonical Go launcher remains the sole normal startup implementation and command.

### Verification status

- Pre-change baseline `29a79cb7`: GitHub Actions run `35501930929` PASS on both Linux and Windows jobs.
- Removal commit `54aeb525d8babe567cec0e8642a158d98eef7221`: CI pending at handoff; do not infer PASS from the predecessor.
- This ledger-refresh commit also requires current-HEAD CI before release claims are advanced.

### Remaining risks / next action

First inspect CI for the current HEAD and repair any real regression before further work. If green, the deterministic prototype satisfies the documented one-step launcher gate on Linux and Windows CI with one launcher implementation. Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made. Live-provider behavior and broader OS/environment combinations remain outside the deterministic demo proof. Prefer security review, dependency health, clean-checkout validation and factual release documentation over new feature growth.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
