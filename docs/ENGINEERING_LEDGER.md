# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

## 2026-09-20 rollback-conflict evidence verification

### Work completed

- Re-inspected current `main`, recent commits, CI, runtime/change ownership code, tests, and this ledger rather than inheriting previous claims.
- Confirmed CI run `35492792635` on predecessor HEAD `b3568dc692069bc002e39ba8f44dd84676dbabc8` completed successfully across Format, Vet, Test, Race test, CLI smoke, and CLI failure/rollback smoke.
- Found an evidence-integrity defect: when rollback refused to overwrite a concurrent user edit, the returned run error included `rollback failed`, but `Evidence.LastFailure` retained only the preceding verification failure. Durable evidence could therefore omit the most important safety failure even while `RolledBack` was correctly false.
- Fixed the rollback helper so a rollback conflict is also recorded in `LastFailure`.
- Strengthened the existing concurrent-edit regression to require durable evidence to contain both the rollback failure and the ownership-conflict detail while preserving the user's concurrent edit.

### Tests actually executed / evidence

- Predecessor GitHub Actions run `35492792635`: Format PASS, Vet PASS, Test PASS, Race test PASS, CLI smoke PASS, CLI failure rollback smoke PASS.
- New source commit: `70a9970e46cd9bb5faac177ac734df122b1cfb00`.
- New regression-test commit: `5e715f78cdbff2c6dbb593caf82c9c9daf9c5009`.
- CI run `35493244735` for the source+test HEAD was queued when inspected. Therefore the new change is **UNVERIFIED** until that run completes successfully; no pass is claimed for it yet.
- No local Go execution is claimed because this run used the connected GitHub repository API rather than a mounted checkout.

### Remaining risks / next action

First inspect CI run `35493244735` (or newer current-HEAD CI) and repair any real format/vet/test/race/CLI-smoke failure before adding functionality. If green, independently challenge evidence persistence when `SaveEvidence` itself fails and verify that safety-critical rollback/conflict information remains available to the caller. Provider subprocesses and explicit verification commands still execute with local user permissions; no sandbox claim should be made.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
