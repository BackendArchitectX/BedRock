# BedRock

BedRock is an early Go 1.22 prototype for local-first, owner-controlled AI software-engineering orchestration. It currently proves one narrow vertical slice: gather bounded repository context, invoke a provider, apply bounded file changes without overwriting pre-existing dirty paths, run explicit verification commands, retry with failure evidence, and roll back BedRock-authored changes when verification never succeeds.

BedRock is **not** production-ready. Hardened subprocess sandboxing, broad cross-platform validation, and a polished adapter ecosystem are not implemented yet.

## One-step start

A fresh supported checkout has exactly one canonical demo/start command:

```sh
go run ./scripts/demo.go
```

Run it from the BedRock checkout root. Go 1.22+ and Git must be on `PATH`; the launcher validates both (including the minimum Go version), creates a disposable owned workspace, builds the real BedRock CLI and deterministic provider, runs the complete orchestration path, waits for actual verification, and prints `READY` plus the workspace/result paths only after the result is proven. It needs no credentials or external services. Re-running it safely recreates only BedRock-owned `bin` and `repository` children inside `$BEDROCK_DEMO_DIR` (or the operating-system temporary `bedrock-demo` directory), and it refuses to reuse an existing unmarked directory.

This is the complete usable **deterministic prototype demo**, not a claim of live-model readiness. There is currently no backend server or web UI, so there are no application URLs to print. Live-model use is an optional advanced path and does not change the zero-secret canonical demo.

## Development and advanced/manual use

These commands are optional development/troubleshooting paths, not the normal demo startup sequence:

```sh
go build ./cmd/bedrock
go test ./...
go test -race ./...
go vet ./...
```

BedRock supports either an external provider executable or a built-in OpenAI-compatible HTTP provider.

External adapter:

```sh
./bedrock run \
  --repo /path/to/repository \
  --task "Fix the failing API validation test" \
  --provider-bin /path/to/provider-adapter \
  --verify "go test ./..."
```

OpenAI-compatible HTTP endpoint (including self-hosted/local endpoints):

```sh
./bedrock run \
  --repo /path/to/repository \
  --task "Fix the failing API validation test" \
  --provider-endpoint http://127.0.0.1:11434/v1/chat/completions \
  --provider-model local-model \
  --verify "go test ./..."
```

For an endpoint that requires authentication, put the key in an environment variable and name that variable with `--provider-api-key-env`; do not put the secret itself on the command line.

Useful options:

- `--provider-arg VALUE` passes an argument to an external adapter; repeat as needed.
- `--provider-env NAME` explicitly passes one environment variable to an external adapter; repeat as needed.
- `--provider-endpoint URL` selects the built-in OpenAI-compatible HTTP provider and is mutually exclusive with `--provider-bin`.
- `--provider-model MODEL` is required with `--provider-endpoint`.
- `--provider-api-key-env NAME` optionally reads the HTTP provider API key from the named environment variable.
- `--verify COMMAND` adds a verification command; repeat as needed.
- `--max-attempts N` controls implementation/repair attempts; default is 2.
- `--provider-timeout` and `--verify-timeout` bound provider and verification execution.
- `--context-files` and `--context-bytes` bound repository context sent to the provider.

### Credential handling

Provider subprocesses do **not** inherit the full parent environment. BedRock passes only a small operational allowlist by default. Credentials for external adapters must be explicitly opted in with `--provider-env`. For the built-in HTTP provider, `--provider-api-key-env NAME` reads the credential from the named environment variable and sends it as a bearer token; endpoint URLs containing credentials are rejected. HTTP error diagnostics redact the configured API key.

Verification output stored in run evidence also redacts values from ambient environment variables whose names look credential-bearing (for example tokens, passwords, API keys, private/access keys, and credentials). This is defense in depth, not a guarantee that arbitrary secrets embedded in files, task text, command arguments, or unrelated environment names cannot appear in output. Do not pass secrets through `--provider-arg`, task text, source files, verification commands, or endpoint URLs.

## Provider behavior

BedRock remains provider-neutral at the orchestration layer. An external `--provider-bin` adapter reads one JSON request from stdin and writes exactly one JSON response to stdout. The built-in `--provider-endpoint` adapter speaks the OpenAI-compatible chat-completions wire contract and expects the assistant content to contain the same BedRock response object.

The BedRock response shape is:

```json
{
  "summary": "what the provider changed",
  "changes": [
    {
      "path": "relative/path/to/file.go",
      "content": "complete replacement file content"
    }
  ]
}
```

Unknown BedRock response fields and trailing JSON/data are rejected. HTTP response bodies are bounded, HTTP envelope trailing data is rejected, request duration is bounded, caller cancellation remains distinguishable from adapter timeout, and configured API credentials are redacted from HTTP error diagnostics. Proposed changes are bounded before writes are applied.

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

- The built-in HTTP provider implements a narrow OpenAI-compatible chat-completions contract; compatibility with a specific live endpoint/model is not implied until exercised against it.
- Provider execution is not a hardened OS/container sandbox.
- Verification commands are user-supplied shell commands and execute with the user's local permissions.
- The canonical Go launcher is exercised independently on Linux and Windows CI, including rerun/idempotency and foreign-workspace refusal checks.
- Deterministic fake-provider end-to-end CLI scenarios are covered in CI, but no live model-provider end-to-end scenario is claimed.
- Secret redaction is heuristic.
- Concurrent-edit protection narrows overwrite races but does not provide filesystem transactions.

See `docs/ENGINEERING_LEDGER.md` for verification history and `docs/LATEST_HANDOFF.md` for the current continuation point.
