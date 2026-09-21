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
- Fresh reconciliation: `main` is `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; challenged branch product/test HEAD is `78bf9a1a0226b9dd34c317b50d2872a08a3b5638`. CURRENT remains M2.
- Adversarial case exercised: the boundary-captured journal primitive trusted a caller-supplied `JournalOriginal` whose `SHA256` could contradict its `Content`, or whose `Existed=false` capture could still carry file bytes/mode/hash. That could durably persist internally contradictory recovery evidence before any filesystem write and later make ownership/conflict decisions depend on evidence BedRock never validated.
- Fix: `recordCapturedMutationIntent` now validates captures before adding or updating journal ownership. Existing files require a SHA-256 matching the exact captured content; non-existent originals must not carry content, mode, or hash metadata. Invalid captures fail before journal persistence.
- Regression coverage added: one test supplies an intentionally wrong hash for captured bytes and proves the journal gains no original; another supplies content for an `Existed=false` capture and requires rejection. Existing TOCTOU and PREPARED-state coverage remains.
- Verification actually run this challenge: authenticated GitHub branch/main reconciliation plus exact source inspection. This runtime still has no executable checkout/Go toolchain path, so the new tests were NOT executed and no gofmt, vet, test, race, integration, or smoke PASS is claimed.
- Remaining risk: `ChangeSet.Apply` still captures/writes without invoking the durable primitive, Engine does not yet prepare/transition/complete the journal around real mutation, and interrupted-run discovery/recovery is absent. Thus this hardening prevents poisoned durable evidence but does not yet satisfy M2 crash recovery.
- Single next acceptance gap: refactor the actual `ChangeSet.Apply` write boundary so one capture feeds both in-memory rollback and `recordCapturedMutationIntent` before `os.WriteFile`, prove callback/persistence failure causes zero write, then wire journal lifecycle from `Engine.Run`.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
