#!/usr/bin/env bash
# tui-screenshot.sh — Capture docker-tui TUI screenshot via tmux.
#
# Launches docker-tui inside a tmux session, waits for the TUI to render,
# then dumps the visible pane contents to a plain-text file. ANSI color
# escapes are preserved by default; pass --strip-ansi to remove them.
#
# Requirements:
#   - tmux (>= 2.5)            — sudo apt install tmux
#   - ./dist/docker-tui binary  — run `just build` first
#
# Usage:
#   scripts/tui-screenshot.sh [OPTIONS] [OUTPUT_FILE]
#
# Options:
#   -w, --width   N    Pane width  (default: 140)
#   -h, --height  N    Pane height (default: 40)
#   -b, --binary  PATH Path to docker-tui binary (default: ./dist/docker-tui)
#       --strip-ansi     Strip ANSI color escapes from output
#       --wait    SEC    Seconds to wait before capture (default: 2)
#       --help           Show this help and exit
#
# Examples:
#   scripts/tui-screenshot.sh /tmp/dtui.txt
#   scripts/tui-screenshot.sh --width 160 --height 50 --strip-ansi shot.txt
#   scripts/tui-screenshot.sh --wait 5 /tmp/dtui-with-error.txt

set -euo pipefail

cd "$(dirname "$0")/.."

# Defaults
OUTPUT="/tmp/dtui-screenshot.txt"
WIDTH=140
HEIGHT=40
BINARY="./dist/docker-tui"
STRIP_ANSI=0
WAIT_SEC=2

usage() {
    sed -n '2,30p' "$0" | sed 's/^# \{0,1\}//'
    exit "${1:-0}"
}

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        -w|--width)  WIDTH="$2"; shift 2 ;;
        -h|--height) HEIGHT="$2"; shift 2 ;;
        -b|--binary) BINARY="$2"; shift 2 ;;
        --strip-ansi) STRIP_ANSI=1; shift ;;
        --wait)      WAIT_SEC="$2"; shift 2 ;;
        --help)      usage 0 ;;
        -*)          echo "unknown flag: $1" >&2; usage 1 ;;
        *)           OUTPUT="$1"; shift ;;
    esac
done

# Preconditions
if ! command -v tmux >/dev/null 2>&1; then
    echo "ERROR: tmux not found. Install with: sudo apt install tmux" >&2
    exit 1
fi

if [[ ! -x "$BINARY" ]]; then
    echo "ERROR: binary not found or not executable: $BINARY" >&2
    echo "Build it first: just build" >&2
    exit 1
fi

SESSION="dtui-shot-$$"

cleanup() {
    if tmux has-session -t "$SESSION" 2>/dev/null; then
        tmux kill-session -t "$SESSION" 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo "==> launching tmux session: $SESSION (${WIDTH}x${HEIGHT})"
tmux new-session -d -s "$SESSION" -x "$WIDTH" -y "$HEIGHT" "$BINARY"

# Wait for the TUI to render its first frame
echo "==> waiting ${WAIT_SEC}s for TUI to render"
sleep "$WAIT_SEC"

echo "==> capturing pane to $OUTPUT"
tmux capture-pane -t "$SESSION" -p -S -200 > "$OUTPUT"

if [[ "$STRIP_ANSI" -eq 1 ]]; then
    # Remove CSI sequences (\x1b[...letter) while preserving visible text
    sed -i -E 's/\x1B\[[0-9;]*[a-zA-Z]//g; s/\x1B\][^\x07]*\x07//g' "$OUTPUT"
    echo "==> stripped ANSI escape sequences"
fi

LINES=$(wc -l < "$OUTPUT")
BYTES=$(wc -c < "$OUTPUT")
echo "==> done: $OUTPUT ($LINES lines, $BYTES bytes)"