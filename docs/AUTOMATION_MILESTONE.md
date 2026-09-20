# Automation Milestone Gate

> Development-pipeline coordination only. This file, worker roles, and scheduler cadence are not BedRock runtime or product architecture.

## CURRENT: M1 — Truthful outcomes and evidence

**Verdict:** IN PROGRESS — NOT ACCEPTED

### Gate baseline

- Main HEAD observed this pass: `2d6aace2512eced6c784754c97aae81794caaa43`.
- Milestone branch: `automation/bedrock-current`.
- Branch HEAD before this handoff update: `1470715303afd6d18d12073aaf6f486b7c9c16c5`.
- Main remained unchanged while the milestone branch advanced; no main reconciliation conflict was observed.

### M1 acceptance contract

M1 remains open until executable evidence demonstrates all of the following together:

- configured verification is run and recorded before BedRock edits and after edits;
- checks passing does not by itself become a task/goal-completion claim;
- an already-green repository plus a zero-change/no-op provider cannot produce a misleading task-success claim;
- provider summary plus useful attempt and failure evidence are persisted;
- persistent evidence includes sufficient before/after and diff information (or a stable diff hash plus changed paths) for later review/repair;
- existing safe-change and rollback behavior remains intact;
- focused behavior tests and applicable broader regression tests pass.

### Independent challenge this pass

The branch's M1 evidence still recorded changed paths but had no content-sensitive diff evidence. Two runs that touched the same path with different patches were therefore indistinguishable from persisted evidence alone, preventing later review from proving which mutation was actually attempted/applied without rereading mutable repository state.

Focused repair through `1470715303afd6d18d12073aaf6f486b7c9c16c5`:

- adds a deterministic SHA-256 `diffHash` over sorted changed paths plus before/after content digests;
- persists the hash alongside changed paths without copying raw changed content into evidence;
- leaves zero-change runs with no diff hash, preserving truthful no-op semantics;
- adds regressions requiring identical mutations to hash identically, different mutations to hash differently, engine evidence to include the hash for a changed path, and persisted evidence not to contain raw patch content.

### Verification actually performed

- Fetched current `main` and `automation/bedrock-current` and re-fetched `main` after the edits; main remained `2d6aace2512eced6c784754c97aae81794caaa43`.
- Inspected the M1 contract, engine finish/evidence path, evidence schema, ChangeSet original/written tracking, and existing focused tests.
- Attempted a fresh local clone again; the execution environment still failed before checkout because DNS could not resolve `github.com`.
- GitHub reports no commit status checks for `1470715303afd6d18d12073aaf6f486b7c9c16c5`; therefore no `gofmt`, compile, `go vet`, `go test`, integration, or CI PASS is claimed for these changes.

### Remaining M1 acceptance gap

The branch still needs executable formatting/compilation and focused/broader Go verification. The prior `engine_test.go` commit also requires `gofmt`; its compact formatting remains unverified. Integrate must not accept M1 until the branch is actually formatted, compiled, and tested and the persisted `diffHash` regression passes.

### One next action

Obtain an executable checkout, run `gofmt` on the modified Go files (especially `engine_test.go`), then run `go test ./internal/bedrock`, `go vet ./...`, and `go test ./...`; correct any compile/test failures without advancing CURRENT beyond M1.
