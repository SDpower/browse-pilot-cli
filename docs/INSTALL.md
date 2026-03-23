# 安裝指南

## 系統需求

| 項目 | 最低版本 |
|------|---------|
| Go | 1.22+ |
| Node.js | 18+（Extension 建置用） |
| Firefox | 109+ |
| Chrome | 110+ |
| Edge | 110+ |
| 作業系統 | macOS、Linux、Windows 10+ |

---

## 從原始碼建置

```bash
# 取得原始碼
git clone https://github.com/SDpower/browse-pilot-cli.git
cd browse-pilot-cli

# 建置 CLI binary
go build -o bp ./cmd/bp/

# 建置 Extension（需要 Node.js）
npm install
bash scripts/build-extensions.sh
# 輸出至 dist/firefox/ dist/chrome/ dist/edge/
```

建置成功後，將 `bp` 移至 PATH 內的目錄：

```bash
# macOS / Linux
sudo mv bp /usr/local/bin/bp

# 驗證
bp --version
```

---

## Extension 安裝步驟

### Firefox

1. 開啟 Firefox，網址列輸入 `about:debugging`
2. 點選「此 Firefox」→「載入暫時性附加元件…」
3. 選擇 `dist/firefox/manifest.json`
4. Extension 圖示出現於工具列即完成

> 正式發佈後可從 AMO 安裝，免除步驟 1–3。

### Chrome

1. 開啟 Chrome，網址列輸入 `chrome://extensions`
2. 開啟右上角「開發人員模式」
3. 點選「載入未封裝項目」，選擇 `dist/chrome/` 資料夾
4. Extension 圖示出現於工具列即完成

### Edge

1. 開啟 Edge，網址列輸入 `edge://extensions`
2. 開啟左側「開發人員模式」
3. 點選「載入解壓縮的擴充功能」，選擇 `dist/edge/` 資料夾
4. Extension 圖示出現於工具列即完成

---

## Native Messaging Host 安裝

Chrome 與 Edge 透過 Native Messaging 與 CLI 通訊，需要安裝 NM Host 清單：

```bash
bp setup firefox   # 安裝 Firefox NM Host
bp setup chrome    # 安裝 Chrome NM Host
bp setup edge      # 安裝 Edge NM Host
```

`bp setup` 會自動：
1. 將 `bp` binary 的絕對路徑寫入 NM Host 清單（JSON）
2. 將清單複製至各平台標準目錄

### 各平台 NM Host 路徑

#### Firefox

| 平台 | 路徑 |
|------|------|
| macOS | `~/Library/Application Support/Mozilla/NativeMessagingHosts/browse_pilot.json` |
| Linux | `~/.mozilla/native-messaging-hosts/browse_pilot.json` |
| Windows | `HKCU\Software\Mozilla\NativeMessagingHosts\browse_pilot` |

#### Chrome

| 平台 | 路徑 |
|------|------|
| macOS | `~/Library/Application Support/Google/Chrome/NativeMessagingHosts/browse_pilot.json` |
| Linux | `~/.config/google-chrome/NativeMessagingHosts/browse_pilot.json` |
| Windows | `HKCU\Software\Google\Chrome\NativeMessagingHosts\browse_pilot` |

#### Edge

| 平台 | 路徑 |
|------|------|
| macOS | `~/Library/Application Support/Microsoft Edge/NativeMessagingHosts/browse_pilot.json` |
| Linux | `~/.config/microsoft-edge/NativeMessagingHosts/browse_pilot.json` |
| Windows | `HKCU\Software\Microsoft\Edge\NativeMessagingHosts\browse_pilot` |

---

## 驗證安裝

```bash
bp doctor
```

`bp doctor` 會回報：

- 偵測到的瀏覽器類型
- Extension 連線狀態（WebSocket / Native Messaging）
- NM Host 清單是否存在
- 通訊延遲（round-trip time）

輸出範例：

```
瀏覽器：Firefox 124.0
通訊方式：WebSocket (ws://localhost:9222)
Extension：已連線
延遲：12ms
狀態：正常
```

若有任何錯誤，`bp doctor` 會輸出具體的修復建議。
