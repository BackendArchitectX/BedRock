# Automation Milestone Gate

> Development-pipeline coordination only. This file, worker roles, and scheduler cadence are not BedRock runtime or product architecture.

## CURRENT: M1 — Truthful outcomes and evidence

**Verdict:** IN PROGRESS — NOT ACCEPTED

### Gate baseline

- Main HEAD at gate initialization: `2d6aace2512eced6c784754c97aae81794caaa43`.
- Milestone branch: `automation/bedrock-current`.
- Main CI run `35527885270` for the initialization baseline completed successfully.
- Latest main was fetched again before this handoff and remained `2d6aace2512eced6c784754c97aae81794caaa43`; no reconciliation was required.

### M1 acceptance contract

M1 remains open until executable evidence demonstrates all of the following together:

- configured verification is run and recorded before BedRock edits and after edits;
- checks passing does not by itself become a task/goal-completion claim;
- an already-green repository plus a zero-change/no-op provider cannot produce a misleading task-success claim;
- provider summary plus useful attempt and failure evidence are persisted;
- persistent evidence includes sufficient before/after and diff information (or a stable diff hash plus changed paths) for later review/repair;
- existing safe-change and rollback behavior remains intact;
- focused behavior tests and applicable broader regression tests pass.

### Challenge pass

**Challenged assumption:** the pre-M1 engine treated only post-change verification as evidence and emitted `VERIFIED` whenever configured checks passed. That conflated check success with task completion and made an already-green/no-op run indistinguishable from an actual improvement. Provider response summaries were also discarded, weakening later repair/review evidence.

**Implementation HEAD before this handoff:** `bfd516fbd5df0c997b6e8e77314678e5b3c7c4e0`.

Changes on the milestone branch:

- `Evidence` now persists `baselineVerification` separately from post-change `verification` and retains per-attempt `providerSummaries`.
- The engine runs configured verification before provider execution, records its exact results, and passes baseline failure evidence into the first provider attempt so an initially-red repository can be repaired deliberately.
- Passing post-change checks now produce `CHECKS_PASSED`, not `VERIFIED`; this is intentionally a check-evidence statement rather than a task-completion claim.
- Successful post-change verification clears stale failure text so persisted evidence does not report a repaired baseline failure as the final failure.

### Verification actually performed

- Repository state, milestone file, engine, evidence types, CLI status output, and canonical demo launcher were inspected through the GitHub repository API.
- A local checkout was attempted for `gofmt`/`go vet`/`go test`, but the execution environment could not resolve `github.com`; therefore no local Go verification was executed and no pass is claimed.
- No workflow run was available for `bfd516f...` at handoff time, so CI is also not claimed.
- Known compatibility risk requiring the next worker's immediate attention: existing tests include a literal `VERIFIED` expectation and must be updated alongside focused M1 tests before this slice can be accepted.

### Remaining M1 acceptance gap

This materially advances M1 but does not close it. Missing executable evidence includes the already-green + zero-change case, explicit assertions for baseline/post-change persistence and provider summaries, regression updates for the status semantic change, and stable diff evidence (hash and/or persisted diff) suitable for later review/repair. Broader Go verification also remains unexecuted for this branch.

### One next action

Add focused engine tests for (1) already-green + no-op => `CHECKS_PASSED` with identical baseline/post verification and no task-success claim, and (2) initially-red => provider repair => passing post verification with baseline failure preserved; update the stale `VERIFIED` assertion, then run `gofmt`, `go vet ./...`, and `go test ./...`. Do not advance CURRENT beyond M1.
