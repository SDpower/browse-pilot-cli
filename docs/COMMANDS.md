# 指令參考

## 全域選項

所有指令均支援以下全域選項：

| 選項 | 預設值 | 說明 |
|------|--------|------|
| `--browser` | 自動偵測 | 指定瀏覽器（`firefox`、`chrome`、`edge`） |
| `--port` | `9222` | WebSocket 伺服器埠號（Firefox 用） |
| `--json` | false | 以 JSON 格式輸出結果 |
| `--timeout` | `30s` | 指令逾時時間 |
| `--verbose` | false | 顯示詳細的通訊紀錄 |
| `--mcp` | false | 以 MCP Server 模式啟動（stdio） |

---

## Exit Code 對照表

| Code | 意義 |
|------|------|
| 0 | 成功 |
| 1 | 一般錯誤 |
| 2 | 連線錯誤（Extension 未連線） |
| 3 | 逾時（操作超過 --timeout） |

---

## 導航

### `bp_cli open <url>`

開啟指定 URL。

```bash
bp_cli open https://example.com
bp_cli open https://example.com --json
```

**輸出**（`--json`）：
```json
{"success": true, "url": "https://example.com", "title": "Example Domain"}
```

---

### `bp_cli back`

瀏覽器上一頁。

```bash
bp_cli back
```

---

### `bp_cli forward`

瀏覽器下一頁。

```bash
bp_cli forward
```

---

### `bp_cli reload`

重新載入目前頁面。

```bash
bp_cli reload
```

---

### `bp_cli scroll <direction>`

捲動頁面。

| 參數 | 說明 |
|------|------|
| `direction` | `up`、`down`、`left`、`right`、`top`、`bottom` |
| `--amount` | 捲動距離（像素，預設 300） |

```bash
bp_cli scroll down
bp_cli scroll down --amount 600
bp_cli scroll top
```

---

## 檢查

### `bp_cli state`

取得目前頁面的互動元素清單（索引、標籤、名稱、選擇器）。

```bash
bp_cli state
bp_cli state --json
```

**輸出範例**：
```
[0] <a> "首頁" href=/ visible
[1] <button> "搜尋" type=submit visible
[2] <input> "關鍵字" type=text visible
```

---

### `bp_cli screenshot`

截取目前頁面截圖。

| 選項 | 說明 |
|------|------|
| `--full` | 截取完整頁面（非可視區域） |
| `--output <path>` | 儲存至指定路徑（預設輸出 base64） |

```bash
bp_cli screenshot --output page.png
bp_cli screenshot --full --output full.png
bp_cli screenshot --json   # 輸出 base64 JSON
```

---

## 互動

### `bp_cli click <index>`

點擊指定索引的元素（索引來自 `bp_cli state`）。

```bash
bp_cli click 1
```

也可指定座標：

```bash
bp_cli click --x 100 --y 200
```

---

### `bp_cli type <text>`

在目前焦點元素輸入文字（模擬鍵盤）。

```bash
bp_cli type "Hello World"
```

---

### `bp_cli input <index> <text>`

直接設定指定輸入欄位的值。

```bash
bp_cli input 2 "search keyword"
```

---

### `bp_cli keys <keys>`

傳送特殊按鍵。

```bash
bp_cli keys Enter
bp_cli keys "Ctrl+A"
bp_cli keys "Shift+Tab"
```

---

### `bp_cli select <index> <value>`

選取下拉選單的選項。

```bash
bp_cli select 3 "option-value"
```

---

### `bp_cli hover <index>`

滑鼠移至指定元素（觸發 hover 效果）。

```bash
bp_cli hover 5
```

---

### `bp_cli dblclick <index>`

雙擊指定元素。

```bash
bp_cli dblclick 4
```

---

### `bp_cli rightclick <index>`

右鍵點擊指定元素。

```bash
bp_cli rightclick 2
```

---

## 分頁

### `bp_cli tabs`

列出所有分頁。

```bash
bp_cli tabs
bp_cli tabs --json
```

**輸出範例**：
```
[0] https://example.com — Example Domain (目前)
[1] https://google.com — Google
```

---

### `bp_cli tab <index>`

切換至指定分頁。

```bash
bp_cli tab 1
```

---

### `bp_cli close-tab`

關閉目前分頁，或指定索引的分頁。

