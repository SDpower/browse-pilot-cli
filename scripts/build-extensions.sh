#!/bin/bash
# build-extensions.sh — 從 shared + browser-specific 組合產生三份 extension
# 用法: bash scripts/build-extensions.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
EXT_DIR="$PROJECT_DIR/extension"
DIST_DIR="$PROJECT_DIR/dist"

echo "=== Browse Pilot Extension Build ==="
echo "來源: $EXT_DIR"
echo "輸出: $DIST_DIR"
echo ""

# 清理舊的 dist 目錄
rm -rf "$DIST_DIR"

# --- Firefox (MV2) ---
# Firefox 使用 persistent background page，可直接載入多個 script
# shared code 透過 manifest 的 background.scripts 陣列依序載入，不需要合併
echo "▸ 建置 Firefox extension..."
mkdir -p "$DIST_DIR/firefox/background"
cp -r "$EXT_DIR/shared" "$DIST_DIR/firefox/shared"
cp -r "$EXT_DIR/firefox/background/"* "$DIST_DIR/firefox/background/"
cp "$EXT_DIR/firefox/manifest.json" "$DIST_DIR/firefox/"
echo "  ✓ dist/firefox/"

# --- Chrome (MV3) ---
# MV3 service worker 使用 module type，但 shared code 是非 module 的全域變數風格
# 解法：將 shared utils + handlers 合併（concat）到 service-worker.js 前方
echo "▸ 建置 Chrome extension..."
mkdir -p "$DIST_DIR/chrome/background"
cp -r "$EXT_DIR/shared" "$DIST_DIR/chrome/shared"

# 合併 shared code 到 service worker 前方
{
  echo "// === 自動合併的共用程式碼（由 build-extensions.sh 產生）==="
  echo "// 來源: extension/shared/utils/ 與 extension/shared/handlers/"
  echo ""
  cat "$EXT_DIR/shared/utils/error.js"
  echo ""
  cat "$EXT_DIR/shared/utils/timeout.js"
  echo ""
  cat "$EXT_DIR/shared/utils/router.js"
  echo ""
  cat "$EXT_DIR/shared/handlers/navigation.js"
  echo ""
  cat "$EXT_DIR/shared/handlers/tabs.js"
  echo ""
  cat "$EXT_DIR/shared/handlers/cookies.js"
  echo ""
  cat "$EXT_DIR/shared/handlers/screenshot.js"
  echo ""
  cat "$EXT_DIR/shared/handlers/session.js"
  echo ""
  echo "// === Service Worker 主程式 ==="
  echo ""
  cat "$EXT_DIR/chrome/background/service-worker.js"
} > "$DIST_DIR/chrome/background/service-worker.js"

cp "$EXT_DIR/chrome/manifest.json" "$DIST_DIR/chrome/"
echo "  ✓ dist/chrome/"

# --- Edge (MV3, 基於 Chrome) ---
# Edge 使用與 Chrome 相同的 MV3 架構
# service-worker.js 直接複製 Chrome 合併後的版本
echo "▸ 建置 Edge extension..."
mkdir -p "$DIST_DIR/edge/background"
cp -r "$EXT_DIR/shared" "$DIST_DIR/edge/shared"

# 複用 Chrome 合併好的 service worker
cp "$DIST_DIR/chrome/background/service-worker.js" "$DIST_DIR/edge/background/"
cp "$EXT_DIR/edge/manifest.json" "$DIST_DIR/edge/"
echo "  ✓ dist/edge/"

echo ""
echo "=== 建置完成 ==="
echo "Firefox: $DIST_DIR/firefox/"
echo "Chrome:  $DIST_DIR/chrome/"
echo "Edge:    $DIST_DIR/edge/"
