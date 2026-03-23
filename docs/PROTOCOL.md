# 通訊協議規格

browse-pilot-cli 的所有訊息均採用 JSON-RPC 2.0 風格格式，透過兩種傳輸通道（WebSocket 或 Native Messaging）傳遞。

---

## JSON-RPC 2.0 訊息格式

### Request（CLI → Extension）

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "navigate",
  "params": {
    "url": "https://example.com"
  }
}
```

| 欄位 | 類型 | 說明 |
|------|------|------|
| `id` | string | UUID v4，用於配對回應 |
| `method` | string | 指令名稱（見下方 Method 表） |
| `params` | object | 指令參數，依 method 不同而異 |

---

### Success Response（Extension → CLI）

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "result": {
    "success": true,
    "url": "https://example.com",
    "title": "Example Domain"
  }
}
```

| 欄位 | 類型 | 說明 |
|------|------|------|
| `id` | string | 對應 Request 的 UUID |
| `result` | object | 回傳資料，依 method 不同而異 |

---

### Error Response（Extension → CLI）

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "error": {
    "code": -32000,
    "message": "Element not found",
    "data": {
      "index": 99
    }
  }
}
```

| 欄位 | 類型 | 說明 |
|------|------|------|
| `id` | string | 對應 Request 的 UUID |
| `error.code` | integer | 錯誤碼（見下方錯誤碼表） |
| `error.message` | string | 人類可讀的錯誤訊息 |
| `error.data` | object | 可選，附加除錯資訊 |

---

## Native Messaging Wire Format（Chrome / Edge）

Chrome/Edge 採用 stdin/stdout 通訊，每則訊息格式為：

```
[ 4 bytes: uint32 little-endian 長度 ][ JSON bytes ]
```

- 長度欄位：無號 32 位元整數，小端序（Little-Endian）
- 訊息最大限制：**1 MB**（1,048,576 bytes）
- 超過限制的大型回應（如全頁截圖）需分塊傳輸

**讀取範例（Go）：**

```go
var length uint32
binary.Read(os.Stdin, binary.LittleEndian, &length)
buf := make([]byte, length)
io.ReadFull(os.Stdin, buf)
```

---

## Method 定義表

### 導航

| Method | Params | Response |
|--------|--------|----------|
| `navigate` | `{url: string}` | `{success, url, title}` |
| `go_back` | `{}` | `{success, url}` |
| `go_forward` | `{}` | `{success, url}` |
| `reload` | `{}` | `{success}` |
| `scroll` | `{direction: string, amount?: number}` | `{success, scrollY}` |

### 頁面狀態

| Method | Params | Response |
|--------|--------|----------|
| `get_state` | `{}` | `{url, title, elements[]}` |
| `screenshot` | `{full?: boolean}` | `{data: base64, width, height}` |
| `get_title` | `{}` | `{title}` |
| `get_html` | `{selector?: string}` | `{html}` |

### 元素擷取

| Method | Params | Response |
|--------|--------|----------|
| `get_text` | `{index: number}` | `{text}` |
| `get_value` | `{index: number}` | `{value}` |
| `get_attributes` | `{index: number}` | `{attributes: object}` |
| `get_bbox` | `{index: number}` | `{x, y, width, height}` |

### 互動

| Method | Params | Response |
|--------|--------|----------|
| `click` | `{index: number}` 或 `{x: number, y: number}` | `{success}` |
| `type_text` | `{text: string}` | `{success}` |
| `input_text` | `{index: number, text: string}` | `{success}` |
| `send_keys` | `{keys: string}` | `{success}` |
| `select_option` | `{index: number, value: string}` | `{success}` |
| `hover` | `{index: number}` | `{success}` |
| `dblclick` | `{index: number}` | `{success}` |
| `rightclick` | `{index: number}` | `{success}` |
| `upload_file` | `{index: number, path: string}` | `{success}` |

### 等待

| Method | Params | Response |
|--------|--------|----------|
| `wait_selector` | `{selector: string, state?: string, timeout?: number}` | `{success, found}` |
| `wait_text` | `{text: string, timeout?: number}` | `{success, found}` |
| `wait_url` | `{pattern: string, timeout?: number}` | `{success, url}` |

### 分頁

| Method | Params | Response |
|--------|--------|----------|
| `get_tabs` | `{}` | `{tabs[]}` |
| `switch_tab` | `{index: number}` | `{success, tabId}` |
| `close_tab` | `{index?: number}` | `{success}` |

### Cookie

| Method | Params | Response |
|--------|--------|----------|
| `get_cookies` | `{url?: string}` | `{cookies[]}` |
| `set_cookie` | `{name, value, domain?, path?, ...}` | `{success}` |
| `clear_cookies` | `{url?: string}` | `{success}` |

### 執行

| Method | Params | Response |
|--------|--------|----------|
| `eval_js` | `{code: string}` | `{result}` |

---

## 錯誤碼表

| Code | 名稱 | 說明 |
|------|------|------|
| `-32700` | Parse Error | JSON 解析失敗 |
| `-32600` | Invalid Request | 訊息格式不正確 |
| `-32601` | Method Not Found | 不支援的 method |
| `-32602` | Invalid Params | 參數格式或型別錯誤 |
| `-32603` | Internal Error | Extension 內部錯誤 |
| `-32000` | Element Not Found | 指定索引的元素不存在 |
| `-32001` | Timeout | 等待操作逾時 |
| `-32002` | Navigation Failed | 導航失敗（網路錯誤等） |
| `-32003` | Script Error | `eval_js` 執行期間發生 JS 錯誤 |
| `-32004` | Tab Not Found | 指定分頁不存在 |
| `-32005` | Cookie Error | Cookie 讀寫失敗 |
| `-32006` | Message Too Large | 回應超過 NM 1MB 限制 |

---

## State 元素格式

`get_state` 回傳的 `elements` 陣列，每個元素格式如下：

```json
{
  "index": 0,
  "tag": "a",
  "type": "",
  "role": "link",
  "name": "首頁",
  "selector": "#nav > a:first-child",
  "visible": true,
  "bbox": {
    "x": 10,
    "y": 20,
    "width": 80,
    "height": 32
  }
}
```

| 欄位 | 說明 |
|------|------|
| `index` | 元素索引（供 CLI 指令使用） |
| `tag` | HTML 標籤名稱（小寫） |
| `type` | `<input>` 的 type 屬性，其他為空 |
| `role` | ARIA role |
| `name` | accessible name（順序：`aria-label` → `placeholder` → 文字內容） |
| `selector` | CSS 選擇器（唯一） |
| `visible` | 是否可見 |
| `bbox` | 邊界框，單位像素 |

### 互動元素篩選規則

納入元素：
- `<a>` 且有 `href`
- `<button>`、`<input>`、`<select>`、`<textarea>`
- `[role="button|link|checkbox|menuitem|tab|switch|...]`
- `[onclick]`、`[tabindex]`（非 -1）

排除元素：
- `display: none`
- `visibility: hidden`
- `aria-hidden="true"`
- `disabled`
