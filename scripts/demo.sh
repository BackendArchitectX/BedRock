#!/bin/sh
set -eu

fail() {
  printf 'BedRock demo: %s\n' "$*" >&2
  exit 1
}

command -v go >/dev/null 2>&1 || fail 'Go 1.22+ is required and was not found on PATH.'
command -v git >/dev/null 2>&1 || fail 'Git is required and was not found on PATH.'

version="$(go env GOVERSION 2>/dev/null || true)"
case "$version" in
  go1.22*|go1.2[3-9]*|go1.[3-9][0-9]*|go[2-9].*) ;;
  *) fail "Go 1.22+ is required; found ${version:-unknown}." ;;
esac

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
work="${BEDROCK_DEMO_DIR:-${TMPDIR:-/tmp}/bedrock-demo}"
marker="$work/.bedrock-demo-owned"
bin="$work/bin"
repo="$work/repository"

# Never recursively clean a caller-selected directory unless a previous BedRock
# demo run marked it as owned. This keeps reruns deterministic without turning
# BEDROCK_DEMO_DIR into an arbitrary-directory deletion primitive.
if [ -e "$work" ] && [ ! -f "$marker" ]; then
  fail "refusing to reuse $work because it is not marked as a BedRock demo directory; choose an empty BEDROCK_DEMO_DIR or remove it yourself."
fi
mkdir -p "$work"
: > "$marker"
rm -rf "$bin" "$repo"
mkdir -p "$bin" "$repo"
git -C "$repo" init -q

printf 'BedRock demo: building CLI and deterministic provider...\n'
go build -o "$bin/bedrock" "$root/cmd/bedrock"
go build -o "$bin/fake-provider" "$root/cmd/bedrock/testdata/fakeprovider"

printf 'BedRock demo: starting verified orchestration run...\n'
"$bin/bedrock" run \
  --repo "$repo" \
  --task 'write deterministic result' \
  --provider-bin "$bin/fake-provider" \
  --verify 'test "$(cat result.txt)" = good'

test "$(cat "$repo/result.txt")" = good || fail 'verification output did not match the expected result.'
printf 'BedRock demo: READY\n'
printf 'Workspace: %s\n' "$repo"
printf 'Result: %s\n' "$repo/result.txt"
