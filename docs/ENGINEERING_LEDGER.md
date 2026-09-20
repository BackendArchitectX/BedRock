# Engineering Ledger

## Current state

BedRock is a Go 1.22 local-first orchestration prototype on `main`. The current vertical slice accepts a task, gathers bounded repository context, invokes a provider-neutral command adapter, applies bounded file changes while protecting pre-existing dirty paths and repository metadata, runs explicit verification commands, retries once by default with failure evidence, and rolls changes back when verification never succeeds.

## 2026-09-20 one-step launcher cleanup safety

### Work completed

- Reconciled against current `origin/main` baseline `fa1bcc77a1224cf6f470f4beec4affa16e9396ba` before editing. That baseline's GitHub Actions run `35495019168` completed successfully, and user-reported Windows verification after rebase/push remains authoritative execution evidence for build/test/race/vet.
- Challenged the new canonical launcher against the product acceptance requirement to clean up partial/stale startup safely.
- Found a destructive boundary: `BEDROCK_DEMO_DIR` was caller-controlled and the launcher unconditionally executed `rm -rf "$work"`. A typo or intentionally broad path could therefore recursively delete unrelated user data.
- Replaced whole-workspace deletion with an ownership marker. Existing unmarked directories are rejected with an actionable diagnostic; BedRock only recreates its own `bin` and `repository` children after the workspace has been marked as launcher-owned.
- Added CI coverage that points `BEDROCK_DEMO_DIR` at an existing foreign directory, requires startup to fail, and proves a sentinel user file is preserved. Existing successful-start and rerun checks remain in the same canonical launcher job.

### Tests actually executed / evidence

- Before push, `sh -n` executed successfully against the revised launcher locally.
- A local shell safety scenario executed the revised guard against an existing foreign temporary directory: the launcher returned non-zero, emitted the refusal diagnostic, and preserved the sentinel file.
- Source commit: `05d1b33f941d6645a520ede22afe0270f9e6e2cb`.
- CI regression commit: `ba186ee7f58f6e4f2a9dc0d1fe86c46ed8329ac0`.
- GitHub Actions run `35495378765` for the regression HEAD was still in progress when last inspected. Therefore full format/vet/test/race/canonical-start/failure-rollback verification of this HEAD is **UNVERIFIED** until that run completes successfully.

### Remaining risks / next action

First inspect CI run `35495378765` (or newer current-HEAD CI) and repair any real failure before adding functionality. The canonical fresh-checkout launcher is still POSIX-shell-only even though the Go code itself has now been independently verified on Windows; the next one-step-start design decision should challenge whether a small cross-platform Go bootstrap (`go run ...`) can replace OS-specific launchers without duplicating runtime logic. Do not claim Windows one-step startup until the canonical launcher itself is exercised there. Provider subprocesses and explicit verification commands still execute with local user permissions; no sandbox claim should be made.

## Historical notes

Earlier detailed passes remain available in Git history. Current repository state and current CI evidence override stale PASS/UNVERIFIED claims from those notes.
