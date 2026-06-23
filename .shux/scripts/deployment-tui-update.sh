#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SHUX="${SHUX:-shux}"
OUT="$ROOT/.shux/out"
WORKSPACE="${YUKTI_WORKSPACE_DIR:-}"
SESSION="yukti-deploy-tui-update-$$"
VERIFY_SESSION="yukti-deploy-tui-update-verify-$$"

if [[ -z "$WORKSPACE" ]]; then
  echo "YUKTI_WORKSPACE_DIR is required" >&2
  exit 2
fi

mkdir -p "$OUT"

cleanup() {
  "$SHUX" session kill "$SESSION" >/dev/null 2>&1 || true
  "$SHUX" session kill "$VERIFY_SESSION" >/dev/null 2>&1 || true
}
trap cleanup EXIT

cd "$ROOT"
make build >/dev/null
EXPECTED_VERSION="$("$ROOT/bin/yukti" versions list --dir "$WORKSPACE" \
  | sed -nE 's/.*Version[^v]*v([0-9]+).*/\1/p' \
  | sort -n \
  | tail -1)"
EXPECTED_VERSION="${EXPECTED_VERSION:-0}"
EXPECTED_VERSION="$((EXPECTED_VERSION + 1))"

"$SHUX" --format json session create "$SESSION" -d --title "yukti deployment tui update" --cwd "$WORKSPACE" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" >/dev/null
"$SHUX" pane set-size -s "$SESSION" --cols 150 --rows 46 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Press Enter" --timeout-ms 15000 >/dev/null
"$SHUX" pane send-keys -s "$SESSION" --data 'DQ==' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "YOUR PROJECTS" --timeout-ms 30000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Yukti Release Proof" --timeout-ms 30000 >/dev/null
"$SHUX" pane send-keys -s "$SESSION" --data 'DQ==' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "JavaScript (Server)" --timeout-ms 30000 >/dev/null
"$SHUX" pane send-keys -s "$SESSION" --text 'D' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Deployments" --timeout-ms 30000 >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "WEB_APP" --timeout-ms 30000 >/dev/null

# Move from the automatic HEAD deployment to the first versioned deployment,
# then confirm an update. This verifies stable-URL TUI deployment updates.
"$SHUX" pane send-keys -s "$SESSION" --text 'j' >/dev/null
"$SHUX" pane send-keys -s "$SESSION" --text 'u' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Press u again" --timeout-ms 15000 >/dev/null
"$SHUX" --format json pane snapshot -s "$SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/deployment-tui-update-confirm.png"
"$SHUX" pane send-keys -s "$SESSION" --text 'u' >/dev/null
"$SHUX" pane wait-for -s "$SESSION" --text "Moved" --timeout-ms 60000 >/dev/null

# Relaunch before the final screenshot so the proof is not satisfied by stale
# screen scrollback or Apps Script's eventually consistent first refresh.
"$SHUX" session kill "$SESSION" >/dev/null 2>&1 || true
"$SHUX" --format json session create "$VERIFY_SESSION" -d --title "yukti deployment tui update verify" --cwd "$WORKSPACE" -- \
  env TERM=xterm-256color "$ROOT/bin/yukti" >/dev/null
"$SHUX" pane set-size -s "$VERIFY_SESSION" --cols 150 --rows 46 >/dev/null
"$SHUX" pane wait-for -s "$VERIFY_SESSION" --text "Press Enter" --timeout-ms 15000 >/dev/null
"$SHUX" pane send-keys -s "$VERIFY_SESSION" --data 'DQ==' >/dev/null
"$SHUX" pane wait-for -s "$VERIFY_SESSION" --text "YOUR PROJECTS" --timeout-ms 30000 >/dev/null
"$SHUX" pane wait-for -s "$VERIFY_SESSION" --text "Yukti Release Proof" --timeout-ms 30000 >/dev/null
"$SHUX" pane send-keys -s "$VERIFY_SESSION" --data 'DQ==' >/dev/null
"$SHUX" pane wait-for -s "$VERIFY_SESSION" --text "JavaScript (Server)" --timeout-ms 30000 >/dev/null
"$SHUX" pane send-keys -s "$VERIFY_SESSION" --text 'D' >/dev/null
"$SHUX" pane wait-for -s "$VERIFY_SESSION" --text "WEB_APP" --timeout-ms 30000 >/dev/null
found_version=0
for _ in {1..12}; do
  if "$SHUX" pane capture -s "$VERIFY_SESSION" | grep -Fq "v$EXPECTED_VERSION"; then
    found_version=1
    break
  fi
  "$SHUX" pane send-keys -s "$VERIFY_SESSION" --text 'r' >/dev/null
  sleep 5
done
if [[ "$found_version" != "1" ]]; then
  echo "deployment version v$EXPECTED_VERSION did not appear in final TUI screen" >&2
  "$SHUX" pane capture -s "$VERIFY_SESSION" >&2 || true
  exit 1
fi
"$SHUX" --format json pane snapshot -s "$VERIFY_SESSION" \
  | jq -r .png_base64 | base64 -d > "$OUT/deployment-tui-update.png"

echo "$OUT/deployment-tui-update-confirm.png"
echo "$OUT/deployment-tui-update.png"
