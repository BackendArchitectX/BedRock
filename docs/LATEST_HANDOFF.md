# Latest engineering handoff

## 2026-09-20 — verification-failure recovery proven

### Current state

Reconciled against current `origin/main` at `217ae2cc91378dc2af2e128b1381b47afc85857a`. The canonical fresh-checkout path remains `go run ./scripts/demo.go`; scheduler staggering is external only and has no product/runtime meaning.

### Work verified

- CI run `35520365385` for exact HEAD `217ae2cc` completed successfully.
- The new verification-stage recovery regression executes a real canonical launch with injected verification failure after provider output, requires the launch to fail without `READY`, requires generated `result.txt` to be absent after rollback, preserves unrelated caller state, then reruns the canonical launcher and requires `status: VERIFIED`, `READY`, and a fresh `result.txt = good`.
- This closes the prior ledger's verification-stage recovery gap. Provider-failure recovery and repository-init failure recovery were already covered on preceding green descendants.
- Preserve the user's Windows evidence: ordinary Windows build/test/vet are valid; Windows race is not executed due to the independently reproduced ThreadSanitizer startup failure. Linux CI remains the authoritative race gate.

### Challenge / next action

Do not spend the next pass adding more launcher path denylist cases unless a concrete defect is found. The deterministic one-step prototype is now well covered across startup, rerun, path ownership, init failure, provider failure, verification failure, rollback, and Windows execution.

The highest-value product gap is functional: BedRock still requires an external provider executable for a real model. Next evaluate and implement the smallest built-in self-hostable provider path, preferably an OpenAI-compatible HTTP adapter usable with local servers such as Ollama/vLLM-compatible endpoints, without coupling the orchestration core to one vendor. Keep the deterministic zero-secret launcher as the default CI acceptance path. Any live-provider mode must validate endpoint/model configuration explicitly, never fabricate secrets, bound HTTP execution/output, and must not claim readiness until the configured provider is actually reachable.

Provider subprocesses and verification commands still execute with local-user permissions; do not claim sandboxing.
