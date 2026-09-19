# Current engineering handoff

## 2026-09-20 — context secret-safety pass

### Current state

BedRock remains a Go 1.22 local-first orchestration prototype on `main`. The existing end-to-end CLI success and rollback scenarios are covered by CI; the latest pre-change HEAD (`ade8bdd`) completed CI successfully.

### Work completed

- Re-inspected recent commits, current CI history, README, engineering ledger, context selection, verification handling, and TODO/FIXME search rather than assuming earlier scheduled work happened.
- Identified a high-priority trust-boundary issue: bounded context selection could read common credential files such as `.env`, private keys, service-account JSON, and SSH private keys and send their contents to the configured model/provider adapter.
- Added conservative context exclusions for `.env` variants, common credential/config files, private-key/keystore extensions, and `.ssh` directories while deliberately not blocking ordinary source names such as `credentials.go` or `key.go`.
- Added deterministic tests proving representative secret-like files are excluded while normal source/configuration files remain available.

### Verification actually observed

- GitHub Actions run `35476232374` for pre-change commit `ade8bdd3512656f7733e9bf54ea95b675e0b1839`: **PASS**.
- GitHub Actions run `35477044394` for test commit `6fe3099dfccf03ac48e372ff5cb19734ed77e6be` was **IN_PROGRESS** when this handoff was written. Do not treat the new context filter as verified until that run (or a later run containing it) completes successfully.
- No local Go execution is claimed; this run operated through the connected GitHub repository/API.

### Risks / next action

1. First inspect CI on current HEAD and repair any format/vet/test/race/CLI-smoke failure before adding features.
2. If green, independently challenge the secret-file policy for false negatives and false positives. Consider whether Git-ignored-file exclusion is justified, but do not blindly hide all ignored files because generated/ignored source may still be legitimate engineering context.
3. Update README safety documentation once the exclusion behavior is verified.
4. Remaining larger risks include unsandboxed provider subprocesses and intentionally user-supplied shell verification commands; do not claim OS-level sandboxing.