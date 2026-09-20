# Automation Milestone Gate

> Development-pipeline coordination only. This file, worker roles, and scheduler cadence are not BedRock runtime or product architecture.

## CURRENT: M1 — Truthful outcomes and evidence

**Verdict:** IN PROGRESS — NOT ACCEPTED

### Gate baseline

- Main HEAD observed this pass: `2d6aace2512eced6c784754c97aae81794caaa43`.
- Milestone branch: `automation/bedrock-current`.
- Milestone implementation HEAD inspected: `e8053edbd361e6114d0c891fee4c0a3e2421f1cc`.
- Main remained unchanged; the milestone branch is based on that main baseline.

### M1 acceptance contract

M1 remains open until executable evidence demonstrates all of the following together:

- configured verification is run and recorded before BedRock edits and after edits;
- checks passing does not by itself become a task/goal-completion claim;
- an already-green repository plus a zero-change/no-op provider cannot produce a misleading task-success claim;
- provider summary plus useful attempt and failure evidence are persisted;
- persistent evidence includes sufficient before/after and diff information (or a stable diff hash plus changed paths) for later review/repair;
- existing safe-change and rollback behavior remains intact;
- focused behavior tests and applicable broader regression tests pass.

### Independent acceptance evidence this pass

GitHub Actions CI run `35538102190` executed against PR merge commit `89fd6a61884e0b88d34dbf409a1df565980b0c80`, combining milestone HEAD `e8053edbd361e6114d0c891fee4c0a3e2421f1cc` with unchanged main `2d6aace2512eced6c784754c97aae81794caaa43`.

The run FAILED and therefore M1 is rejected this cycle:

- Linux prerequisite-safety step passed.
- Linux `Format` failed because `internal/bedrock/engine_test.go` is not gofmt-clean. `go vet`, tests, Linux race, canonical demo, and rollback smoke were consequently skipped.
- Windows canonical launcher executed successfully through the BedRock run itself and emitted `status: CHECKS_PASSED` plus `BedRock demo: READY`, but the existing CI assertion still requires the obsolete `status: VERIFIED` string and failed with `missing VERIFIED status`.
- This Windows failure is an acceptance-suite compatibility regression caused by the intentional M1 truthful-status change, not evidence that M1 should revert to the misleading VERIFIED state. The gate must update the assertion to the truthful contract and then prove the full suite.

### M1 implementation under review

The branch contains the intended M1 direction: baseline/post verification evidence, truthful `CHECKS_PASSED` semantics rather than task-completion `VERIFIED`, no-op evidence behavior, provider summaries/attempt evidence, changed paths, and deterministic content-sensitive `diffHash` evidence without persisting raw patch content. Existing rollback/safe-change behavior is represented by focused tests but is not accepted until those tests actually execute successfully after the formatting gate is repaired.

### Remaining M1 acceptance gap

Do not merge or advance CURRENT. Two concrete blockers must be repaired on `automation/bedrock-current`:

1. run `gofmt` on `internal/bedrock/engine_test.go` and commit only the formatting result;
2. update the Windows canonical-launcher CI expectation from obsolete `status: VERIFIED` to the truthful M1 status contract (`CHECKS_PASSED`) without weakening the `READY`/result assertions.

After those repairs, require a fresh CI run to execute and pass format, vet, focused/broader tests, Linux race, canonical launcher, rollback/safety coverage, and Windows ordinary launcher execution. M1 remains NOT ACCEPTED until that executable evidence is green.

### One next action

Apply the two bounded acceptance repairs above, push the milestone branch, and use the resulting fresh CI run as the next Integrate gate. Do not advance to M2 before that run is green and the M1 behavioral contract is independently inspected.
