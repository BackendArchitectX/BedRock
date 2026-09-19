# Verification handoff

## 2026-09-20 independent verification pass

### Repository reality inspected

- Branch: `main`.
- Pre-pass HEAD: `516685af9efb3364c0eddd3620ba2e8ec323b01e`.
- Existing engineering ledger, CI workflow, engine implementation, deterministic fake provider, and recent Git-safety / CLI-smoke commits were inspected before changing anything.
- GitHub Actions run `35474715417` for the pre-pass HEAD completed successfully. Its job executed Format, Vet, Test, Race test, and CLI smoke successfully.

### Work completed

Added an independent negative-path CLI smoke check in commit `a6bfc6b6ac92484ad2790335ceb9612b858f3b8a`. It uses the built real `bedrock` CLI and deterministic fake provider against an isolated Git repository, deliberately fails verification, and requires all of the following:

- BedRock exits non-zero.
- the provider-created `result.txt` is rolled back;
- pre-existing untracked `user-work.txt` remains byte-for-byte intact;
- CLI output reports `status: FAILED`.

This closes an evidence gap left by the successful-path smoke test without adding product architecture.

### Tests actually executed

Before this change, Actions run `35474715417` completed successfully with Format, `go vet ./...`, `go test ./...`, `go test -race ./...`, and the successful real-CLI smoke all green.

For the new negative-path check, Actions run `35475066925` started for commit `a6bfc6b6ac92484ad2790335ceb9612b858f3b8a`. At the last inspection in this pass it was still in progress; checkout had completed and setup-go was running, while the verification steps were pending. Therefore the new rollback smoke is **UNVERIFIED** until that run completes successfully.

### Next action

Inspect run `35475066925` first. If it fails, repair the exact failing step. If it succeeds, consider moving the two CLI smoke scenarios out of workflow shell into a maintainable integration-test harness only if doing so materially improves portability or diagnostics. Higher-value remaining risks are provider/verifier subprocess sandboxing, verification-output secret handling, and Windows/macOS behavior; do not claim those are solved.
