#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="Track App"
BUNDLE="$ROOT/$APP_NAME.app"
ICON_SRC="$ROOT/assets/icon.png"

cd "$ROOT"
VERSION="$(git rev-parse --short HEAD 2>/dev/null || echo dev)"
go build -ldflags "-X github.com/prestonw/track-app-go/ui.buildVersion=${VERSION}" -o trackapp .

rm -rf "$BUNDLE"
mkdir -p "$BUNDLE/Contents/MacOS"
mkdir -p "$BUNDLE/Contents/Resources"

cp trackapp "$BUNDLE/Contents/MacOS/trackapp"
cp packaging/Info.plist "$BUNDLE/Contents/Info.plist"
chmod +x "$BUNDLE/Contents/MacOS/trackapp"

if [[ -f "$ICON_SRC" ]]; then
  ICNS="$BUNDLE/Contents/Resources/AppIcon.icns"
  if command -v iconutil >/dev/null && command -v sips >/dev/null; then
    ICONSET="$(mktemp -d)/AppIcon.iconset"
    mkdir -p "$ICONSET"
    for size in 16 32 128 256 512; do
      sips -z "$size" "$size" "$ICON_SRC" --out "$ICONSET/icon_${size}x${size}.png" >/dev/null
      dbl=$((size * 2))
      sips -z "$dbl" "$dbl" "$ICON_SRC" --out "$ICONSET/icon_${size}x${size}@2x.png" >/dev/null
    done
    iconutil -c icns "$ICONSET" -o "$ICNS"
    rm -rf "$(dirname "$ICONSET")"
    echo "Bundled AppIcon.icns"
  else
    cp "$ICON_SRC" "$BUNDLE/Contents/Resources/AppIcon.png"
    /usr/libexec/PlistBuddy -c "Set :CFBundleIconFile AppIcon.png" "$BUNDLE/Contents/Info.plist" 2>/dev/null || true
    echo "Bundled AppIcon.png (install iconutil for .icns)"
  fi
fi

echo "Built $BUNDLE (version $VERSION)"
echo "Open with: open \"$BUNDLE\""