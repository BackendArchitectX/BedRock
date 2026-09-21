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
- Reconciled branch against unchanged main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; branch was 6 commits ahead and 0 behind before this slice.
- Product commit `79cf7230ce19a3ebd9f10deb4786ac7f09cd22d2` adds durable forward-only `PREPARED -> MUTATING -> COMPLETED` journal transitions. Each transition reloads persisted state, validates journal version/run identity, and atomically rewrites through a mode-0600 synced temporary file plus rename.
- Tests actually run in this invocation: none. This environment had repository read/write access through the GitHub connector but no executable checkout, so this commit is NOT claimed formatted, compiled, tested, or passing CI.
- Risk: the transition implementation has not yet been gofmt/compile checked and is not wired into the actual `ChangeSet.Apply` mutation boundary.
- Single next acceptance gap: wire journal preparation/transition ownership into the real mutation path, then add executable crash-injection/restart coverage proving interrupted-run detection and conflict-preserving recovery before Integrate considers M2.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
