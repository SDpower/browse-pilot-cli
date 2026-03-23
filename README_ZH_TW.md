# browse-pilot-cli (bp)

[![Version](https://img.shields.io/badge/version-0.1.0-blue)](CHANGELOG.md)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

跨瀏覽器自動化 CLI 工具，透過 WebExtension API 控制 Firefox、Chrome、Edge，不依賴 CDP。

## 特色

- 🌐 支援 Firefox、Chrome、Edge 三大瀏覽器
- 🔒 使用真實瀏覽器 profile（cookie、登入狀態完整保留）
- 🚫 不依賴 CDP — 透過 WebExtension API，反偵測能力強
- 🤖 原生 MCP 支援 — 直接對接 Claude Code
- 🐍 Python session — 透過 `browser` 物件編寫自動化腳本
- ⚡ 單一 Go binary — 同時提供 CLI、WebSocket server、Native Messaging host、MCP server

> 📖 **English version**: [README.md](README.md)

## 安裝

### 前置需求

- Go 1.22+
- Node.js 18+（Extension 建置與 lint）
- Firefox 109+ / Chrome 110+ / Edge 110+

### 從原始碼建置

```bash
# 安裝 Go binary
go install github.com/SDpower/browse-pilot-cli/cmd/bp@latest

# 或從原始碼建置
git clone https://github.com/SDpower/browse-pilot-cli.git
cd browse-pilot-cli
make build
```

### Extension 安裝

先建置 Extension：

```bash
bash scripts/build-extensions.sh
```

#### Firefox

1. 開啟 `about:debugging`
2. 點選「這個 Firefox」
3. 點選「載入暫時性附加元件」
4. 選取 `dist/firefox/manifest.json`

#### Chrome

1. 開啟 `chrome://extensions`
2. 啟用右上角「開發者模式」
3. 點選「載入未封裝項目」
4. 選取 `dist/chrome/` 目錄

#### Edge

1. 開啟 `edge://extensions`
2. 啟用左側「開發者模式」
3. 點選「載入未封裝項目」
4. 選取 `dist/edge/` 目錄

### Native Messaging Host 設定

Chrome 和 Edge 透過 Native Messaging 通訊，需先執行設定指令：

```bash
bp_cli setup firefox
bp_cli setup chrome
bp_cli setup edge

# 或一次設定所有瀏覽器
bp_cli setup --all
```

## 快速開始

### 檢查環境

```bash
bp_cli doctor
```

### 基本操作

```bash
# 開啟網頁
bp_cli open https://example.com

# 取得頁面狀態（列出所有可互動元素）
bp_cli state

# 點擊元素（依索引）
bp_cli click 0

# 輸入文字到指定欄位
bp_cli input 1 "hello world"

# 截圖
bp_cli screenshot output.png
```

### 等待與擷取

```bash
# 等待元素出現
bp_cli wait selector "table.result"

# 等待頁面文字
bp_cli wait text "載入完成"

# 執行 JavaScript
bp_cli eval "document.querySelectorAll('tr').length"

# 取得頁面資訊
bp_cli get title
bp_cli get html --selector "table"
bp_cli get text 2
```

### Python 自動化

```bash
# 執行 Python 程式碼（可存取 browser 物件）
bp_cli python "result = browser.state(); print(len(result['elements']))"

# 執行 Python 腳本
bp_cli python --file script.py

# 列出 session 變數
bp_cli python --vars

# 重置 session
bp_cli python --reset
```

## 指令參考

### 全域選項

| 選項 | 說明 | 預設值 |
|------|------|--------|
| `--browser` | 目標瀏覽器（firefox / chrome / edge / auto） | auto |
| `--port` | WebSocket 埠號 | 9222 |
| `--json` | JSON 格式輸出 | false |
| `--timeout` | 逾時時間（毫秒） | 30000 |
| `--verbose` | 詳細日誌 | false |
| `--mcp` | 以 MCP server 模式執行 | false |
| `--session` | 連線 session 名稱 | default |
| `--native-messaging` | 以 NM host 模式啟動 | false |

### 導航

| 指令 | 說明 |
|------|------|
| `bp_cli open <url>` | 開啟指定網址 |
| `bp_cli back` | 上一頁 |
| `bp_cli forward` | 下一頁 |
| `bp_cli reload` | 重新載入 |
| `bp_cli scroll <up\|down>` | 捲動頁面（可加 `--amount <px>`） |

### 頁面檢查

| 指令 | 說明 |
|------|------|
| `bp_cli state` | 列出目前 URL、標題與所有可互動元素 |
| `bp_cli screenshot [path]` | 截圖（省略路徑則輸出 base64，可加 `--full`） |

### 互動

| 指令 | 說明 |
|------|------|
| `bp_cli click <index>` | 點擊元素（也可 `bp_cli click <x> <y>` 座標點擊） |
| `bp_cli dblclick <index>` | 雙擊元素 |
| `bp_cli rightclick <index>` | 右鍵點擊元素 |
| `bp_cli hover <index>` | 滑鼠懸停 |
| `bp_cli type <text>` | 輸入文字（焦點元素） |
| `bp_cli input <index> <text>` | 輸入文字到指定欄位 |
| `bp_cli keys <keys>` | 傳送按鍵（如 `Enter`、`Ctrl+a`） |
| `bp_cli select <index> <value>` | 選取下拉選單選項 |
| `bp_cli upload <index> <path>` | 上傳檔案至 file input |

### 分頁管理

| 指令 | 說明 |
|------|------|
| `bp_cli tabs` | 列出所有分頁 |
| `bp_cli tab <index>` | 切換到指定分頁 |
| `bp_cli close-tab [index]` | 關閉分頁（預設當前） |

### Cookie

| 指令 | 說明 |
|------|------|
| `bp_cli cookies get [--url <url>]` | 取得 cookie 列表 |
| `bp_cli cookies set <name> <value>` | 設定 cookie（可加 `--domain`、`--secure`、`--same-site`） |
| `bp_cli cookies clear [--url <url>]` | 清除 cookie |
| `bp_cli cookies export <file>` | 匯出 cookie 到 JSON 檔 |
| `bp_cli cookies import <file>` | 從 JSON 檔匯入 cookie |

### 等待

| 指令 | 說明 |
|------|------|
| `bp_cli wait selector <css>` | 等待 CSS selector 出現（可加 `--hidden` 等待消失） |
| `bp_cli wait text <text>` | 等待頁面出現指定文字 |
| `bp_cli wait url <pattern>` | 等待 URL 符合 pattern |

所有 wait 指令支援 `--timeout <ms>`（預設 30000）。

### 擷取

| 指令 | 說明 |
|------|------|
| `bp_cli get title` | 取得頁面標題 |
| `bp_cli get html [--selector <css>]` | 取得 HTML 內容 |
| `bp_cli get text <index>` | 取得元素文字 |
| `bp_cli get value <index>` | 取得表單欄位值 |
| `bp_cli get attributes <index>` | 取得元素所有屬性 |
| `bp_cli get bbox <index>` | 取得元素位置與尺寸 |

### 執行

| 指令 | 說明 |
|------|------|
| `bp_cli eval <code>` | 在頁面 context 執行 JavaScript |
| `bp_cli python <code>` | 執行 Python（可存取 `browser` 物件） |
| `bp_cli python --file <path>` | 執行 Python 腳本 |
| `bp_cli python --vars` | 列出 session 變數 |
| `bp_cli python --reset` | 重置 Python session |

### 系統

| 指令 | 說明 |
|------|------|
| `bp_cli doctor` | 檢測瀏覽器連線狀態 |
| `bp_cli status` | 顯示目前連線資訊 |
| `bp_cli sessions` | 列出活躍 session |
| `bp_cli close [--all]` | 關閉連線 |
| `bp_cli setup <browser>` | 設定 Native Messaging Host（可加 `--all`） |

## MCP 整合（Claude Code）

`bp_cli` 支援 [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)，可直接作為 Claude Code 的瀏覽器控制工具。

### 設定

在 `.claude/mcp.json` 或 `claude_desktop_config.json` 加入：

```json
{
  "mcpServers": {
    "browse-pilot": {
      "command": "bp_cli",
      "args": ["--mcp"],
      "env": {
        "BP_BROWSER": "firefox",
        "BP_PORT": "9222"
      }
    }
  }
}
```

### 可用 MCP Tools

| Tool | 說明 |
|------|------|
| `bp_navigate` | 開啟網址 |
| `bp_state` | 取得頁面狀態與可互動元素 |
| `bp_click` | 點擊元素 |
| `bp_input` | 輸入文字 |
| `bp_type` | 焦點元素輸入 |
| `bp_screenshot` | 截圖 |
| `bp_eval` | 執行 JavaScript |
| `bp_scroll` | 捲動頁面 |
| `bp_keys` | 送出鍵盤事件 |
| `bp_wait` | 等待條件 |
| `bp_get` | 取得元素資訊 |
| `bp_tabs` | 管理分頁 |
| `bp_cookies` | 管理 cookie |
| `bp_upload` | 上傳檔案 |
| `bp_hover` | 滑鼠懸停 |
| `bp_dblclick` | 雙擊 |
| `bp_rightclick` | 右鍵 |
| `bp_select` | 選取下拉選單 |
| `bp_back` / `bp_forward` / `bp_reload` | 導航控制 |

## 架構

```
Claude Code / AI Agent
    │ MCP protocol (stdio)
    ▼
browse-pilot-cli (Go binary)
    ├── WebSocket Server ──→ Firefox Extension (MV2，持久背景頁)
    └── Native Messaging ──→ Chrome/Edge Extension (MV3，service worker)
                                    │
                                    ▼
                              Content Script → 目標網頁
```

### 雙軌通訊策略

| 瀏覽器 | 通訊方式 | 原因 |
|--------|----------|------|
| Firefox | WebSocket | MV2 持久背景頁支援長效連線 |
| Chrome | Native Messaging | MV3 service worker 有閒置逾時限制 |
| Edge | Native Messaging | 與 Chrome 相同（Chromium 核心） |

兩種通訊方式均使用 JSON-RPC 2.0 訊息格式。上層指令邏輯與通訊方式無關。

## 開發

```bash
# 建置 Go binary
make build

# 執行測試
make test

# 執行 lint
make lint

# 建置 Extension
make extension-build

# Extension lint
make extension-lint
```

## 文件

- [安裝指南](docs/INSTALL.md)
- [完整指令參考](docs/COMMANDS.md)
- [通訊協議](docs/PROTOCOL.md)
- [跨瀏覽器相容性](docs/BROWSERS.md)
- [MCP 整合指南](docs/MCP.md)
- [使用範例](docs/EXAMPLES.md)

## 變更紀錄

詳見 [CHANGELOG.md](CHANGELOG.md)。

## 授權

[MIT](LICENSE) © [@SteveLuo](https://github.com/sdpower)
