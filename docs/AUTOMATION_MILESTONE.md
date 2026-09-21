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

### Red-team handoff
- Fresh reconciliation: `main` is `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; branch product/test HEAD before this handoff is `83b7520d0b87c2668eb09e71bfe1eb213e5104d9`. CURRENT remains M2.
- Adversarial case exercised: a caller could previously invoke `RecordMutationIntent` while the durable journal still said `PREPARED`, creating contradictory evidence: mutation ownership/intended bytes existed while the state machine claimed mutation had not begun.
- Fix: `RecordMutationIntent` now accepts only `MUTATING`; PREPARED and COMPLETED both fail closed. Existing intent tests transition first, and `TestRecordMutationIntentRejectsPreparedJournal` asserts rejection leaves PREPARED state and intended hash unchanged.
- Verification: source-level regression was added, but no executable checkout was available in this runtime. No gofmt, vet, test, race, integration, or smoke PASS is claimed. The next executable worker must run gofmt on the touched Go tests before relying on CI because connector writes do not run formatting.
- Remaining risk: journal ownership is still not wired into the actual `ChangeSet.Apply` write boundary. There is still no interrupted-run discovery/recovery or conflict-preserving restore/remove path.
- Single next acceptance gap: integrate a journal transaction/callback into the real prepared `ChangeSet.Apply` write loop so MUTATING is durable first and each first-original + latest intended hash is persisted immediately before its corresponding write, then crash-inject at that boundary and prove restart recovery preserves external conflicts.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
