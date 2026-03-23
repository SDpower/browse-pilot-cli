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

### `bp open <url>`

開啟指定 URL。

```bash
bp open https://example.com
bp open https://example.com --json
```

**輸出**（`--json`）：
```json
{"success": true, "url": "https://example.com", "title": "Example Domain"}
```

---

### `bp back`

瀏覽器上一頁。

```bash
bp back
```

---

### `bp forward`

瀏覽器下一頁。

```bash
bp forward
```

---

### `bp reload`

重新載入目前頁面。

```bash
bp reload
```

---

### `bp scroll <direction>`

捲動頁面。

| 參數 | 說明 |
|------|------|
| `direction` | `up`、`down`、`left`、`right`、`top`、`bottom` |
| `--amount` | 捲動距離（像素，預設 300） |

```bash
bp scroll down
bp scroll down --amount 600
bp scroll top
```

---

## 檢查

### `bp state`

取得目前頁面的互動元素清單（索引、標籤、名稱、選擇器）。

```bash
bp state
bp state --json
```

**輸出範例**：
```
[0] <a> "首頁" href=/ visible
[1] <button> "搜尋" type=submit visible
[2] <input> "關鍵字" type=text visible
```

---

### `bp screenshot`

截取目前頁面截圖。

| 選項 | 說明 |
|------|------|
| `--full` | 截取完整頁面（非可視區域） |
| `--output <path>` | 儲存至指定路徑（預設輸出 base64） |

```bash
bp screenshot --output page.png
bp screenshot --full --output full.png
bp screenshot --json   # 輸出 base64 JSON
```

---

## 互動

### `bp click <index>`

點擊指定索引的元素（索引來自 `bp state`）。

```bash
bp click 1
```

也可指定座標：

```bash
bp click --x 100 --y 200
```

---

### `bp type <text>`

在目前焦點元素輸入文字（模擬鍵盤）。

```bash
bp type "Hello World"
```

---

### `bp input <index> <text>`

直接設定指定輸入欄位的值。

```bash
bp input 2 "search keyword"
```

---

### `bp keys <keys>`

傳送特殊按鍵。

```bash
bp keys Enter
bp keys "Ctrl+A"
bp keys "Shift+Tab"
```

---

### `bp select <index> <value>`

選取下拉選單的選項。

```bash
bp select 3 "option-value"
```

---

### `bp hover <index>`

滑鼠移至指定元素（觸發 hover 效果）。

```bash
bp hover 5
```

---

### `bp dblclick <index>`

雙擊指定元素。

```bash
bp dblclick 4
```

---

### `bp rightclick <index>`

右鍵點擊指定元素。

```bash
bp rightclick 2
```

---

## 分頁

### `bp tabs`

列出所有分頁。

```bash
bp tabs
bp tabs --json
```

**輸出範例**：
```
[0] https://example.com — Example Domain (目前)
[1] https://google.com — Google
```

---

### `bp tab <index>`

切換至指定分頁。

```bash
bp tab 1
```

---

### `bp close-tab`

關閉目前分頁，或指定索引的分頁。

```bash
bp close-tab
bp close-tab 2
```

---

## Cookie

### `bp cookies get`

取得目前頁面的 Cookie。

| 選項 | 說明 |
|------|------|
| `--url <url>` | 指定 URL（預設為目前頁面） |

```bash
bp cookies get
bp cookies get --url https://example.com --json
```

---

### `bp cookies set <name> <value>`

設定 Cookie。

```bash
bp cookies set session_id abc123 --domain example.com
```

---

### `bp cookies clear`

清除 Cookie。

```bash
bp cookies clear
bp cookies clear --url https://example.com
```

---

### `bp cookies export <file>`

匯出 Cookie 至 JSON 檔案（Netscape 格式相容）。

```bash
bp cookies export cookies.json
```

---

### `bp cookies import <file>`

從 JSON 檔案匯入 Cookie。

```bash
bp cookies import cookies.json
```

---

## 等待

### `bp wait selector <selector>`

等待 CSS 選擇器元素出現。

| 選項 | 說明 |
|------|------|
| `--state` | `visible`（預設）、`hidden`、`attached` |
| `--timeout` | 逾時秒數（覆蓋全域設定） |

```bash
bp wait selector "#content"
bp wait selector ".loading" --state hidden --timeout 60s
```

---

### `bp wait text <text>`

等待指定文字出現在頁面上。

```bash
bp wait text "載入完成"
bp wait text "Error" --timeout 10s
```

---

### `bp wait url <pattern>`

等待目前 URL 符合指定的模式（支援萬用字元）。

```bash
bp wait url "*/dashboard*"
bp wait url "https://example.com/success"
```

---

## 擷取

### `bp get title`

取得目前頁面標題。

```bash
bp get title
```

---

### `bp get html`

取得頁面 HTML。

| 選項 | 說明 |
|------|------|
| `--selector <css>` | 僅取得符合選擇器的元素 HTML |

```bash
bp get html
bp get html --selector "#main-content"
```

---

### `bp get text <index>`

取得指定元素的純文字內容。

```bash
bp get text 3
```

---

### `bp get value <index>`

取得表單欄位目前的值。

```bash
bp get value 2
```

---

### `bp get attributes <index>`

取得指定元素的所有屬性。

```bash
bp get attributes 1 --json
```

---

### `bp get bbox <index>`

取得指定元素的邊界框（位置與尺寸）。

```bash
bp get bbox 0 --json
```

**輸出**（`--json`）：
```json
{"x": 10, "y": 50, "width": 200, "height": 40}
```

---

## 執行

### `bp eval <code>`

在頁面執行 JavaScript 並回傳結果。

```bash
bp eval "document.title"
bp eval "window.scrollY"
bp eval "JSON.stringify(performance.timing)"
```

---

## Python Session

### `bp python start`

啟動 Python 互動 Session（持久化子程序）。

```bash
bp python start
```

---

### `bp python exec <code>`

在 Python Session 中執行程式碼，可透過 `browser` 物件操作瀏覽器。

```bash
bp python exec "browser.navigate('https://example.com')"
bp python exec "title = browser.get_title(); print(title)"
```

---

### `bp python stop`

結束 Python Session。

```bash
bp python stop
```

---

## 系統

### `bp doctor`

診斷安裝狀態，回報瀏覽器偵測結果與連線狀態。

```bash
bp doctor
bp doctor --json
```

---

### `bp status`

顯示目前連線的 Extension 狀態。

```bash
bp status
```

---

### `bp sessions`

列出目前所有連線的 Extension Session（多分頁、多瀏覽器）。

```bash
bp sessions --json
```

---

### `bp setup <browser>`

安裝 Native Messaging Host 清單至系統。

```bash
bp setup firefox
bp setup chrome
bp setup edge
```

---

### `bp close`

關閉目前 Session 連線。

```bash
bp close
```
