#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="Track App"
BUNDLE="$ROOT/$APP_NAME.app"

cd "$ROOT"
VERSION="$(git rev-parse --short HEAD 2>/dev/null || echo dev)"
go build -ldflags "-X github.com/prestonw/track-app-go/ui.buildVersion=${VERSION}" -o trackapp .

rm -rf "$BUNDLE"
mkdir -p "$BUNDLE/Contents/MacOS"
mkdir -p "$BUNDLE/Contents/Resources"

cp trackapp "$BUNDLE/Contents/MacOS/trackapp"
cp packaging/Info.plist "$BUNDLE/Contents/Info.plist"
chmod +x "$BUNDLE/Contents/MacOS/trackapp"

echo "Built $BUNDLE"
echo "Open with: open \"$BUNDLE\""