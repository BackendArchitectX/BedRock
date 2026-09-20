# Latest engineering handoff

## 2026-09-20 — interruptible CLI runs

### Current state

BedRock remains a Go 1.22 local-first orchestration prototype on `main`. Before this pass, current HEAD `a3e462ae4996a9e1dbfaa13793bdaabcbf60a795` had successful GitHub Actions run `35489221373`, including the repository's format, vet, unit, race, and CLI smoke workflow.

### Work completed

- Re-inspected current `main`, recent commits, CI, `docs/ENGINEERING_LEDGER.md`, `docs/LATEST_HANDOFF.md`, runtime write/rollback behavior, context selection, verifier secret redaction, CLI, and TODO/FIXME search.
- Rejected adding a locking/distributed subsystem for the documented final check-to-write race: no small cross-platform primitive in the current design would make the whole filesystem operation transactionally safe, so the limitation remains explicit.
- Closed a recovery/operability gap instead: `bedrock run` previously invoked the engine with `context.Background()`, so the provider and verifier cancellation paths could not be driven by CLI termination.
- The CLI now creates a signal-aware context for `os.Interrupt` and `SIGTERM` and passes it through the existing engine/provider/verifier context chain. This lets normal interactive interruption and process-manager termination request cancellation without adding any scheduler-derived runtime behavior.

### Verification actually observed

- Baseline GitHub Actions run `35489221373` on `a3e462ae4996a9e1dbfaa13793bdaabcbf60a795`: **PASS**.
- Source commit: `b679603b164eefa1413f878e3911cff7de805435` (`fix(cli): cancel active runs on termination`).
- CI for `b679603b...` had not appeared in the Actions listing when this handoff was written. Therefore the new CLI change is **UNVERIFIED** until a run containing it completes successfully.
- No local Go execution is claimed; this pass used the connected GitHub repository/API.

### Remaining risks / next action

First inspect CI for `b679603b...` or its current descendant and repair any real format/vet/unit/race/CLI-smoke failure. If green, add a deterministic process-level interruption regression on Linux only if it can be kept stable; otherwise avoid flaky signal tests and move to the next highest-value product capability. The narrow final dirty-check-to-write race remains documented. Provider and verification subprocesses still execute with local-user permissions; no sandbox claim should be made. Windows/macOS remain unverified, and no live model-provider end-to-end scenario is claimed.
