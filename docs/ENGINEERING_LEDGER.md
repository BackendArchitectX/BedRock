# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

## 2026-09-20 Git-aware context hardening

### Work completed

- Re-inspected current `main`, recent commits, CI, context selection, tests, and this ledger rather than carrying forward prior PASS claims.
- Confirmed preceding HEAD `470786989a7d623e03025417cad04b5df574f93a` had successful CI run `35479303942`.
- Challenged the growing credential filename denylist and found a broader repository-reality gap: `Snapshot` walked the filesystem directly, so files intentionally excluded by `.gitignore` could still be sent to the model provider.
- Kept the fix small: for Git worktrees, context candidates are now constrained to `git ls-files -co --exclude-standard -z` (tracked files plus non-ignored untracked files). Non-Git directories retain the existing filesystem behavior.
- Added a regression test proving ordinary tracked/untracked context remains available while directory- and glob-ignored local files do not enter provider context.

### Tests actually executed / evidence

- No local Go execution is claimed because this run used the connected GitHub repository API rather than a mounted checkout.
- CI run `35479867536` started on regression-test HEAD `fff324804f4732e54f64b3aeac282274adf156b5` and was still `in_progress` when last inspected. Therefore this change remains **UNVERIFIED** until that run completes successfully.

### Remaining risks / next action

First inspect CI run `35479867536` and repair any format/vet/test/race/CLI-smoke failure before adding features. If green, challenge Git worktree detection for `.git` files used by linked worktrees/submodules: the current implementation detects `.git` existence but `git ls-files` should be independently verified in those layouts. Filename filtering remains defense in depth; tracked ordinary files may still contain secrets. Provider subprocesses and verification shell commands still run with local user permissions, so no sandbox claim should be made.

## 2026-09-20 credential-context hardening pass

### Work completed

- Re-inspected current `main`, recent commits, CI, context selection, tests, TODO search, and this ledger rather than carrying forward old PASS claims.
- Confirmed CI run `35478912967` succeeded on the preceding context-filter HEAD `2bc70a1dcadbe775ea34f79fecf0e105df6c6c2d`.
- Found remaining provider-context credential exposures not covered by the filename filter: `.aws`, `.azure`, `.docker`, `.kube`, and `.gnupg` credential directories; `.git-credentials`; and Terraform state files.
- Extended the existing small denylist instead of adding content-scanning or a new security subsystem. Added regression fixtures for each new class and retained safe ordinary source/configuration examples.

### Tests actually executed / evidence

- No local Go execution is claimed because this run used the connected GitHub repository API rather than a mounted checkout.
- GitHub Actions run `35479268886` started for source commit `d3fa36078f33c059a5906291d59d15c0834edd68` and was still `in_progress` when last inspected.
- The subsequent regression-test commit `17c521266432b5d36009c4e92729a50d28da48f0` did not yet have a visible completed CI run when this note was written. Therefore the new source + tests remain **UNVERIFIED** until CI executes against a commit containing both.

### Remaining risks / next action

First inspect CI on current HEAD and repair any formatter/compiler/test/race/smoke failure before adding features. Then challenge context safety against additional high-value credential locations without turning the denylist into broad false-positive filtering. Filename filtering is defense in depth, not a guarantee that arbitrary ordinary source/config files contain no secrets. Provider subprocesses and user-supplied verification commands still run with local user permissions; no sandbox claim should be made.

## 2026-09-20 CLI vertical-slice challenge pass

### Work completed

- Re-inspected current `main`, recent commits, repository tree, CI, README/ledger, and the prior Git-safety work rather than assuming earlier runs succeeded.
- Confirmed the preceding Git-safety changes reached green CI on current history before selecting new work.
- Challenged the backlog and kept the next objective narrow: prove the existing architecture end to end rather than add providers, agents, persistence, or orchestration abstractions.
- Added a deterministic fake provider under `cmd/bedrock/testdata/fakeprovider` that consumes the real provider request protocol and emits one bounded file edit.
- Extended CI to run `go test -race ./...`, build the real `bedrock` CLI and fake provider, initialize an isolated Git repository, execute `bedrock run`, verify the resulting file, and require the CLI to report `status: VERIFIED`.

