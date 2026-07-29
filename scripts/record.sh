#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: $0 [basename] [runs]"
  echo "Records screen recordings of 'go run .' through predefined sizes."
  echo
  echo "Sizes: 1920x1080, 1920x1080 (path overlay), 1280x720, 800x600"
  echo "Defaults: basename=demo, runs=3"
  echo
  echo "Output: <basename>_<width>x<height>[_<suffix>]_<run>.gif"
  echo
  echo "Examples:"
  echo "  $0"
  echo "  $0 myapp"
  echo "  $0 demo 5"
  exit 1
}

if [[ $# -gt 2 ]]; then
  usage
fi

BASE="${1:-demo}"
RUNS="${2:-3}"

if ! [[ "$RUNS" =~ ^[0-9]+$ ]] || [ "$RUNS" -lt 1 ]; then
  echo "Error: runs must be a positive integer"
  exit 1
fi

SIZES=(
  "1920 1080"
  "1920 1080 path"
  "1280 720"
  # "960 540"
  "800 600"
)

if ! command -v vhs &>/dev/null; then
  echo "Error: vhs is not installed. Install it with: brew install vhs"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

for SIZE in "${SIZES[@]}"; do
  read -r WIDTH HEIGHT EXTRA <<< "$SIZE"
  for RUN in $(seq 1 "$RUNS"); do
    SUFFIX="${EXTRA:+_${EXTRA}}"
    OUTPUT="${BASE}_${WIDTH}x${HEIGHT}${SUFFIX}_${RUN}.gif"
    TAPE_FILE="$(mktemp "$TEMP_DIR/hazard-record-XXXXXX")"

    PATH_CMD=""
    if [[ "$EXTRA" == "path" ]]; then
      PATH_CMD=$'Sleep 1s\nType "p"'
    fi

    cat > "$TAPE_FILE" <<EOF
Output ${OUTPUT}

Set Width ${WIDTH}
Set Height ${HEIGHT}
Set FontSize 17
Set Theme { "background": "#212734" }
Set CursorBlink false

Hide
Type "go run ."
Enter
Sleep 2s
Show
${PATH_CMD}
Sleep 60s
EOF

    echo "Recording (${RUN}/${RUNS}): ${WIDTH}x${HEIGHT} -> ${OUTPUT}"
    cd "$REPO_ROOT"
    vhs "$TAPE_FILE"
  done
done

echo "All recordings complete."
