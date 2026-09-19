# BedRock

BedRock is an early Go 1.22 prototype for local-first, owner-controlled AI software-engineering orchestration. It currently proves one narrow vertical slice: gather bounded repository context, invoke a provider-neutral command adapter, apply bounded file changes without overwriting pre-existing dirty paths, run explicit verification commands, retry with failure evidence, and roll back BedRock-authored changes when verification never succeeds.

BedRock is **not** production-ready. Provider-specific integrations, sandboxing beyond the current file/Git safeguards, broad cross-platform validation, and a polished adapter ecosystem are not implemented yet.

## Requirements

- Go 1.22+
- Git, when running against a Git repository
- A provider adapter executable that implements the JSON stdin/stdout contract described below

No external Go dependencies are currently required.

## Build and test

```sh
go build ./cmd/bedrock
go test ./...
go vet ./...
```

CI also requires the repository to be `gofmt` clean.

## Run

```sh
./bedrock run \
  --repo /path/to/repository \
  --task "Fix the failing API validation test" \
  --provider-bin /path/to/provider-adapter \
  --verify "go test ./..."
```

On Windows, invoke the built `bedrock.exe` and use a verification command valid for `cmd.exe`.

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

Explicitly passed environment values are redacted from provider failure stderr. Do not pass secrets through `--provider-arg`, task text, source files, or verification commands.

## Provider adapter contract

BedRock is provider-neutral at the core. `--provider-bin` is an executable that:

1. reads one JSON request from stdin;
2. performs provider/model interaction itself;
3. writes exactly one JSON response to stdout;
4. writes diagnostics to stderr and exits non-zero on failure.

The request includes the engineering task, attempt number, previous verification failure (when retrying), bounded repository files, and paths that were already dirty before the run. The response shape is:

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

Unknown response fields and trailing JSON/data are rejected. A response is limited to bounded file counts and sizes before writes are applied.

## Safety behavior currently implemented

- Pre-existing Git dirty paths are protected from BedRock writes.
- Absolute paths, repository traversal, `.git`, `.bedrock`, symlink targets/components, and non-regular replacement targets are rejected.
- Proposed changes are bounded by file count, per-file bytes, and total bytes.
- Verification is explicit; no verification commands means the run is reported `UNVERIFIED`, not `VERIFIED`.
- Failed final verification rolls back files changed by the run.
- Run evidence is stored outside the repository under the operating-system user cache directory.
- Provider stdout/stderr is size bounded; provider and verification commands have timeouts.

Repository content is context data, not trusted BedRock control instructions. A provider may still produce unsafe changes, so verification commands and code review remain important trust boundaries.

## Current limitations

- There is no built-in OpenAI, Anthropic, or local-model adapter yet; an external adapter executable is required.
- Provider execution is a local subprocess, not a hardened OS/container sandbox.
- Verification commands are intentionally user-supplied shell commands and therefore execute with the user's local permissions.
- No real provider end-to-end smoke scenario has been committed yet.
- Windows and macOS behavior has not been independently verified.
- Race testing and stronger Git porcelain edge-case coverage remain outstanding.

See `docs/ENGINEERING_LEDGER.md` for current verification evidence, known risks, and the next engineering target.
