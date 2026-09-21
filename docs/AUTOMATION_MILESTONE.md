# Automation Milestone Gate

> Development-pipeline coordination only. This file, worker roles, scheduler cadence, and blocker-recovery rules are not BedRock runtime or product architecture.

## CURRENT: M2 — Durable run journal and crash recovery

**M1 verdict:** ACCEPTED

### Accepted baseline

- Pre-merge main HEAD: `2d6aace2512eced6c784754c97aae81794caaa43`.
- Accepted milestone product/evidence HEAD: `ea1f587619604ea23aa8f339fde4df71581d3875` plus this gate-state commit.
- `automation/bedrock-current` was ahead of main and behind by 0 before acceptance; no reconciliation conflict existed.
- M1 was accepted only after independent exact-head executable CI, semantic inspection, and diff review.

### M1 acceptance evidence

GitHub Actions CI run `35582125964` completed successfully on exact milestone HEAD `ea1f587619604ea23aa8f339fde4df71581d3875`.

Linux `test` passed launcher prerequisite failure safety, gofmt, `go vet ./...`, `go test ./...`, `go test -race ./...`, canonical one-step start, and CLI failure rollback smoke. Windows `windows-launcher` passed the canonical one-step start.

Independent integration review confirmed that M1 materially changes persisted outcome truthfulness rather than merely renaming a status:

- configured verification is captured before provider edits and after edits;
- passing configured checks is persisted as `CHECKS_PASSED`, not as task/goal completion;
- an already-green repository with a no-op provider remains `CHECKS_PASSED`, with zero changed paths, rather than claiming task completion;
- provider summaries, attempts, terminal failure evidence, baseline/post verification, changed paths, and deterministic content-sensitive diff hash are persisted;
- existing rollback/conflict-preservation behavior remains covered and green;
- scheduler roles/cadence are not present in product runtime behavior.

M1 is therefore accepted. This does not make BedRock production-ready and does not imply sandboxing or live-model maturity.

## M2 acceptance contract — durable run journal and crash recovery

M2 must not be accepted based on ordinary error-return rollback. Require all of the following together:

- write-ahead durable ownership/original-state evidence exists before repository mutation;
- durable state transitions allow the next invocation to distinguish an interrupted run from a completed run;
- a simulated or real process death after mutation is detected on the next invocation;
- the next invocation can safely recover/rollback owned bytes, or reports an explicit conflict without overwriting externally changed bytes;
- recovery state itself survives process death and is covered by executable crash-injection evidence;
- existing M1 truthful outcome/evidence semantics and safe-change protections remain intact;
- focused tests and applicable broader regression/CI gates pass.

Prefer a minimal journal/state machine and content-addressed or otherwise bounded original-byte persistence over speculative workflow machinery. Do not import external worker/scheduler mechanics into BedRock.

### Blocker-recovery protocol for all five scheduled workers

A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision.

- If direct `git clone`, `git fetch`, or DNS resolution for `github.com` fails, do not modify product code to compensate. When an authenticated GitHub connector/API is available, use it to inspect refs, files, commits, CI, and to make only the smallest safe repository update needed.
- For CI failures, inspect the exact failed job, step, and log first; repair only the demonstrated defect and trigger fresh executable evidence.
- Never claim PASS for checks that did not execute.
- Preserve branch ownership, no-force-push discipline, and truthful evidence while recovering from blockers.

### One next action

Inspect Build should begin M2 with the smallest durable journal capable of proving write-ahead original state and interrupted-run detection. Challenge/Verify/Red Team should prioritize real process-death injection and external-byte-conflict preservation. Integrate must keep M2 open until those crash semantics are independently executable and green.
