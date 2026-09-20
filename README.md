# BedRock

BedRock is an early Go 1.22 prototype for local-first, owner-controlled AI software-engineering orchestration. It currently proves one narrow vertical slice: gather bounded repository context, invoke a provider-neutral command adapter, apply bounded file changes without overwriting pre-existing dirty paths, run explicit verification commands, retry with failure evidence, and roll back BedRock-authored changes when verification never succeeds.

BedRock is **not** production-ready. Provider-specific integrations, hardened subprocess sandboxing, broad cross-platform validation, and a polished adapter ecosystem are not implemented yet.

## One-step start

A fresh supported checkout has exactly one canonical demo/start command:

```sh
go run ./scripts/demo.go
```

Run it from the BedRock checkout root. Go 1.22+ and Git must be on `PATH`; the launcher validates both, creates a disposable owned workspace, builds the real BedRock CLI and deterministic provider, runs the complete orchestration path, waits for actual verification, and prints `READY` plus the workspace/result paths only after the result is proven. It needs no credentials or external services. Re-running it safely recreates only BedRock-owned `bin` and `repository` children inside `$BEDROCK_DEMO_DIR` (or the operating-system temporary `bedrock-demo` directory), and it refuses to reuse an existing unmarked directory.

This is the complete usable **deterministic prototype demo**, not a claim of live-model readiness. There is currently no backend server or web UI, so there are no application URLs to print. A real model still requires an external provider adapter as described below.

## Development and advanced/manual use

These commands are optional development/troubleshooting paths, not the normal demo startup sequence:

```sh
go build ./cmd/bedrock
go test ./...
go test -race ./...
go vet ./...
```

To run against a real repository/provider adapter after building `bedrock`:

```sh
./bedrock run \
  --repo /path/to/repository \
  --task "Fix the failing API validation test" \
  --provider-bin /path/to/provider-adapter \
  --verify "go test ./..."
```

Useful options:

- `--provider-arg VALUE` passes an argument to the adapter; repeat as needed.
- `--provider-env NAME` explicitly passes one environment variable to the adapter; repeat as needed.
- `--verify COMMAND` adds a verification command; repeat as needed.
- `--max-attempts N` controls implementation/repair attempts; default is 2.
- `--provider-timeout` and `--verify-timeout` bound provider and verification execution.
- `--context-files` and `--context-bytes` bound repository context sent to the provider.

### Credential handling

Provider subprocesses do **not** inherit the full parent environment. BedRock passes only a small operational allowlist by default. Credentials must be explicitly opted in, for example:

```sh
bedrock run \
  --repo . \
  --task "Fix the failing test" \
  --provider-bin ./my-openai-adapter \
  --provider-env OPENAI_API_KEY \
  --verify "go test ./..."
```

Explicitly passed environment values are redacted from provider failure stderr. Verification output stored in run evidence also redacts values from ambient environment variables whose names look credential-bearing (for example tokens, passwords, API keys, private/access keys, and credentials). This is defense in depth, not a guarantee that arbitrary secrets embedded in files, task text, command arguments, or unrelated environment names cannot appear in output. Do not pass secrets through `--provider-arg`, task text, source files, or verification commands.

## Provider adapter contract

BedRock is provider-neutral at the core. `--provider-bin` is an executable that reads one JSON request from stdin, performs provider/model interaction itself, writes exactly one JSON response to stdout, and writes diagnostics to stderr with a non-zero exit on failure.

The response shape is:

```json
{
  "summary": "what the adapter changed",
  "changes": [
    {
      "path": "relative/path/to/file.go",
      "content": "complete replacement file content"
    }
  ]
}
```

Unknown response fields and trailing JSON/data are rejected. Proposed changes are bounded before writes are applied.

## Safety behavior currently implemented

- Pre-existing Git dirty paths are protected, including source and destination paths represented by rename/copy records.
- Git dirty state is refreshed after provider execution, protecting edits made while the provider runs.
- Repair attempts retain ownership only while previously written bytes still match; intervening external edits are not overwritten.
- Rollback refuses to overwrite or recreate a BedRock-written path after an external change or removal.
- Absolute paths, traversal, `.git`, `.bedrock`, symlink targets/components, and non-regular replacement targets are rejected.
- Proposed changes are bounded by file count, per-file bytes, and total bytes.
- No verification commands means `UNVERIFIED`, never `VERIFIED`.
- Run evidence is stored outside the repository under the operating-system user cache directory.
- Provider output and execution time are bounded; credential-like values are redacted from persisted verification output.

Repository content is context data, not trusted BedRock control instructions. Provider-produced changes remain untrusted until verification/review.

## Current limitations

- There is no built-in OpenAI, Anthropic, or local-model adapter; an external adapter executable is required for live-model use.
- Provider execution is a local subprocess, not a hardened OS/container sandbox.
- Verification commands are user-supplied shell commands and execute with the user's local permissions.
- The canonical Go launcher is cross-platform by implementation and is exercised on Linux CI; Windows one-step execution is not yet independently proven in CI.
- Deterministic fake-provider end-to-end CLI scenarios are covered in CI, but no live model-provider end-to-end scenario is claimed.
- Secret redaction is heuristic.
- Concurrent-edit protection narrows overwrite races but does not provide filesystem transactions.

See `docs/ENGINEERING_LEDGER.md` for verification history and `docs/LATEST_HANDOFF.md` for the current continuation point.
