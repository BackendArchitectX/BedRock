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

### Challenge Build handoff

- Reconciled against main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; main remained unchanged at the final fetch. Branch product HEAD before this handoff commit: `9136c8871e087ce5afba5efe587b58407d6c721e`.
- Challenged assumption: the new repository-external journal was safe merely because changed repository paths were validated. It was not: `RunID` was concatenated into the journal filename and temp-file pattern without validation, allowing path separators/traversal to escape the intended per-repository journal directory.
- Changes made: `cf25b970024072b225133500f1bf794d1a691973` validates run identifiers before journal construction and again at persistence; `9136c8871e087ce5afba5efe587b58407d6c721e` adds regression coverage for empty, dot, traversal, backslash and nested run IDs. No M2 runtime architecture was added.
- Tests actually run in this invocation: none. GitHub had not surfaced an Actions run for `9136c8871e087ce5afba5efe587b58407d6c721e` at inspection time, so this slice is not claimed PASS.
- Remaining acceptance gap: the journal is still not wired into `ChangeSet.Apply`, has no durable `MUTATING`/`COMPLETED` transition API, and cannot detect/recover an interrupted mutation on the next invocation.
- Single next action: wire journal preparation plus durable state transitions into the mutation boundary and add a process-death/restart test that proves conflict-preserving recovery.

### Blocker-recovery protocol for all five scheduled workers

A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
