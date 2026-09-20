# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

## 2026-09-20 canonical launcher Windows verification

### Work completed

- Reconciled against current `origin/main` at `f574dba00d1e44f1f2677bd27bccc00c7f33f0b0`, not the superseded pre-rewrite hashes. GitHub Actions run `35497452398` for that baseline completed successfully.
- Challenged the cross-platform `go run ./scripts/demo.go` launcher against the one-step acceptance requirement. The implementation already branches verifier syntax by `runtime.GOOS`, but CI only exercised the canonical launcher on Ubuntu. Therefore Windows one-step support was implemented but not independently proven.
- Added a dedicated `windows-latest` CI job that executes the exact canonical command, requires the real orchestration result to equal `good`, requires emitted `status: VERIFIED` and `BedRock demo: READY`, reruns the same command to exercise idempotent owned-workspace cleanup, and exercises the foreign-workspace ownership guard while proving its sentinel remains unchanged.
- Kept the Linux race gate unchanged. The user-reproduced Windows ThreadSanitizer startup failure is environmental and this Windows launcher job intentionally does not run `go test -race`; Linux CI remains authoritative for race detection.

### Tests actually executed / evidence

- Baseline `f574dba00d1e44f1f2677bd27bccc00c7f33f0b0`: GitHub Actions run `35497452398` completed successfully, including Linux format, vet, tests, race tests, canonical launcher, rerun, foreign-workspace guard, and failure rollback smoke.
- Windows CI coverage commit: `631cda1506a1671a342dd0b1c802f8bf86ea5e82`.
- At the first post-push inspection, a workflow run for `631cda1` was not yet visible in the Actions listing. Therefore the new Windows job is **UNVERIFIED** until a run for this commit or a descendant executes successfully. Do not infer PASS from the green predecessor.

### Remaining risks / next action

First inspect current-HEAD CI. If the Windows job fails, repair the actual launcher/PowerShell portability defect rather than weakening assertions. If it passes, Windows one-step startup can be claimed for the deterministic prototype. The next one-step acceptance gap should then be prerequisite diagnostics: the launcher says Go 1.22+ is required but currently only checks that a `go` executable exists; determine whether explicit version validation adds actionable value beyond the `go run`/`go.mod` failure without duplicating Go's own compatibility logic. Provider subprocesses and verification commands still execute with local user permissions; no sandbox claim should be made.

## 2026-09-20 one-step launcher cleanup safety

The canonical launcher uses an ownership marker and refuses existing unmarked workspaces. It only recreates BedRock-owned `bin` and `repository` children. This protects caller-selected foreign directories from recursive cleanup. Detailed historical commits before the controlled author rewrite are superseded; current repository state is authoritative.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
