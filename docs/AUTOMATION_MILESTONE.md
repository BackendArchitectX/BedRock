# Automation Milestone Gate

> Development-pipeline coordination only. This file, worker roles, scheduler cadence, and blocker-recovery rules are not BedRock runtime or product architecture.

## CURRENT: M2 — Durable run journal and crash recovery

**M1 verdict:** ACCEPTED

### Accepted baseline

- M1 accepted product/evidence HEAD: `ea1f587619604ea23aa8f339fde4df71581d3875`.
- M1 gate commit on main: `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`.
- M1 was accepted only after independent exact-head executable CI, semantic inspection, and diff review.

### M1 acceptance evidence

GitHub Actions CI run `35582125964` completed successfully on exact milestone HEAD `ea1f587619604ea23aa8f339fde4df71581d3875`.

Linux `test` passed launcher prerequisite failure safety, gofmt, `go vet ./...`, `go test ./...`, `go test -race ./...`, canonical one-step start, and CLI failure rollback smoke. Windows `windows-launcher` passed the canonical one-step start.

Independent integration review confirmed that M1 materially changes persisted outcome truthfulness rather than merely renaming a status: baseline/post verification, `CHECKS_PASSED`, provider summaries, attempts, terminal failure evidence, changed paths, and deterministic content-sensitive diff hash are persisted without claiming task completion.

## M2 acceptance contract — durable run journal and crash recovery

M2 must not be accepted based on ordinary error-return rollback. Require all of the following together:

- write-ahead durable ownership/original-state evidence exists before repository mutation;
- durable state transitions allow the next invocation to distinguish an interrupted run from a completed run;
- a simulated or real process death after mutation is detected on the next invocation;
- the next invocation can safely recover/rollback owned bytes, or reports an explicit conflict without overwriting externally changed bytes;
- recovery state itself survives process death and is covered by executable crash-injection evidence;
- existing M1 truthful outcome/evidence semantics and safe-change protections remain intact;
- focused tests and applicable broader regression/CI gates pass.

Prefer a minimal journal/state machine and content-addressed or otherwise bounded original-byte persistence over speculative workflow machinery. Do not import external worker/scheduler mechanics into BedRock.

### Inspect Build handoff

- Reconciled `automation/bedrock-current` with main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; they were identical before this M2 slice.
- Product slice commits: `88dd0258d0a0f8d67debef17b03721c19f71d112` adds a repository-external, atomic write-ahead run journal that records bounded path identity, existence, mode, SHA-256 and exact original bytes in `PREPARED` state; `8053db8e4d9ce58b8bf9d87b5057ccd6dfa998ed` adds executable tests for persistence and unsafe/duplicate path rejection.
- Verification actually executed in this invocation: none. GitHub had not surfaced an Actions run for `8053db8e4d9ce58b8bf9d87b5057ccd6dfa998ed` at inspection time, so this slice is not claimed PASS.
- This is intentionally only the first M2 primitive. It is not yet wired into `ChangeSet.Apply`, has no `MUTATING`/`COMPLETED` transition API, and cannot yet recover an interrupted run.
- Single next acceptance gap: wire journal preparation and durable state transitions into the mutation boundary so a process-death test can prove next-invocation detection and conflict-preserving rollback.

### Blocker-recovery protocol for all five scheduled workers

A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
