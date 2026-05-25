#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SHUX="${SHUX:-/Users/indrasvat/.local/bin/shux}"
OUT="$ROOT/.shux/out"
WORKSPACE="$OUT/workspace-sync-fixture"
SESSION="yukti-sync-smoke-$$"

mkdir -p "$OUT"
rm -rf "$WORKSPACE"
mkdir -p "$WORKSPACE"

cat >"$WORKSPACE/yukti.json" <<'JSON'
{
  "version": 1,
  "script_id": "script-smoke",
  "title": "Smoke Sync Project",
  "last_remote_hash": "remote-hash",
  "last_pulled_at": "2026-05-25T00:00:00Z",
  "files": {}
}
JSON

cat >"$WORKSPACE/appsscript.json" <<'JSON'
{
  "timeZone": "America/Los_Angeles",
  "exceptionLogging": "STACKDRIVER",
  "runtimeVersion": "V8"
}
JSON

cat >"$WORKSPACE/Code.gs" <<'JS'
function main() {
  console.log("shux workspace sync smoke");
}
JS

cd "$ROOT"
make build >/dev/null

cleanup() {
  "$SHUX" session kill "$SESSION" >/dev/null 2>&1 || true
}
trap cleanup EXIT

"$SHUX" --format json session create "$SESSION" -d --title "yukti workspace sync" --cwd "$WORKSPACE" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" >/dev/null

"$SHUX" pane set-size -s "$SESSION" --cols 132 --rows 42 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Workspace" --timeout-ms 10000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Smoke Sync Project" --timeout-ms 10000 >/dev/null

"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/workspace-sync-welcome.png"

"$SHUX" session kill "$SESSION" >/dev/null

"$SHUX" --format json session create "$SESSION" -d --title "yukti diff" --cwd "$WORKSPACE" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" diff >/dev/null

"$SHUX" pane set-size -s "$SESSION" --cols 120 --rows 28 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Local changes" --timeout-ms 10000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "2 added" --timeout-ms 10000 >/dev/null

"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/workspace-sync-diff.png"

echo "$OUT/workspace-sync-welcome.png"
echo "$OUT/workspace-sync-diff.png"
