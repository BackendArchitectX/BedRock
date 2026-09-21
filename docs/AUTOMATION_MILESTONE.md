# Automation Milestone Gate

> Development-pipeline coordination only. This file, worker roles, scheduler cadence, and blocker-recovery rules are not BedRock runtime or product architecture.

## CURRENT: M2 — Durable run journal and crash recovery

**M1 verdict:** ACCEPTED

### Accepted baseline
- M1 accepted product/evidence HEAD: `ea1f587619604ea23aa8f339fde4df71581d3875`.
- M1 gate commit on main: `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`.
- M1 was accepted after independent exact-head executable CI, semantic inspection, and diff review.

## M2 acceptance contract — durable run journal and crash recovery
M2 requires write-ahead durable ownership/original-state evidence before mutation; durable transitions that distinguish interrupted from completed runs; process-death detection on the next invocation; safe rollback of owned bytes or explicit conflict preservation; recovery evidence that itself survives process death; preservation of M1 truthfulness and safe-change protections; and focused plus applicable regression/CI verification.

Prefer a minimal journal/state machine and bounded original-byte persistence. External automation mechanics must not enter BedRock runtime semantics.

### Challenge Build handoff
- Reconciled against unchanged main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; inspected branch product HEAD `21b8d5dc19134d311e31959c71334ecf00a3f9d3` and prior coordination HEAD `c2502af191ba041671ed6fd44b7c080ba2f3f85d`.
- Challenged assumption: `RecordMutationIntent` is a sound ownership primitive, but wiring it only immediately before `os.WriteFile` is not sufficient. The real mutation boundary starts before `os.MkdirAll(filepath.Dir(target), ...)`: for a previously absent nested path, BedRock can create parent directories and then die before the file write. Those filesystem mutations are currently neither journaled nor recoverable. The existing in-memory rollback also never removes created parent directories, so blindly copying that behavior into durable recovery would leave crash debris.
- The intended-content hash remains necessary: `ChangeSet.Rollback` currently preserves conflicts by comparing current bytes with the exact bytes BedRock wrote; after process death only the durable intended hash can reproduce that ownership check. First originals must remain immutable across repair attempts.
- No product code was changed in this challenge because the correct ownership scope must be decided at the mutation boundary rather than bolting journal calls around only `os.WriteFile`. No tests were executed in this invocation; no PASS is claimed.
- Remaining acceptance gap: integrate the journal with `ChangeSet.Apply`/engine so every filesystem mutation that M2 promises to recover is either durably owned or deliberately excluded with a tested invariant; then add restart recovery that restores/removes only exact BedRock-owned state and preserves mismatches as explicit conflicts.
- Single next action: choose and implement the smallest safe parent-directory policy (prefer pre-existing-parent-only writes for M2 if compatible; otherwise journal created directories), then wire PREPARED -> MUTATING plus per-write intent into `ChangeSet.Apply` and add a process-death test at the directory-created/file-not-written boundary.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
