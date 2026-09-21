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
- Fresh reconciliation: `main` remains `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; `automation/bedrock-current` was `0a3522044a0f149632e0b505b9d219ec80f409fe` before this handoff. CURRENT remains M2 and no newer main work required reconciliation.
- Challenged assumption: the bounded parent-directory change materially reduces the crash surface and is sound as an M2 safety constraint, but it does not itself move the core recovery gate. The next integration must not merely call `RecordMutationIntent` before `ChangeSet.Apply`: that would persist an original at one time while `Apply` can capture a different in-memory original later, creating two competing ownership truths if another process changes the file between those operations.
- Required sequencing invariant: one durable mutation boundary must own both original capture and intended bytes. Transition the journal `PREPARED -> MUTATING` before the first owned mutation; for each file, capture/persist its first original and latest intended SHA-256 immediately at the actual write boundary, then write. Later repair attempts may update only intended SHA-256, never the first original. A crash after `MUTATING` but before any file intent is still an interrupted run with zero owned file mutations and must be recoverable as such.
- Additional challenge: `RecordMutationIntent` currently permits calls while the journal is still `PREPARED`. The engine integration must not rely on that permissiveness; otherwise a caller can persist mutation ownership while the durable state still says no mutation phase began. Tighten the API or prove the engine always transitions first before treating M2 as accepted.
- Product changes in this invocation: none. I did not add a pre-`Apply` journal call because it would create the stale-original race above and make recovery semantics weaker while appearing more durable.
- Tests actually run: none. Direct executable checkout failed because this runtime could not resolve `github.com`; no gofmt, vet, test, race, or smoke PASS is claimed.
- Remaining acceptance gap: journal ownership is still not wired into the real `ChangeSet.Apply` write boundary; interrupted-run discovery/recovery and conflict-preserving restore/remove remain absent.
- Single next action: integrate a journal callback/transaction object inside the prepared `ChangeSet.Apply` write loop so `MUTATING` is durable first and each first-original + intended hash is persisted immediately before its `os.WriteFile`; then add focused crash/restart tests before implementing broader discovery/recovery.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
