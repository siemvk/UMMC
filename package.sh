#!/usr/bin/env bash
set -euo pipefail

# Navigate to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

APP_NAME="UMMC"
BUNDLE_ID="com.siemvk.UMMC"
VERSION="${1:-1.0.0}"
OUTPUT_DIR="${2:-.}"
APP_BUNDLE="$OUTPUT_DIR/$APP_NAME.app"
CONTENTS_DIR="$APP_BUNDLE/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

echo "==> Packaging $APP_NAME.app (v$VERSION)..."

# 1. Check prerequisites
if ! command -v go >/dev/null 2>&1; then
    echo "Error: 'go' is not installed or not in PATH." >&2
    exit 1
fi

# 2. Prepare app bundle directory
rm -rf "$APP_BUNDLE"
mkdir -p "$MACOS_DIR" "$RESOURCES_DIR"

# 3. Build Go executable
echo "==> Building Go binary..."
go build -ldflags="-s -w" -o "$MACOS_DIR/$APP_NAME" .
chmod +x "$MACOS_DIR/$APP_NAME"

# 4. Generate App Icon
echo "==> Processing application icon..."
ICON_SRC="assets/icon.icon"
IMAGE_SRC="$ICON_SRC/Assets/Image.png"

if command -v actool >/dev/null 2>&1 && [ -d "$ICON_SRC" ]; then
    TEMP_PLIST="$(mktemp -t actool_plist_XXXXXX.plist)"
    actool --compile "$RESOURCES_DIR" \
           --platform macosx \
           --minimum-deployment-target 11.0 \
           --app-icon icon \
           --output-partial-info-plist "$TEMP_PLIST" \
           "$ICON_SRC" >/dev/null 2>&1 || true
    rm -f "$TEMP_PLIST"
fi

# Fallback to sips + iconutil if actool didn't produce icon.icns
if [ ! -f "$RESOURCES_DIR/icon.icns" ] && [ -f "$IMAGE_SRC" ] && command -v iconutil >/dev/null 2>&1 && command -v sips >/dev/null 2>&1; then
    TEMP_ICONSET="$(mktemp -d -t iconset_XXXXXX.iconset)"
    sips -z 16 16     "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_16x16.png" >/dev/null 2>&1
    sips -z 32 32     "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_16x16@2x.png" >/dev/null 2>&1
    sips -z 32 32     "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_32x32.png" >/dev/null 2>&1
    sips -z 64 64     "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_32x32@2x.png" >/dev/null 2>&1
    sips -z 128 128   "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_128x128.png" >/dev/null 2>&1
    sips -z 256 256   "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_128x128@2x.png" >/dev/null 2>&1
    sips -z 256 256   "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_256x256.png" >/dev/null 2>&1
    sips -z 512 512   "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_256x256@2x.png" >/dev/null 2>&1
    sips -z 512 512   "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_512x512.png" >/dev/null 2>&1
    sips -z 1024 1024 "$IMAGE_SRC" --out "$TEMP_ICONSET/icon_512x512@2x.png" >/dev/null 2>&1
    iconutil -c icns "$TEMP_ICONSET" -o "$RESOURCES_DIR/icon.icns"
    rm -rf "$TEMP_ICONSET"
fi

# 5. Create Info.plist
echo "==> Creating Info.plist..."
cat << EOF > "$CONTENTS_DIR/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleDisplayName</key>
    <string>$APP_NAME</string>
    <key>CFBundleIdentifier</key>
    <string>$BUNDLE_ID</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleExecutable</key>
    <string>$APP_NAME</string>
    <key>CFBundleIconFile</key>
    <string>icon</string>
    <key>CFBundleIconName</key>
    <string>icon</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
</dict>
</plist>
EOF

# 6. Ad-hoc code sign bundle
if command -v codesign >/dev/null 2>&1; then
    echo "==> Signing application bundle..."
    codesign --force --deep --sign - "$APP_BUNDLE" >/dev/null 2>&1 || true
fi

echo "==> Successfully created: $APP_BUNDLE"
echo "    Run it with: open \"$APP_BUNDLE\""
