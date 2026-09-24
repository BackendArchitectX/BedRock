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
- Fresh reconciliation: `main` is `dc5b1d872d5eb11c6f4a58f83576cc689e3342e6`; product HEAD for this challenge is `b13950fc098242d49a16ef74f2b08f3b8ebad808`. The branch was 34 commits ahead and 0 behind before this product commit. CURRENT remains M2.
- Adversarial case exercised: the actual `ChangeSet.Apply` mutation boundary captured rollback state and immediately called `os.WriteFile`, leaving no fail-closed point where durable intent persistence could veto the filesystem write. A journal persistence failure therefore could not yet be proven to cause zero write.
- Fix: `ChangeSet.Apply` now delegates to an internal boundary that accepts an optional pre-write hook. The hook receives the exact already-captured original plus a copied intended payload and runs before `os.WriteFile`; hook failure triggers ordinary rollback and returns without writing the current target. Public `Apply` preserves existing behavior with no hook until Engine journal lifecycle wiring is added.
- Verification actually run this challenge: authenticated main/branch reconciliation and source/diff inspection. Local git checkout remained unavailable because DNS could not resolve github.com. GitHub Actions run `35650150994` was started for exact product HEAD; at handoff it was still in progress before Format/Test/Race, so no gofmt, vet, test, race, integration, or smoke PASS is claimed.
- Remaining risk: this boundary is not yet wired to `recordCapturedMutationIntent` or Engine journal state transitions, no focused hook-failure regression has executed yet, and interrupted-run discovery/recovery is absent. The product commit may still require gofmt correction depending on CI Format result.
- Single next acceptance gap: inspect exact-head CI first; if formatting fails, apply only gofmt-equivalent correction. Then add a focused regression proving pre-write persistence failure leaves the current target unchanged, adapt the hook to `recordCapturedMutationIntent`, and wire PREPARED→MUTATING lifecycle from `Engine.Run` without advancing M2.

### Blocker-recovery protocol
A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision. If direct git/DNS fails, use the authenticated GitHub connector when available. For CI failures inspect the exact failed job/step/log before repair. Never claim PASS for checks that did not execute. Preserve branch ownership and never force-push.
