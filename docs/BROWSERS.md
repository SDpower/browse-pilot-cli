# 跨瀏覽器相容性

## 瀏覽器矩陣

| 瀏覽器 | Manifest | 背景執行方式 | 通訊層 | 最低版本 |
|--------|----------|------------|--------|---------|
| Firefox | MV2 | Persistent Background Script | WebSocket | 109+ |
| Chrome | MV3 | Service Worker（非持久） | Native Messaging | 110+ |
| Edge | MV3 | Service Worker（非持久） | Native Messaging | 110+ |

---

## 通訊方式差異

### Firefox（WebSocket）

Firefox 的 MV2 Extension 採用持久背景腳本（`"persistent": true`），可維持長期 WebSocket 連線：

```
CLI (Go) ←→ ws://localhost:9222 ←→ Firefox Background Script ←→ Content Script
```

- Extension 主動連線至 CLI 的 WebSocket 伺服器
- 連線保持常駐，指令延遲低
- 支援同時多個 Tab 連線

### Chrome / Edge（Native Messaging）

Chrome/Edge 的 MV3 Extension 以 Service Worker 運作，有 30 秒閒置逾時限制，故採用 Native Messaging：

```
CLI (Go) ←→ stdin/stdout ←→ Chrome/Edge Service Worker ←→ Content Script
```

- Service Worker 透過 `runtime.connectNative` 建立 NM 通道
- CLI 本身兼任 NM Host（無需獨立程序）
- 需要 `bp setup chrome` / `bp setup edge` 安裝 NM Host 清單
- 訊息最大 1 MB（大型截圖需分塊）

---

## API 差異與解法

| API | Firefox MV2 | Chrome/Edge MV3 | 解法 |
|-----|------------|-----------------|------|
| 背景持久性 | `"persistent": true` | Service Worker，30s 逾時 | NM keepalive 心跳 |
| 工具列按鈕 | `browser_action` | `action` | 各自 manifest 定義 |
| 主機權限 | `permissions` 中的 `<all_urls>` | 獨立的 `host_permissions` | 各自 manifest 定義 |
| `scripting` API | 不需要 | 需要 `scripting` permission | Chrome/Edge manifest 額外宣告 |
| Promise API | 原生支援 | 部分不一致 | `webextension-polyfill` 統一 |
| `browser.*` 命名空間 | 原生 `browser.*` | 原生 `chrome.*` | `webextension-polyfill` 統一 |

---

## 程式碼共用程度

| 元件 | Firefox | Chrome | Edge | 共用率 |
|------|---------|--------|------|--------|
| Content Script（DOM 操作） | 共用 | 共用 | 共用 | 100% |
| 指令 Handler | 共用 | 共用 | 共用 | 100% |
| Popup UI | 共用 | 共用 | 共用 | 100% |
| 工具函式（error codes 等） | 共用 | 共用 | 共用 | 100% |
| Background Script | 獨立 | 獨立 | 與 Chrome 共用 | ~0% / 100% |
| Manifest | 獨立 | 獨立 | 近似 Chrome | 0% |

- `extension/shared/` 目錄為三個瀏覽器 100% 共用
- `extension/chrome/` 與 `extension/edge/` 的 Background Script 幾乎相同

---

## 已知限制

### 跨瀏覽器通用限制

- Content Script 在 `document_idle` 注入，頁面元素可能尚未就緒，需搭配 `bp wait` 指令
- 跨來源 iframe 內的元素無法存取（瀏覽器安全性限制）
- `eval_js` 執行的程式碼在頁面沙箱中，無法存取 Extension 的 API

### Firefox 特有限制

- WebSocket 連線在 Firefox 重啟後需要 Extension 重新連線（Extension 頁面仍常駐）
- MV2 未來可能面臨淘汰壓力（Firefox 目前承諾維持 MV2 支援）

### Chrome / Edge 特有限制

- Native Messaging 訊息大小上限 **1 MB**，全頁截圖等大型資料需分塊處理
- Service Worker 閒置 30 秒後休眠，CLI 須透過 keepalive 維持連線
- 需要手動安裝 NM Host 清單（`bp setup chrome`），無法自動部署
- Windows 上 NM Host 清單路徑為 Registry，需要寫入權限

### Edge 特有注意事項

- Edge Extension 使用與 Chrome 相同的 manifest 結構，但 Extension ID 不同（由 Store 分配）
- NM Host 清單路徑與 Chrome 不同（見 INSTALL.md）
