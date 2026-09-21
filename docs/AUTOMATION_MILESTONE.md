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
- Fresh reconciliation: `main` remains `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`. The shared branch was `9a6582658acd463da724179cd4a4568c32416802` before this slice, so no newer main work required reconciliation. CURRENT remains M2.
- Product slice completed: bounded the real `ChangeSet.Apply` mutation surface by removing its `MkdirAll` behavior. A proposed file now requires its immediate parent directory to already exist and be a real directory; a missing parent is rejected before originals are captured or filesystem mutation begins. This avoids falsely claiming durable ownership of directories that `MkdirAll` cannot race-safely attribute to BedRock.
- Product/test HEAD before this handoff: `8e07c30d35a5f66cb20d45ba680ab259c826df2b` (`fab535dc` product invariant, `8e07c30d` focused regression coverage). Tests prove a missing nested parent is rejected and remains absent, while a new file in an existing parent remains supported.
- Verification actually run: no executable Go checks in this invocation. GitHub combined status for exact product/test HEAD returned no statuses at inspection time, so no gofmt/vet/test/race/smoke PASS is claimed.
- Remaining risk: the durable journal is still not wired into `ChangeSet.Apply`; interrupted-run discovery/recovery is not implemented. Rollback's legacy `MkdirAll` remains for restoring previously existing files, but the new forward-mutation invariant means Apply itself no longer creates unowned directories.
- Single next acceptance gap: wire durable original capture/intended hashes and `PREPARED -> MUTATING` into the now-bounded file write boundary, then detect an interrupted journal on restart and restore/remove only bytes that still match durable BedRock intent while preserving mismatches as explicit conflicts; prove with crash/restart tests and executable CI.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
