# Current Engineering Handoff

## 2026-09-20 rollback concurrency hardening

### Current state

BedRock remains a Go 1.22 local-first orchestration prototype on `main`. The preceding HEAD `439fe4d43c71f6218fbbc6ff490c26a78d6ba395` had successful GitHub Actions CI (`35484046266`).

### Work completed

- Re-inspected current history, CI, `docs/ENGINEERING_LEDGER.md`, and change/rollback code before modifying the repository.
- Found a data-loss race in rollback: after BedRock wrote a file, a user or concurrent process could edit that same path while verification was running; a later failed verification would blindly restore/delete the original and destroy the concurrent edit.
- Hardened `ChangeSet` to retain the exact bytes BedRock last wrote. Rollback now restores/removes a path only when its current contents still equal BedRock's last write. If the file changed or an originally existing file was removed, rollback fails closed and preserves the newer external state.
- Added regression coverage for both an existing file and a newly created file being edited after BedRock's write.

### Tests actually executed / evidence

- GitHub Actions run `35484046266` on the preceding HEAD completed successfully.
- GitHub Actions run `35484645458` started for source commit `9c98224280194f59b811514680727893a4cd3a6c` and was still `in_progress` when inspected.
- The regression-test commit `dd372b184fd23e02f6ac5e40f35612a4b6c4cc8b` was created after that run started. No completed CI run containing both source and new tests was visible before this handoff.
- No local Go execution is claimed; this run used the connected GitHub repository API.

### Verification status

**UNVERIFIED on current HEAD** until CI executes against a commit containing both the rollback implementation and its regression tests.

### Remaining risks / next action

First inspect CI on current HEAD and repair any format, vet, unit, race, or CLI-smoke failure. Then address the adjacent pre-apply race: dirty paths are captured before provider execution, so user changes made while a provider is thinking can become newly dirty before `ChangeSet.Apply`. Re-check/merge dirty paths immediately before applying provider changes, with regression coverage proving such an edit cannot be overwritten. Provider and verification subprocesses still run with local user permissions; no sandbox-security claim should be made.
