# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

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
- `DirtyPaths` parsing needs dedicated coverage for rename/copy porcelain records and unusual filenames before stronger Git-safety claims.
- Provider execution is still a local subprocess rather than a hardened OS/container sandbox.
- Verification commands are user-supplied shell commands and execute with the user's local permissions.

### Next highest-value action

Prioritize a deterministic CLI integration test using a fake provider executable/script plus an isolated temporary Git repository. It should build/run the real `bedrock` command, prove a small edit + explicit verification + evidence flow, and cover rollback on a failed verification. Follow with dedicated `DirtyPaths` porcelain edge-case tests. Avoid adding provider SDKs or larger orchestration abstractions until this end-to-end boundary is proven.

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
- `DirtyPaths` parsing should receive dedicated tests for rename/copy porcelain records and unusual filenames before relying on it for stronger Git-safety guarantees.

### Next highest-value action

Add independent Git-safety tests around `DirtyPaths` (including rename/copy and filenames with spaces), then add a deterministic fake-provider CLI integration test that builds/runs the real `bedrock` command and proves a small repository edit + verification + evidence flow without a live LLM.
