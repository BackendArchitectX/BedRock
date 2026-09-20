# Latest engineering handoff

## 2026-09-20 — one-step prerequisite failure coverage

### Current state

BedRock's canonical fresh-checkout start remains `go run ./scripts/demo.go`. Before this pass, current `main` was `94e9dde` and CI run `35502370111` completed successfully. That baseline includes Linux and Windows execution of the canonical launcher, rerun/idempotency checks, foreign-workspace refusal, checkout-root cleanup protection, Linux race testing, and removal of the obsolete shell launcher.

### Work completed

- Reconciled against current upstream `main` before editing; did not rely on the older rewritten baseline or stale handoff.
- Added an integration gate for the launcher's two unavoidable prerequisites. CI now compiles the real launcher, injects an unsupported `go1.21.13`, and proves it exits with the actionable Go 1.22+ diagnostic before creating `BEDROCK_DEMO_DIR`.
- Added the equivalent missing-Git scenario using a Go-only PATH shim. It proves the launcher reports the missing Git prerequisite and leaves the requested workspace nonexistent.
- Commit: `b46c239373b922500c6a01be0da15f32b2dc950c` (`test(demo): prove prerequisite failures are non-mutating`). GitHub attributes both author and committer to `BackendArchitectX`; no force push or history rewrite was used.

### Verification actually observed

- Baseline CI `35502370111` on `94e9dde`: **PASS**.
- At final inspection no Actions run was yet visible for `b46c239`; therefore the new prerequisite coverage is **UNVERIFIED**. Do not infer PASS from the baseline.
- No local Go execution is claimed in this pass.
- Preserve the user's Windows evidence: ordinary build/test/vet pass locally; Windows `-race` is **not executed due to ThreadSanitizer startup failure**. Linux CI remains the authoritative race-detector gate.

### Remaining risks / next action

First inspect CI for `b46c239` or its current descendant and repair any real failure before extending the launcher. If green, the one-step prototype path has positive Linux/Windows startup coverage plus negative prerequisite/workspace-safety coverage. The next engineering choice should then be driven by the actual product gap rather than adding scheduler-derived runtime behavior. Provider subprocesses and explicit verification commands still execute with local-user permissions; do not claim sandboxing.
