# Verification handoff

## 2026-09-20 nested Git context verification pass

### Repository reality inspected

- Branch: `main`.
- Pre-pass HEAD: `e52f1b4d79df216bbcdc831d407cd6062c8bf2a5`.
- Recent commits, `docs/ENGINEERING_LEDGER.md`, context implementation/tests, and GitHub Actions state were inspected before changes.
- The immediately preceding Git-ignore regression run `35479867536` on commit `fff324804f4732e54f64b3aeac282274adf156b5` completed successfully.

### Defect found and fixed

The Git-aware context filter only treated a directory as a Git worktree when `<root>/.git` existed. A valid CLI invocation such as `bedrock run --repo ./service` from a subdirectory of a larger Git repository has no `service/.git`, so BedRock fell back to raw filesystem scanning and could send parent-`.gitignore`-excluded files to the provider.

The context detector now asks Git whether the selected root is inside a worktree and uses `git ls-files -co --exclude-standard -z` there. This also naturally supports linked-worktree `.git` files. A deterministic regression test initializes a repository, selects a nested `service` directory as the BedRock root, and proves an ignored `*.private` file is excluded while ordinary nested files remain available.

Commits:

- `fb4e0ec67c37c3b6cc7ed90781b4544e8a5afcd5` initial nested-worktree fix; this intermediate commit missed the `errors` import and is not considered verified.
- `2fbee588ea6db4f61851f45deadc09d7bf39b60b` restores compilation by adding the required import.
- `1d7a2fb76fa7b75f0d0167d1e91f33bd18484678` adds the nested-root regression test.

### Tests actually executed / evidence

- Independently reproduced Git path semantics in an isolated temporary repository: from a nested `service` directory, `git ls-files -co --exclude-standard -z` returned `config/example` and `main.go` while excluding `local.private` matched by the parent `.gitignore`.
- GitHub Actions run `35480595297` started on regression-test HEAD `1d7a2fb76fa7b75f0d0167d1e91f33bd18484678`. At last inspection it was still `in_progress` during setup, so Format, Vet, Test, Race test, CLI smoke, and rollback smoke are **not yet claimed as passing for this HEAD**.

### Remaining risks / next action

First inspect run `35480595297` and repair any actual failure before adding features. The Git probe currently treats a normal non-zero `git rev-parse` exit as a non-Git directory; a later hardening pass should consider fail-closed behavior when Git metadata is visibly present but Git itself rejects/cannot inspect the worktree. Provider and verifier subprocesses still execute with local user permissions, and Windows/macOS remain unverified.