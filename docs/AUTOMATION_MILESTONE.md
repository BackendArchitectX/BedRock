# Automation Milestone Gate

> Development-pipeline coordination only. This file, worker roles, scheduler cadence, and blocker-recovery rules are not BedRock runtime or product architecture.

## CURRENT: M1 — Truthful outcomes and evidence

**Verdict:** IN PROGRESS — CI GREEN — INTEGRATE MUST ACCEPT OR REJECT

### Gate baseline

- Main HEAD: `2d6aace2512eced6c784754c97aae81794caaa43`.
- Milestone branch: `automation/bedrock-current`.
- Current milestone HEAD: `a22c0a11bf784746ae4a4128057a6ca2e15dc171`.
- Branch comparison after the fixes: ahead of main by 18 commits and behind by 0.

### M1 acceptance contract

M1 remains open until Integrate independently confirms all of the following together:

- configured verification is run and recorded before BedRock edits and after edits;
- checks passing does not by itself become a task/goal-completion claim;
- an already-green repository plus a zero-change/no-op provider cannot produce a misleading task-success claim;
- provider summary plus useful attempt and failure evidence are persisted;
- persistent evidence includes sufficient before/after and diff information (or a stable diff hash plus changed paths) for later review/repair;
- existing safe-change and rollback behavior remains intact;
- focused behavior tests and applicable broader regression tests pass.

### Blockers repaired in this recovery pass

1. `internal/bedrock/engine_test.go` was not gofmt-clean and blocked all downstream Linux gates. It was formatted without changing intended runtime behavior in commit `4a5527ea9e799a4a4f0e9c7b35fc4be3299a3d0f`.
2. Fresh CI then exposed a stale recovery smoke assertion in `internal/demosmoke/verification_recovery_test.go` that still expected `status: VERIFIED`. It was aligned with the truthful M1 status `CHECKS_PASSED` in commit `a22c0a11bf784746ae4a4128057a6ca2e15dc171`.
3. The previously stale CI workflow assertions for Linux/Windows already use `CHECKS_PASSED`; the launcher prerequisite fixture remains `go env GOVERSION`.

### Fresh executable evidence

GitHub Actions CI run `35581993953` completed successfully on the current milestone head.

Linux `test` job passed all of:
- launcher prerequisite failure safety;
- gofmt gate;
- `go vet ./...`;
- `go test ./...`;
- `go test -race ./...`;
- canonical one-step start;
- CLI failure rollback smoke.

Windows `windows-launcher` job also passed the canonical one-step start.

Do not reinterpret this as automatic milestone acceptance. Integrate still must inspect the M1 semantics, persisted evidence, and branch diff before merging or advancing CURRENT.

### Blocker-recovery protocol for all five scheduled workers

A transient environment/tool failure is a blocked invocation, not a terminal pipeline decision.

- If direct `git clone`, `git fetch`, or DNS resolution for `github.com` fails, do not modify product code to compensate. When an authenticated GitHub connector/API is available, use it to inspect refs, files, commits, PR CI, job logs, and to make only the smallest safe repository update needed.
- If one access path is unavailable but another owner-authorized repository path is available, use the available path instead of declaring the milestone permanently blocked.
- For CI failures, inspect the exact failed job, step, and log first; repair only the demonstrated defect and trigger fresh executable evidence.
- Never claim PASS for checks that did not execute.
- If a write, merge, reconciliation, or verification cannot be performed safely in the current invocation, checkpoint the exact blocker and stop only that invocation. The next scheduled worker/run must retry from current shared state.
- Do not disable, terminate, or treat the five-worker pipeline as complete because one run is blocked. Only an explicit user instruction or an intentionally configured scheduler deadline should end the recurring pipeline.
- Preserve branch ownership, no-force-push discipline, and truthful evidence while recovering from blockers.

### Remaining M1 acceptance gap

The infrastructure/CI blockers are cleared. The only remaining M1 gate is independent Integrate review of the actual M1 behavior and evidence contract on current head `a22c0a11bf784746ae4a4128057a6ca2e15dc171`.

### One next action

Integrate should inspect the current branch diff and fresh green CI evidence. If and only if the complete M1 acceptance contract is satisfied, merge/reconcile into latest main, record M1 ACCEPTED, advance CURRENT to M2, and recreate/reset `automation/bedrock-current` from the accepted main. Otherwise record the exact semantic gap and keep M1 open.
