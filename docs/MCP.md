# MCP 整合指南

## 什麼是 MCP

MCP（Model Context Protocol）是 Anthropic 制定的開放協議，讓 AI Agent（如 Claude Code）能夠以標準化方式呼叫外部工具。

browse-pilot-cli 內建 MCP Server 模式，透過 `bp_cli --mcp` 旗標啟動後，Claude Code 可直接以自然語言指示 CLI 操作瀏覽器，無需手動撰寫指令。

---

## 設定方式

在 Claude Code 的設定檔（`mcp.json`）中加入 browse-pilot 的設定：

```json
{
  "mcpServers": {
    "browse-pilot": {
      "command": "bp_cli",
      "args": ["--mcp"],
      "env": {
        "BROWSER": "firefox"
      }
    }
  }
}
```

或指定特定瀏覽器與連接埠：

```json
{
  "mcpServers": {
    "browse-pilot-chrome": {
      "command": "bp_cli",
      "args": ["--mcp", "--browser", "chrome", "--port", "9223"]
    }
  }
}
```

完整的設定範例檔案位於專案根目錄：`mcp.json.example`。

---

## 可用的 MCP Tools

MCP Server 模式下，以下 20 個工具可供 AI Agent 呼叫：

| Tool 名稱 | 說明 |
|-----------|------|
| `navigate` | 開啟指定 URL |
| `go_back` | 瀏覽器上一頁 |
| `go_forward` | 瀏覽器下一頁 |
| `reload` | 重新載入頁面 |
| `scroll` | 捲動頁面 |
| `get_state` | 取得頁面互動元素清單 |
| `screenshot` | 截取頁面截圖 |
| `click` | 點擊元素 |
| `type_text` | 在目前焦點輸入文字 |
| `input_text` | 設定輸入欄位值 |
| `send_keys` | 傳送特殊按鍵 |
| `select_option` | 選取下拉選單 |
| `get_cookies` | 取得 Cookie |
| `set_cookie` | 設定 Cookie |
| `clear_cookies` | 清除 Cookie |
| `wait_selector` | 等待 CSS 選擇器元素出現 |
| `wait_text` | 等待文字出現 |
| `eval_js` | 執行 JavaScript |
| `get_tabs` | 列出所有分頁 |
| `switch_tab` | 切換分頁 |

---

## 可用的 MCP Resources

| Resource URI | 說明 |
|-------------|------|
| `browser://state` | 目前頁面的互動元素狀態（JSON） |
| `browser://screenshot` | 目前頁面截圖（base64 PNG） |
| `browser://tabs` | 所有分頁清單（JSON） |
| `browser://cookies` | 目前頁面 Cookie（JSON） |
| `browser://url` | 目前頁面 URL |
| `browser://title` | 目前頁面標題 |

---

## Claude Code 使用範例

### 基本網頁操作

在 Claude Code 中，可以直接以繁體中文或英文描述需求：

```
請用瀏覽器開啟 https://tw.yahoo.com，然後搜尋「台積電」，截圖給我看。
```

Claude Code 會自動呼叫：
1. `navigate` → 開啟 Yahoo
2. `get_state` → 找到搜尋欄位
3. `input_text` → 輸入「台積電」
4. `click` → 點擊搜尋按鈕
5. `wait_selector` → 等待結果載入
6. `screenshot` → 截圖回傳

### 資料擷取

```
請前往 https://mops.twse.com.tw 找到台積電最新的月營收資料，以 JSON 格式回傳。
```

### Cookie 管理

```
請匯出目前瀏覽器的所有 Cookie，儲存至 session.json，之後我需要恢復這個登入狀態。
```

---

## SKILL.md 說明

`SKILL.md` 是放置於專案目錄中的技能描述檔案，讓 Claude Code 知道 browse-pilot 的能力與使用方式。

建議在需要自動化瀏覽器操作的專案中加入以下內容至 `CLAUDE.md` 或 `SKILL.md`：

```markdown
## 瀏覽器自動化

本專案使用 browse-pilot-cli（`bp_cli`）控制瀏覽器。
MCP Server 已在 mcp.json 設定完成，可直接呼叫以下工具：

- navigate, get_state, click, input_text, screenshot
- wait_selector, eval_js, get_cookies, set_cookie

操作瀏覽器時，請先呼叫 get_state 取得元素清單，再根據索引操作。
```

---

## 運作原理

`bp_cli --mcp` 啟動時的流程：

1. **啟動 WS/NM server** — 在背景啟動 WebSocket server（Firefox）或 NM host（Chrome/Edge），不阻塞等待 Extension 連入
2. **回應 MCP initialize** — MCP server 立即開始讀取 stdin，回應 Claude Code 的 `initialize` 請求
3. **Extension 連入** — 瀏覽器 Extension 在背景自動連入 WS server
4. **Tool 呼叫** — Claude Code 呼叫 `bp_navigate` 等 MCP tool 時，`Send()` 會自動等待 Extension 連線就緒後再轉發指令

因此 **Extension 不需要在 MCP 啟動前就連上**，MCP server 會先回應協議握手，Extension 隨後連入即可。

## 注意事項

- MCP Server 透過 stdio（標準輸入/輸出）與 Claude Code 通訊，啟動後不應有其他程序佔用 stdin/stdout
- verbose 日誌寫入 stderr，不會干擾 MCP stdio 通訊
- 若瀏覽器 Extension 尚未連入，Tool 呼叫會等待連線（受 `--timeout` 控制，預設 30 秒）
- 大型截圖（全頁）在 Chrome/Edge 下可能因 NM 1MB 限制而失敗，建議使用 Firefox 進行全頁截圖
- Firefox 使用 WebSocket 長連線，MCP 模式下 Extension 連入後保持連線不斷開
