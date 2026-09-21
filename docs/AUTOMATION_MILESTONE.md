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
- Fresh reconciliation: `main` is `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; challenged branch product/test HEAD is `73ede87c333ed5e5860bd96d97d53ff375ac21e6`. CURRENT remains M2.
- Challenged assumption: the journal API itself was still forcing a second filesystem read for newly owned paths. Even if `ChangeSet.Apply` captured the true pre-write original, passing only a path to `RecordMutationIntent` could silently replace that boundary truth with bytes observed later.
- Change made: added `recordCapturedMutationIntent`, which accepts the exact `JournalOriginal` captured by the future write boundary and never re-reads the target. For a newly owned path it persists that capture as the immutable first original; for an already-owned path it preserves the existing original and updates only the intended-content hash. It still fails closed unless the journal is `MUTATING` and validates the captured path against the journal repository.
- Focused coverage added: a TOCTOU regression changes the file after the simulated boundary capture and proves the durable journal retains the captured bytes rather than the later filesystem bytes; a second test proves PREPARED still rejects captured mutation intent.
- Verification actually run this challenge: repository/branch state and source were inspected through the authenticated GitHub connector. A direct executable checkout was attempted and failed because this runtime could not resolve `github.com`; therefore the new tests were NOT executed and no gofmt, vet, test, race, integration, or smoke PASS is claimed.
- Remaining acceptance gap: `ChangeSet.Apply` still performs its own capture and write with no durable callback, so the new primitive is not yet on the mutation path. There is still no proof that journal persistence failure causes zero filesystem write, no Engine wiring, no interrupted-run discovery/recovery, and no process-death conflict-preservation test.
- Single next action: refactor the actual `ChangeSet.Apply` write loop to capture one original, install that same capture into `ChangeSet.originals`, invoke a durable-intent callback with it before `os.WriteFile`, and abort without writing if the callback fails; add focused failure-injection coverage, then wire the callback from `Engine.Run`.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
