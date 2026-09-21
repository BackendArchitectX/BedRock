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
- Reconciled against unchanged main `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; the branch was 8 commits ahead and 0 behind at inspected HEAD `b8f45023bc04dbdeb4aef70fe5681ad49d5e17e5`.
- Challenged assumption: adding forward journal states alone did not yet move crash recovery across the real mutation boundary. The new transition implementation was also not gofmt-normalized and had no transition regression test, so treating it as a ready foundation would overstate its evidence.
- Product commit `213cb39afb24a38c244940953a5db89e3387f4b9` normalizes `journal.go` to gofmt-equivalent source without changing the transition semantics. Product-test commit `53ce5dfa8f4142003c20ed99926261f33eba9946` adds persistence coverage for `PREPARED -> MUTATING -> COMPLETED` and rejects a backward transition after completion.
- Tests actually run in this invocation: none. Repository writes were performed through the GitHub connector and no executable checkout was available, so formatting equivalence is source-inspected only and no compile/test/CI PASS is claimed.
- Remaining acceptance gap: the journal is still not wired into `ChangeSet.Apply`; in-memory rollback remains authoritative. A naive per-attempt call to `PrepareRunJournal` with the same run ID would overwrite earlier originals when later repair attempts introduce new paths, so wiring must preserve one run's complete write-ahead ownership set rather than silently replacing it.
- Single next action: design the smallest append-or-prepare-once mutation-boundary API that preserves originals across attempts, wire `PREPARED/MUTATING/COMPLETED` around actual writes, then add process-death/restart conflict-preservation coverage.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
