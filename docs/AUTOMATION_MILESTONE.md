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

### Red Team handoff
- Fresh reconciliation: `main` is unchanged at `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; `automation/bedrock-current` was `02d112e585720f5c7c08c44980f3aac33d182763` before this handoff and was 19 commits ahead / 0 behind main. CURRENT remains M2.
- Adversarial case: challenged the proposed "journal created directories before/around MkdirAll" repair itself. A naive `CreatedDirectories` or planned-directory list is not sufficient ownership evidence: after BedRock observes a parent as absent and persists intent, an external actor can create that directory before BedRock calls `MkdirAll`. `MkdirAll` then succeeds without telling BedRock whether it created the directory, and recovery could later delete an externally-created empty directory while falsely treating it as BedRock-owned. Conversely, recording ownership only after `MkdirAll` leaves a process-death window between creation and durable ownership.
- Confidence/failure evidence: current `RunJournal` contains only file originals/intended hashes and `ChangeSet.Apply` is not yet journal-integrated. No created-directory ownership primitive exists, so the previously identified directory-created/file-not-written crash remains unresolved. The race above shows that merely adding a directory path list would not safely close it.
- Product fix in this invocation: none. The safe fix needs an ownership boundary that can actually be proven; adding a misleading ownership field would make M2 less truthful. The smallest provable option to evaluate first is to forbid BedRock from creating missing parent directories during the initial M2 integration (require the parent to pre-exist), keeping all crash-owned mutation at the file boundary. If product compatibility requires directory creation, it needs a stronger creation protocol with explicit conflict-safe recovery rather than `Lstat -> journal -> MkdirAll` inference.
- Tests actually run: none in this invocation; no executable checkout was available through the authenticated repository interface, so no gofmt/vet/test/race/smoke PASS is claimed.
- Remaining risk: journal APIs are still not wired into the real mutation path, restart recovery is not implemented, and parent-directory creation remains outside durable ownership.
- Single next acceptance gap: establish the parent-directory invariant at the real write boundary and add a regression proving a missing parent is rejected without filesystem mutation (or, only if compatibility proves that unacceptable, implement a race-safe directory ownership protocol); then wire `PREPARED -> MUTATING`, intended hashes, interrupted-run detection, and conflict-preserving recovery around that bounded mutation surface.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.