#!/bin/bash
# dev-chrome.sh — Chrome/Edge 開發模式
# 先建置 extension，再提示手動載入步驟

set -euo pipefail

echo "=== Browse Pilot Chrome 開發模式 ==="

# 先建置 extension
bash "$(dirname "$0")/build-extensions.sh"

echo ""
echo "▸ 請在 Chrome chrome://extensions 載入: dist/chrome/"
echo "▸ 請確認已執行: bp setup chrome"
echo ""