### Tests actually executed / evidence

- GitHub Actions run `35474682080` on commit `1a210b3839e953625ca08bcb785bcbf69b740cd3` executed and **FAILED** at the format gate because the newly added fake-provider source was not gofmt-aligned. Vet, unit tests, race tests, and CLI smoke were therefore skipped in that run.
- The exact gofmt diff was inspected and fixed in commit `f5bfa41b402e730f61d6e12b094f83832d001cdf`.
- Follow-up run `35474699281` was queued when this note was written. Therefore race detection and the real CLI smoke remain **UNVERIFIED** in this pass until that run completes successfully.

### Remaining risks / next action

First inspect run `35474699281`. If it fails, repair the actual failing step before adding features. If it passes, the next highest-value challenge is failure-path CLI coverage: prove that a provider edit is rolled back when verification fails, and that the CLI exits non-zero while preserving pre-existing user work. Provider subprocesses and user-supplied verification commands still execute with local user permissions; no sandbox claim should be made.

## 2026-09-20 Git-safety hardening pass

### Work completed

- Re-inspected current `main`, recent commits, CI history, source, tests, README, and this ledger rather than assuming prior scheduled work happened.
- Identified a concrete safety bug in `DirtyPaths`: the old parser treated each NUL-delimited token as a complete porcelain record. Git porcelain v1 `-z` emits a second source-path token for rename/copy records, so that source path was not protected. It also used `TrimSpace`, which can corrupt legitimate filenames containing leading/trailing spaces.
- Replaced the ad-hoc parsing with a fail-closed `parsePorcelainV1Z` parser. It preserves path bytes represented as Go strings, protects both destination and source paths for rename/copy records, and rejects malformed/truncated records instead of silently weakening dirty-work protection.
- Added deterministic regression tests for filenames containing spaces, both rename paths, and truncated rename records.

### Tests actually executed / evidence

- No local Go execution is claimed in this pass because repository access is through the connected GitHub API rather than a mounted checkout.
- Before these changes, the latest completed `main` CI run inspected was `35473306842` on commit `2f407296eed0e352baecc047cc55f1a93026b605`, conclusion `success`.
- GitHub Actions run `35473900283` started for source commit `fb7c97318c1bb1dece38a91ebff5d7da801fa399` and was still `in_progress` when inspected. The follow-up regression-test commit had not yet produced a visible run at that instant. Therefore the new parser and tests remain **UNVERIFIED** until CI executes on a commit containing both changes.

### Remaining unverified / risks

- Current Git-safety parser changes still need successful CI (`gofmt`, `go vet ./...`, `go test ./...`).
- `go test -race ./...` has not yet been executed successfully.
- No deterministic real-CLI end-to-end smoke scenario has been committed yet.
- No real provider adapter end-to-end smoke test has executed; provider tests are deterministic/unit-level.
- Windows/macOS behavior is unverified.
- Provider execution is still a local subprocess rather than a hardened OS/container sandbox.
- Verification commands are user-supplied shell commands and execute with the user's local permissions.

### Next highest-value action

First verify CI on current HEAD and repair any formatter/compiler/test failure. Once green, add a deterministic CLI integration test using a fake provider executable/script plus an isolated temporary Git repository. It should exercise the real `bedrock run` command, prove a small edit + explicit verification + evidence flow, and cover rollback on failed verification without a live LLM.

## 2026-09-20 integration / release-quality pass

### Work completed

- Re-inspected current `main`, recent commits, CI history, README, CLI, and the prior verification ledger rather than assuming earlier automation succeeded.
- Confirmed CI run `35472469468` completed successfully on commit `98f8b5cda6d0ae76ff82437742450d0c985c9702`; the immediately preceding independent verification recorded successful `gofmt`, `go vet ./...`, and `go test ./...` on the implementation.
- Replaced the one-line README with factual installation, build/test, CLI, provider-adapter contract, credential handling, implemented safeguards, and current limitations. The README explicitly labels BedRock an early prototype and does not claim built-in model-provider integrations that do not exist.
- Kept the product architecture local-first and provider-neutral; no infrastructure or runtime behavior was added for external development automation.

