# Verification handoff

## Current state

The last fully verified predecessor is `5fb31af47f246edb6ae0f22e111971fd73aea8f7`. GitHub Actions run `35485063316` completed successfully: format, vet, unit tests, race tests, successful CLI smoke, and failure/rollback CLI smoke all passed.

## Work in this pass

Independent review found misleading rollback evidence in `Engine.Run`: every rollback path set `Evidence.RolledBack = true` even when `ChangeSet.Rollback` returned a conflict, and snapshot/provider-error rollback paths discarded rollback errors entirely. This matters because rollback deliberately refuses to overwrite a file changed concurrently after BedRock wrote it.

- `bf69c3a361f75a75539449254576e71d1395849a` centralizes rollback handling, propagates rollback conflicts, and records `RolledBack=true` only after rollback succeeds.
- `461c2854552fb28dc0f8e550799ebcbe351748ac` adds an independent engine-level regression: the verifier simulates a concurrent user edit before failing; BedRock must preserve that edit, return an error containing the rollback conflict, and must not claim successful rollback.

## Verification status

The source-only CI run for `bf69c3a361f75a75539449254576e71d1395849a` was still in progress when inspected. CI run `35485666786` for regression-test HEAD `461c2854552fb28dc0f8e550799ebcbe351748ac` was queued. Therefore the new source and regression test are **UNVERIFIED** until CI completes on a commit containing both. No local Go execution is claimed in this pass.

## Next action

Inspect CI for current HEAD first. Require format, vet, unit, race, successful CLI smoke, and failure/rollback CLI smoke to pass before carrying forward a VERIFIED claim. If green, challenge rollback behavior when an earlier verification attempt modified files and a later provider or context step fails; the shared rollback helper is intended to cover those paths but they do not yet have an engine-level conflict regression.

Provider subprocesses and verification shell commands still execute with local user permissions; BedRock must not be described as sandboxed.
