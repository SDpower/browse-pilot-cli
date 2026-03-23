# 使用範例

## 基本瀏覽操作

```bash
# 開啟網頁
bp open https://tw.yahoo.com

# 捲動頁面
bp scroll down
bp scroll down --amount 800
bp scroll top

# 瀏覽歷史
bp back
bp forward
bp reload
```

---

## 查看頁面元素

```bash
# 列出可互動的元素（含索引）
bp state

# 輸出範例：
# [0] <a> "首頁" href=/ visible
# [1] <input> "搜尋" type=text visible
# [2] <button> "搜尋" type=submit visible
```

---

## 表單填寫與提交

```bash
# 取得元素清單
bp state

# 在索引 1 的輸入框填入文字
bp input 1 "台積電"

# 點擊索引 2 的搜尋按鈕
bp click 2

# 等待結果頁面載入
bp wait selector ".search-results"

# 截圖確認結果
bp screenshot --output result.png
```

---

## 等待頁面載入

```bash
# 等待特定 CSS 選擇器出現
bp wait selector "#content" --timeout 30s

# 等待特定文字出現
bp wait text "載入完成"

# 等待 URL 跳轉
bp wait url "*/dashboard*"

# 等待元素消失（loading spinner 結束）
bp wait selector ".loading" --state hidden
```

---

## JavaScript 執行

```bash
# 取得頁面資訊
bp eval "document.title"
bp eval "window.location.href"
bp eval "document.querySelectorAll('table').length"

# 操作 DOM
bp eval "document.getElementById('btn').click()"

# 取得複雜資料
bp eval "JSON.stringify(Array.from(document.querySelectorAll('tr td:first-child')).map(el => el.textContent.trim()))"
```

---

## Cookie 管理

```bash
# 查看目前頁面的 Cookie
bp cookies get
bp cookies get --json

# 設定 Cookie
bp cookies set session_token "abc123xyz" --domain example.com

# 匯出 Cookie（登入狀態備份）
bp cookies export session.json

# 匯入 Cookie（還原登入狀態）
bp cookies import session.json

# 清除 Cookie
bp cookies clear --url https://example.com
```

---

## 多分頁管理

```bash
# 列出所有分頁
bp tabs

# 切換至索引 1 的分頁
bp tab 1

# 關閉目前分頁
bp close-tab

# 關閉索引 2 的分頁
bp close-tab 2
```

---

## 批次抓取（Shell Script）

以下範例抓取多個頁面的標題：

```bash
#!/bin/bash

URLS=(
  "https://mops.twse.com.tw"
  "https://tw.stock.yahoo.com"
  "https://invest.cnyes.com"
)

for url in "${URLS[@]}"; do
  bp open "$url"
  bp wait selector "title" --timeout 15s
  TITLE=$(bp get title)
  echo "$url -> $TITLE"
  sleep 2
done
```

---

## 帶 JSON 輸出的批次腳本

```bash
#!/bin/bash

# 開啟頁面並以 JSON 取得所有元素
bp open "https://example.com" && \
bp state --json | jq '.elements[] | select(.tag == "a") | .name'
```

---

## Python Session 自動化

Python Session 允許在持久的 Python 環境中透過 `browser` 物件操作瀏覽器：

```bash
# 啟動 Python Session
bp python start

# 執行 Python 程式碼
bp python exec "
import json

browser.navigate('https://tw.stock.yahoo.com')
browser.wait_selector('.Fw\\(b\\)')

# 取得頁面狀態
state = browser.get_state()
print(f'找到 {len(state[\"elements\"])} 個元素')

# 截圖
browser.screenshot('/tmp/stock.png')
"

# 停止 Session
bp python stop
```

Python Session 適合需要迴圈處理或複雜邏輯的自動化場景。

---

## Claude Code MCP 操作流程

以下為 Claude Code 透過 MCP 操作瀏覽器的完整流程範例。

### 設定（一次性）

在 `mcp.json` 加入：

```json
{
  "mcpServers": {
    "browse-pilot": {
      "command": "bp",
      "args": ["--mcp"]
    }
  }
}
```

### 操作範例：自動登入並抓取資料

在 Claude Code 中：

```
請幫我：
1. 開啟 https://www.example.com/login
2. 在帳號欄位填入 user@example.com
3. 在密碼欄位填入 (從環境變數 $PASSWORD 讀取)
4. 點擊登入
5. 等待跳轉至 dashboard
6. 截圖儲存為 dashboard.png
```

Claude Code 將依序呼叫：

```
navigate → get_state → input_text(email) → input_text(password)
→ click(submit) → wait_url(*/dashboard*) → screenshot
```

### 操作範例：資料擷取後處理

```
請前往台灣證交所 mops.twse.com.tw，
搜尋公司代號 2330，
找到最新一期的月營收資料，
以表格方式整理後回傳給我。
```

Claude Code 會：

1. 呼叫 `navigate` 開啟頁面
2. 呼叫 `get_state` 找到搜尋欄位
3. 填入資料並送出
4. 呼叫 `wait_selector` 等待結果
5. 呼叫 `eval_js` 擷取表格資料
6. 整理成易讀格式回傳

---

## 診斷與除錯

```bash
# 檢查安裝狀態
bp doctor

# 查看目前連線 Session
bp sessions

# 詳細模式（顯示 JSON-RPC 訊息）
bp open https://example.com --verbose

# 以 JSON 輸出任何指令結果
bp state --json | jq .
bp cookies get --json | jq '.cookies[].name'
```
