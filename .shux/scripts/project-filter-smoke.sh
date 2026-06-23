#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SHUX="${SHUX:-shux}"
OUT="$ROOT/.shux/out"
SESSION="yukti-project-filter-$$"

mkdir -p "$OUT"

cleanup() {
  "$SHUX" session kill "$SESSION" >/dev/null 2>&1 || true
}
trap cleanup EXIT

cd "$ROOT"
make build >/dev/null

"$SHUX" --format json session create "$SESSION" -d --title "yukti project filter" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" >/dev/null
"$SHUX" pane set-size -s "$SESSION" --cols 150 --rows 46 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Press Enter" --timeout-ms 15000 >/dev/null
"$SHUX" pane send-keys -s "$SESSION" --data 'DQ==' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "YOUR PROJECTS" --timeout-ms 30000 >/dev/null
"$SHUX" pane send-keys -s "$SESSION" --text '/' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "type filter" --timeout-ms 15000 >/dev/null
"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/project-filter-mode.png"

echo "$OUT/project-filter-mode.png"
