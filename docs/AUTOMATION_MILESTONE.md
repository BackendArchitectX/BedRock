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

### Inspect Build handoff
- Reconciled against unchanged main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; branch was 12 commits ahead / 0 behind before this slice. Prior coordination HEAD was `15c0c474a4386d9d18bddcc88579941cbd3bd2f3`.
- Product slice completed at `21b8d5dc19134d311e31959c71334ecf00a3f9d3`: journal entries now persist `intended_sha256`, and `RecordMutationIntent` durably updates the intended-content hash while preserving the first pre-run original. Unprepared paths and completed journals fail closed. Repair attempts may replace only the intended hash, enabling later recovery to distinguish BedRock-owned bytes from conflicting external edits.
- Tests added: intent hash persistence, repair-attempt intent replacement without original-byte loss, rejection of unprepared paths, and rejection after COMPLETED.
- Tests actually run in this invocation: none. No executable checkout was available; GitHub commit status for the exact product HEAD had no checks when inspected. No PASS is claimed.
- Blockers: mutation intent is still a primitive only; `ChangeSet.Apply` does not yet invoke it immediately before each owned write, and restart recovery/crash-injection coverage is absent.
- Single next acceptance gap: wire `RecordMutationIntent` and the MUTATING transition into the real mutation boundary, then implement restart recovery that restores/removes only when current bytes match durable intended hashes and preserves mismatches as explicit conflicts.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
