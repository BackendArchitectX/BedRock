# Latest engineering handoff

## 2026-09-20 — integration and release-quality pass

### Current state

BedRock remains a Go 1.22 local-first orchestration prototype on `main`. Current HEAD before this pass was `95862c177aa65f01e133c281c7ed55108c0728fb`. GitHub Actions run `35483701580` completed successfully on that exact commit: format, vet, unit tests, race tests, successful CLI smoke, and failed-verification rollback smoke all passed.

### Work completed

- Re-inspected current `main`, recent commits, repository tree, CI, README, `docs/ENGINEERING_LEDGER.md`, and the separate handoff files rather than trusting prior scheduled-run claims.
- Confirmed the latest nested-root dirty-path correction is present and green on current HEAD. `DirtyPaths` scopes Git status to the selected run root and translates Git paths into the same coordinate space used by provider changes.
- Identified documentation entropy: `CURRENT.md`, `VERIFY_HANDOFF.md`, and `RED_TEAM_HANDOFF.md` had become stale competing sources of "current" state while their durable history already exists in `ENGINEERING_LEDGER.md`.
- Consolidated continuation state here and removed those three stale handoff files. `ENGINEERING_LEDGER.md` remains the append-only engineering history; this file is the single concise current handoff.
- No runtime architecture, dependency, scheduler concept, or product feature was added in this pass.

### Verification actually observed

GitHub Actions run `35483701580` on pre-pass HEAD `95862c177aa65f01e133c281c7ed55108c0728fb`: **PASS**.

Executed CI stages:

- format: PASS
- `go vet ./...`: PASS
- `go test ./...`: PASS
- `go test -race ./...`: PASS
- real built-CLI success smoke with deterministic fake provider: PASS
- failed-verification rollback smoke preserving pre-existing user work: PASS

No local Go execution is claimed; this pass used the connected GitHub repository/API. This documentation-only cleanup still requires its own CI run before the resulting HEAD should be described as green.

### Remaining risks

- Provider subprocesses execute with local-user permissions; BedRock does not provide an OS/container sandbox.
- Verification commands are intentionally user-supplied shell commands and execute with local-user permissions.
- Dirty paths are captured before provider execution. A user/process editing a previously clean target after that snapshot but before BedRock writes it is a remaining concurrent-mutation/data-loss risk.
- No live OpenAI/Anthropic/local-model adapter end-to-end run is claimed; CI uses a deterministic fake provider.
- Windows and macOS behavior remain unverified.
- Secret filtering/redaction is defense in depth, not a proof that arbitrary secrets cannot appear in ordinary tracked source or command output.

### Next highest-value action

First inspect CI for the current documentation-cleanup HEAD. If green, prioritize concurrent-mutation protection at the file-write boundary: BedRock should detect when a target changed after its run snapshot / after BedRock's previous write and refuse to overwrite external user changes. Prove the behavior with deterministic regression tests before adding broader provider or orchestration features.