```bash
bp_cli close-tab
bp_cli close-tab 2
```

---

## Cookie

### `bp_cli cookies get`

取得目前頁面的 Cookie。

| 選項 | 說明 |
|------|------|
| `--url <url>` | 指定 URL（預設為目前頁面） |

```bash
bp_cli cookies get
bp_cli cookies get --url https://example.com --json
```

---

### `bp_cli cookies set <name> <value>`

設定 Cookie。

```bash
bp_cli cookies set session_id abc123 --domain example.com
```

---

### `bp_cli cookies clear`

清除 Cookie。

```bash
bp_cli cookies clear
bp_cli cookies clear --url https://example.com
```

---

### `bp_cli cookies export <file>`

匯出 Cookie 至 JSON 檔案（Netscape 格式相容）。

```bash
bp_cli cookies export cookies.json
```

---

### `bp_cli cookies import <file>`

從 JSON 檔案匯入 Cookie。

```bash
bp_cli cookies import cookies.json
```

---

## 等待

### `bp_cli wait selector <selector>`

等待 CSS 選擇器元素出現。

| 選項 | 說明 |
|------|------|
| `--state` | `visible`（預設）、`hidden`、`attached` |
| `--timeout` | 逾時秒數（覆蓋全域設定） |

```bash
bp_cli wait selector "#content"
bp_cli wait selector ".loading" --state hidden --timeout 60s
```

---

### `bp_cli wait text <text>`

等待指定文字出現在頁面上。

```bash
bp_cli wait text "載入完成"
bp_cli wait text "Error" --timeout 10s
```

---

### `bp_cli wait url <pattern>`

等待目前 URL 符合指定的模式（支援萬用字元）。

```bash
bp_cli wait url "*/dashboard*"
bp_cli wait url "https://example.com/success"
```

---

## 擷取

### `bp_cli get title`

取得目前頁面標題。

```bash
bp_cli get title
```

---

### `bp_cli get html`

取得頁面 HTML。

| 選項 | 說明 |
|------|------|
| `--selector <css>` | 僅取得符合選擇器的元素 HTML |

```bash
bp_cli get html
bp_cli get html --selector "#main-content"
```

---

### `bp_cli get text <index>`

取得指定元素的純文字內容。

```bash
bp_cli get text 3
```

---

### `bp_cli get value <index>`

取得表單欄位目前的值。

```bash
bp_cli get value 2
```

---

### `bp_cli get attributes <index>`

取得指定元素的所有屬性。

```bash
bp_cli get attributes 1 --json
```

---

### `bp_cli get bbox <index>`

取得指定元素的邊界框（位置與尺寸）。

```bash
bp_cli get bbox 0 --json
```

**輸出**（`--json`）：
```json
{"x": 10, "y": 50, "width": 200, "height": 40}
```

---

## 執行

### `bp_cli eval <code>`

在頁面執行 JavaScript 並回傳結果。

```bash
bp_cli eval "document.title"
bp_cli eval "window.scrollY"
bp_cli eval "JSON.stringify(performance.timing)"
```

---

## Python Session

### `bp_cli python start`

啟動 Python 互動 Session（持久化子程序）。

```bash
bp_cli python start
```

---

### `bp_cli python exec <code>`

在 Python Session 中執行程式碼，可透過 `browser` 物件操作瀏覽器。

```bash
bp_cli python exec "browser.navigate('https://example.com')"
bp_cli python exec "title = browser.get_title(); print(title)"
```

---

### `bp_cli python stop`

結束 Python Session。

```bash
bp_cli python stop
```

---

## 系統

### `bp_cli doctor`

診斷安裝狀態，回報瀏覽器偵測結果與連線狀態。

```bash
bp_cli doctor
bp_cli doctor --json
```

---

### `bp_cli status`

顯示目前連線的 Extension 狀態。

```bash
bp_cli status
```

---

### `bp_cli sessions`

列出目前所有連線的 Extension Session（多分頁、多瀏覽器）。

```bash
bp_cli sessions --json
```

---

### `bp_cli setup <browser>`

安裝 Native Messaging Host 清單至系統。

```bash
bp_cli setup firefox
bp_cli setup chrome
bp_cli setup edge
```

---

### `bp_cli close`

關閉目前 Session 連線。

```bash
bp_cli close
```
