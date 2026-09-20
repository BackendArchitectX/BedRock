# Latest engineering handoff

## 2026-09-20 — nested-root dirty-work protection correction

### Current state

Before this pass, current `main` (`33cc7d58e1835107561f6ee062fa8828921ca11b`) had successful CI run `35481923304`, including format, vet, unit, race, successful CLI smoke, and failed-verification rollback smoke.

### Work completed

- Re-inspected current commits, CI, `changes.go`, tests, and `docs/ENGINEERING_LEDGER.md` rather than trusting the preceding handoff.
- Challenged the preceding nested-root test repair and found that it had encoded the wrong coordinate system as expected behavior. `DirtyPaths(nested)` returned `nested/owned.txt` relative to the enclosing Git worktree, while `ChangeSet.Apply(nested, ...)` compares provider paths relative to the BedRock run root (`owned.txt`). The two names therefore never matched, so a provider could overwrite dirty user work when BedRock was invoked from a subdirectory of a larger worktree.
- Fixed `DirtyPaths` to ask Git for `--relative` status scoped to `.` so protected paths use the same run-root-relative coordinates as provider changes.
- Strengthened the regression test beyond checking a map key: it now calls `ChangeSet.Apply` against the dirty nested file, requires the overwrite to be rejected, and verifies the original bytes remain unchanged.

### Commits

- `5310d117634eb897c6c19860e30df5a056774df6` — `fix(git): align nested dirty paths with run root`
- `f0ff2b13494911036d3addf699a0fcacacf5ac22` — `test(git): prove nested dirty overwrite rejection`

### Verification status

- No local Go execution is claimed because this run used the connected GitHub repository API rather than a mounted checkout.
- CI run `35482552410` for the regression-test HEAD was **in progress** when last inspected. The new correction is therefore **UNVERIFIED** until that run completes successfully.

### Next highest-value action

Inspect CI run `35482552410` first. If green, add an engine-level nested-root regression using a scripted provider to prove the complete `Engine.Run` path refuses to overwrite dirty work from a nested invocation. If CI fails, repair the actual failure before adding capabilities. Do not add larger orchestration abstractions while this safety boundary is still being proven.