### Tests actually executed / evidence

No new source code was changed in this pass, so no new local test execution is claimed. The latest fully completed CI evidence inspected during this pass is GitHub Actions run `35472469468` on commit `98f8b5cda6d0ae76ff82437742450d0c985c9702`, which succeeded. The README-only commit created by this pass had not yet appeared in the Actions run listing at the time of inspection, so its CI status remains pending/unobserved rather than being reported as passing.

### Previous failures already fixed and retained

1. Literal NUL source byte that prevented Go tooling from parsing `changes.go`.
2. Opaque formatter failure diagnostics in CI.
3. Provider subprocess inheriting the entire parent environment; credentials now require explicit environment-variable opt-in and explicitly passed values are redacted from provider failure stderr.

### Remaining unverified / risks

- `go test -race ./...` has not yet been executed successfully.
- No deterministic real-CLI end-to-end smoke scenario has been committed yet.
- No real provider adapter end-to-end smoke test has executed; provider tests are deterministic/unit-level.
- Windows/macOS behavior is unverified.
- Provider execution is still a local subprocess rather than a hardened OS/container sandbox.
- Verification commands are user-supplied shell commands and execute with the user's local permissions.

### Next highest-value action

Prioritize a deterministic CLI integration test using a fake provider executable/script plus an isolated temporary Git repository. It should build/run the real `bedrock` command, prove a small edit + explicit verification + evidence flow, and cover rollback on a failed verification. Avoid adding provider SDKs or larger orchestration abstractions until this end-to-end boundary is proven.

## 2026-09-20 independent verification

### Work completed

- Inspected current `main`, recent commits, repository tree, CI workflow, runtime/provider/context/change code, and tests.
- Found that CI had never reached `go vet` or `go test`: `internal/bedrock/changes.go` contained a literal NUL byte in Go source, so `gofmt` failed immediately.
- Replaced the invalid source byte with the valid `"\\x00"` delimiter used to parse `git status -z` output.
- Applied `gofmt`-equivalent formatting to `changes.go` and `provider.go`.
- Kept the CI format check diagnostic: on future failures it prints the files and formatter diff instead of only exiting silently.
- Independently reviewed the provider boundary added immediately before this verification run. It now inherits only a small operational environment allowlist by default; additional variables require explicit `--provider-env NAME`, and explicitly passed values are redacted from provider failure stderr.

### Tests actually executed

GitHub Actions run `35472428815` on commit `8f485d4c9a6aa2e0675e6beefd52cfd0e9c44560` executed successfully on Ubuntu 24.04 with Go 1.22.12:

- `gofmt -l .` cleanliness check: PASS
- `go vet ./...`: PASS
- `go test ./...`: PASS

A local clone/build/race-test attempt was made from the automation container, but outbound DNS to github.com was unavailable, so those commands did **not** execute against the repository and are not counted as passing verification.

### Failures discovered and fixed

1. **Build/format blocker:** literal NUL byte in `changes.go` prevented Go tooling from parsing the repository. Fixed.
2. **Opaque CI diagnostics:** formatter failures did not identify offending files. CI now reports files and diffs.
3. **Provider credential exposure risk:** provider subprocesses previously inherited the entire parent environment. The current implementation restricts inherited variables and requires explicit opt-in for provider credentials; tests cover allow/block behavior and redaction.

### Remaining unverified / risks

- `go test -race ./...` has not yet been executed successfully.
- No real provider adapter end-to-end smoke test has executed yet; provider tests are deterministic/unit-level.
- CLI behavior has not yet been smoke-tested from a built binary in CI.
- Windows/macOS behavior is unverified.

### Next highest-value action

Add a deterministic fake-provider CLI integration test that builds/runs the real `bedrock` command and proves a small repository edit + verification + evidence flow without a live LLM.