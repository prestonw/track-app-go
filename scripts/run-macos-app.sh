#!/usr/bin/env bash
# Quit any running copy, then open the .app built in this repo (not /Applications).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BUNDLE="$ROOT/Track App.app"

if [[ ! -d "$BUNDLE" ]]; then
  echo "Missing $BUNDLE — run ./scripts/build-macos-app.sh first" >&2
  exit 1
fi

echo "Quitting any running Track App…"
killall trackapp 2>/dev/null || true
sleep 0.4

echo "Opening: $BUNDLE"
open "$BUNDLE"