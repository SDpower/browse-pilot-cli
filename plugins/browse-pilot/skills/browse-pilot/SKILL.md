# Browse Pilot — 瀏覽器自動化 Skill

## 概述
Browse Pilot 透過 MCP 讓你直接控制使用者的真實瀏覽器（Firefox/Chrome/Edge），
不使用 CDP，不需要 headless 模式。

## 核心工作流程

1. **導航** → `bp_navigate` 開啟目標頁面
2. **觀察** → `bp_state` 取得可互動元素列表（帶索引）
3. **互動** → `bp_click` / `bp_input` 操作元素
4. **驗證** → `bp_wait` 等待結果 + `bp_state` 確認

## 可用 Tools

### 導航
| Tool | 說明 | 必要參數 |
|------|------|---------|
| `bp_navigate` | 開啟 URL | `url` |
| `bp_back` | 上一頁 | — |
| `bp_forward` | 下一頁 | — |
| `bp_reload` | 重新載入 | — |
| `bp_scroll` | 捲動頁面 | `direction` (up/down) |

### 頁面檢查
| Tool | 說明 | 必要參數 |
|------|------|---------|
| `bp_state` | 取得頁面元素列表 | — |
| `bp_screenshot` | 截圖 | — |
| `bp_get` | 取得元素資訊 | `what` (title/html/text/value/attributes/bbox) |

### 互動
| Tool | 說明 | 必要參數 |
|------|------|---------|
| `bp_click` | 點擊元素 | `index` |
| `bp_input` | 輸入文字 | `index`, `text` |
| `bp_type` | 焦點元素輸入 | `text` |
| `bp_keys` | 鍵盤事件 | `keys` (例: "Enter", "Ctrl+a") |
| `bp_select` | 選擇下拉選單 | `index`, `value` |
| `bp_hover` | 滑鼠移入 | `index` |
| `bp_dblclick` | 雙擊 | `index` |
| `bp_rightclick` | 右鍵 | `index` |

### 等待
| Tool | 說明 | 必要參數 |
|------|------|---------|
| `bp_wait` | 等待條件 | `type` (selector/text/url), `value` |

### 分頁與 Cookie
| Tool | 說明 | 必要參數 |
|------|------|---------|
| `bp_tabs` | 管理分頁 | `action` (list/switch/close) |
| `bp_cookies` | 管理 cookies | `action` (get/set/clear) |

### 執行
| Tool | 說明 | 必要參數 |
|------|------|---------|
| `bp_eval` | 執行 JavaScript | `code` |

## 使用要點

### 元素索引
- 所有互動指令使用 `bp_state` 回傳的 `index` 欄位
- 每次頁面變化後應重新呼叫 `bp_state` 取得最新索引
- 索引從 0 開始

### 等待
- 頁面操作後建議使用 `bp_wait` 等待結果載入
- 預設逾時 30 秒，可透過 `timeout` 參數調整

### 錯誤處理
- 元素不存在會回傳 ElementNotFound 錯誤
- 索引過期需重新呼叫 `bp_state`

## 範例：查詢台灣上櫃股票資料

```
1. bp_navigate → https://www.tpex.org.tw/web/stock/aftertrading/broker_trading/brokerBS.php
2. bp_state → 取得表單元素
3. bp_input(index=0, text="6488") → 輸入股票代號
4. bp_click(index=3) → 點擊查詢按鈕
5. bp_wait(type="selector", value="table.result-table") → 等待結果
6. bp_eval(code="...") → 擷取表格資料
```
