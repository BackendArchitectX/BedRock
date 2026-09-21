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
- Reconciled against unchanged main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`. Exact product/test HEAD before this handoff commit: `09e595b70ee4a255afebfdb4d4e389a9e11fd3dd`; this handoff commit is the only subsequent branch change.
- Challenged assumption: requiring every path to be known at initial `PrepareRunJournal` is incompatible with the existing bounded repair loop. A later provider attempt may legitimately introduce a new path. Rejecting it would either break repair or tempt callers to recreate/replace the journal, losing the first-attempt ownership set.
- Changed `RecordMutationIntent` so a path first introduced by a later repair attempt has its current original captured and persisted atomically with its intended-content hash before mutation. Existing paths still retain their immutable first original while only the intended hash changes. Added focused regression coverage proving a late path introduced after `PREPARED -> MUTATING` persists its pre-repair bytes and intended hash. This materially closes the multi-attempt ownership gap without adding a second journal or agent abstraction.
- The previous parent-directory challenge remains valid and is now the next blocker: `ChangeSet.Apply` can mutate the filesystem via `MkdirAll` before the file write. Those created directories are not yet owned by the journal or recoverable after process death.
- Tests actually run: none in this invocation. GitHub contents writes were available but no executable checkout was available, so no gofmt/vet/test/race/smoke PASS is claimed. The added Go is written in gofmt-compatible form but still requires executable verification.
- Remaining acceptance gap: integrate the journal at the real mutation boundary; define ownership/recovery for parent directories created by BedRock; detect interrupted journals on restart; restore/remove only state whose current bytes still match durable BedRock intent and preserve mismatches as explicit conflicts; prove this with process-death tests and applicable regression/CI.
- Single next action: extend the journal with the smallest explicit created-directory ownership record, persist that ownership before/around directory creation, and add a crash-boundary test for directory-created/file-not-written before wiring full restart recovery.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
