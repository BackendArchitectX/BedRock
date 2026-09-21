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
- Reconciled against unchanged main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; inspected branch HEAD was `d7925e14aa81bebbfb5c55c5f9b16cdc9c972d8e`.
- Challenged assumption: preserving originals across repair attempts is necessary but not sufficient for crash-safe rollback. The current durable journal stores only pre-run bytes while `ChangeSet` keeps BedRock's last written bytes in memory. After process death, recovery therefore cannot distinguish "current bytes are exactly BedRock's write and may be rolled back" from "a user/process changed the path after BedRock wrote it and must be preserved." Blind rollback from originals would violate M2 conflict preservation.
- Required design correction: before each actual write, durably preserve the path's first pre-run original exactly once and the hash of the bytes BedRock intends to write. Later attempts may update only the intended hash, never the original. Recovery may restore/remove only when current bytes match that durable intended hash; mismatch is an explicit conflict and must be preserved. This should remain a small journal/mutation-boundary protocol, not a general agent framework.
- Product changes in this invocation: none. An attempted repository write implementing the intent-hash primitive was blocked by the execution safety layer, so no partial remote edit was left behind.
- Tests actually run in this invocation: none. No executable checkout was available and no PASS is claimed.
- Remaining acceptance gap: journal intent/original ownership is not wired atomically enough around `ChangeSet.Apply`; interrupted-run discovery/recovery and process-death tests are still absent.
- Single next action: add the minimal write-ahead mutation-intent record (first original + latest intended-content hash), wire it immediately before each owned write while keeping the journal MUTATING across repair attempts, then implement restart recovery that rolls back only exact intended bytes and preserves mismatches as conflicts.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
