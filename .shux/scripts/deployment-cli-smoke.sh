#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SHUX="${SHUX:-shux}"
OUT="$ROOT/.shux/out"
WORKSPACE="${YUKTI_WORKSPACE_DIR:-}"
SESSION="yukti-deploy-cli-$$"

if [[ -z "$WORKSPACE" ]]; then
  echo "YUKTI_WORKSPACE_DIR is required" >&2
  exit 2
fi

mkdir -p "$OUT"

cleanup() {
  "$SHUX" session kill "$SESSION" >/dev/null 2>&1 || true
}
trap cleanup EXIT

cd "$ROOT"
make build >/dev/null

"$SHUX" --format json session create "$SESSION" -d --title "yukti deployment cli" --cwd "$WORKSPACE" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" deployments list >/dev/null
"$SHUX" pane set-size -s "$SESSION" --cols 132 --rows 34 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Deployments" --timeout-ms 15000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "WEB_APP" --timeout-ms 15000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Deployment  AK" --timeout-ms 15000 >/dev/null
"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/deployment-cli-list.png"

"$SHUX" session kill "$SESSION" >/dev/null

"$SHUX" --format json session create "$SESSION" -d --title "yukti versions cli" --cwd "$WORKSPACE" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" versions list >/dev/null
"$SHUX" pane set-size -s "$SESSION" --cols 120 --rows 30 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Versions" --timeout-ms 15000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Version  v" --timeout-ms 15000 >/dev/null
"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/deployment-cli-versions.png"

echo "$OUT/deployment-cli-list.png"
echo "$OUT/deployment-cli-versions.png"
