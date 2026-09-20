# Latest engineering handoff

## 2026-09-20 — verification-aware provider context

### Current state

Reconciled against `origin/main` through product/test commit `4de07761`. The canonical fresh-checkout path remains `go run ./scripts/demo.go`; scheduler staggering is external only and has no product/runtime meaning.

### Work verified before this change

- CI run `35527173288` for `26138ea7` completed successfully. That baseline exercised Linux formatting, vet, tests, authoritative race checks, canonical one-step startup/rollback, Windows launcher acceptance, and an actual CLI run through the built-in OpenAI-compatible HTTP provider fixture.
- The built-in HTTP provider is therefore no longer only unit-tested wiring: the CLI fixture verifies authorization/model request behavior and a provider-produced repository edit through verification.
- Preserve the user's Windows evidence: ordinary Windows build/test/vet are valid; Windows race is not executed because ThreadSanitizer could not initialize. Linux CI remains the authoritative race gate.

### Change in this pass

- `ProviderRequest` now carries `verificationCommands` when the configured verifier is `ShellVerifier`, so a real model knows the acceptance commands before its first edit instead of discovering them only after a failed attempt.
- The plan is copied from verifier configuration and redacted with the same sensitive-environment-value mechanism used for verification evidence before it crosses the provider boundary.
- The built-in HTTP-provider CLI integration test now decodes the actual provider context and requires the verification plan to arrive while an API-key value embedded in the test command is absent and replaced by `[REDACTED]`.
- CI for the newest product/test commit was still running when this handoff was written; do not call `4de07761` verified until its exact-head workflow succeeds.

### Product usefulness assessment / next action

BedRock can now gather bounded repository context, call either a command provider or built-in OpenAI-compatible HTTP provider, apply bounded protected changes, run configured verification, feed verification failure/output into a repair attempt, rollback on terminal failure, and persist run evidence. The verification-aware first request materially improves real-model first-attempt quality without importing external scheduler semantics.

The highest-value remaining functional gap is change expressiveness: provider changes currently only write/replace regular files. A software-engineering run cannot intentionally delete an obsolete tracked file. Next implement an explicit, bounded delete operation with the same dirty-path, symlink/path-containment, concurrent-modification, rollback, repair-attempt, evidence, and HTTP-provider integration guarantees as writes. Do not encode deletion as magic empty content.

Provider subprocesses and verification commands still execute with local-user permissions; do not claim sandboxing.
