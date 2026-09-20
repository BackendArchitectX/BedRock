# Latest engineering handoff

## 2026-09-20 — protect checkout from launcher workspace mutation

### Current state

BedRock's canonical fresh-checkout start remains `go run ./scripts/demo.go`. This pass reconciled against current upstream `main` at `673fa3d`, whose CI and dedicated demo-path-safety workflow both completed successfully. That baseline already rejects direct and ancestor symlinks for `BEDROCK_DEMO_DIR`.

### Work completed

- Found an uncovered overlap direction: the launcher rejected a workspace that contained the checkout, but still allowed `BEDROCK_DEMO_DIR` itself to be inside the checkout. A fresh in-checkout path could therefore create launcher-owned files inside the user's source tree.
- `7576403` (`fix(demo): reject checkout-contained workspace`) now rejects a workspace equal to or beneath the checkout before marker creation or cleanup. The existing inverse guard still rejects a workspace that contains the checkout.
- `3883822` (`test(demo): protect checkout from workspace mutation`) adds an integration assertion that an in-checkout workspace is rejected, remains nonexistent, and emits the intended diagnostic.
- No force push or history rewrite was used.

### Verification actually observed

- Baseline `673fa3d`: CI run `35507972281` and Demo path safety run `35507972295` both **PASS**.
- CI for `7576403` had started when inspected; current descendant `3883822` had not yet produced a completed run. The new protection is therefore **UNVERIFIED on current HEAD** and must not be reported as passing yet.
- No local Go execution is claimed in this pass.
- Preserve the user's Windows evidence: ordinary build/test/vet pass locally; Windows `-race` is **not executed due to ThreadSanitizer startup failure**. Linux CI remains the authoritative race-detector gate.

### Remaining risks / next action

First inspect CI and Demo path safety for `3883822` or its current descendant and repair any real failure. If green, continue from the actual latest `origin/main`; do not add scheduler-derived runtime behavior. Provider subprocesses and explicit verification commands still execute with local-user permissions, so BedRock must not claim sandboxing.
