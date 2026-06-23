#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SHUX="${SHUX:-shux}"
OUT="$ROOT/.shux/out"
WORKSPACE="${YUKTI_WORKSPACE_DIR:-}"
SESSION="yukti-deploy-tui-$$"

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

"$SHUX" --format json session create "$SESSION" -d --title "yukti deployment tui" --cwd "$WORKSPACE" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" >/dev/null
"$SHUX" pane set-size -s "$SESSION" --cols 150 --rows 46 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Press Enter" --timeout-ms 15000 >/dev/null
"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/deployment-tui-welcome.png"

"$SHUX" pane send-keys -s "$SESSION" --data 'DQ==' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "YOUR PROJECTS" --timeout-ms 30000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Yukti Release Proof" --timeout-ms 30000 >/dev/null
"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/deployment-tui-projects.png"

"$SHUX" pane send-keys -s "$SESSION" --data 'DQ==' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "JavaScript (Server)" --timeout-ms 30000 >/dev/null
"$SHUX" pane send-keys -s "$SESSION" --text 'D' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Deployments" --timeout-ms 30000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "WEB_APP" --timeout-ms 30000 >/dev/null
"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/deployment-tui-deployments.png"

echo "$OUT/deployment-tui-welcome.png"
echo "$OUT/deployment-tui-projects.png"
echo "$OUT/deployment-tui-deployments.png"
