#!/bin/bash
# dev-firefox.sh — Firefox 開發模式
# 先建置 extension，再啟動 CLI WebSocket server

set -euo pipefail

echo "=== Browse Pilot Firefox 開發模式 ==="

# 先建置 extension
bash "$(dirname "$0")/build-extensions.sh"

echo ""
echo "▸ 啟動 CLI WebSocket server..."
echo "  請在 Firefox about:debugging 載入: dist/firefox/manifest.json"
echo ""

# 啟動 CLI（verbose 模式）
exec go run ./cmd/bp/ --verbose 2>&1
