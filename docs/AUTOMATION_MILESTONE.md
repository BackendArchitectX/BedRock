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
- Fresh reconciliation: `main` is `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; challenged branch product/test HEAD is `5772990df191b919381a5013f269050fedbcf3e2`, 28 commits ahead and 0 behind main. CURRENT remains M2.
- Inspect Build materially improved the state contract: `RecordMutationIntent` now fails closed unless the journal is already `MUTATING`; PREPARED can no longer contain mutation ownership/intended bytes.
- Challenged assumption: wiring the current `RecordMutationIntent` immediately before `ChangeSet.Apply` would still be unsafe. `RecordMutationIntent` captures a newly introduced path's original by reading the filesystem, while `Apply` independently captures its in-memory original later. An external write between those reads can give durable recovery and ordinary rollback two different pre-run originals.
- Stronger invariant: there must be one capture at the actual prepared write boundary. The exact captured original must be supplied to both durable ownership and the in-memory `ChangeSet`; after durable intent persistence succeeds, and only then, may `os.WriteFile` execute. Existing owned paths keep their immutable first original while later attempts update only intended content hash.
- Failure semantics: if durable intent persistence fails, do not write. If the file changes between preparation and the boundary capture, capture the boundary state once; if it changes after intent persistence but before/during write, later recovery must use intended-hash matching and preserve mismatches as conflicts rather than restoring blindly.
- Verification actually run this challenge: repository/branch state and source were inspected through the authenticated GitHub connector. A direct executable checkout was attempted and failed because this runtime could not resolve `github.com`; therefore no gofmt, vet, test, race, integration, or smoke PASS is claimed.
- Remaining acceptance gap: journal ownership is still not wired into the actual write boundary; there is no interrupted-run discovery/recovery or conflict-preserving durable restore/remove path, and no process-death test proves those semantics.
- Single next action: refactor `ChangeSet.Apply` to expose one boundary-owned original capture to a durable intent callback (MUTATING first, persist exact original + intended hash, then write), add focused TOCTOU/failure-injection coverage proving no write occurs when persistence fails, then wire that callback from `Engine.Run` before implementing restart recovery.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
