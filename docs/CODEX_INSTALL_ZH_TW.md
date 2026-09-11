# Codex 安裝與使用指南

本指南說明如何以 Codex Plugin 使用 Browse Pilot 控制本機 Firefox。單一 Streamable HTTP MCP 服務可供多個 Codex 工作階段共用，避免各工作階段重複占用 Firefox WebSocket 埠號。

## 系統需求

- 已安裝可執行的 Go，以及 Codex CLI 或 Codex App。
- 已安裝 Firefox；需能使用 `about:debugging` 暫時載入擴充套件。
- 可從 `PATH` 執行 `bp_cli`。
- 已取得此專案的原始碼，且能找到 `dist/firefox/manifest.json`；以下範例以 `/absolute/path/to/browse-pilot-cli` 代表專案目錄。

## 建置並安裝 CLI

在專案根目錄建置 CLI，並確認 `~/.local/bin` 已在 `PATH` 中：

```bash
go build -o ~/.local/bin/bp_cli ./cmd/bp/
bp_cli --version
```

## 暫時載入 Firefox Extension

1. 先在專案根目錄執行 `bash scripts/build-extensions.sh`。
2. 在 Firefox 開啟 `about:debugging`，選擇「此 Firefox」。
3. 選取「載入暫用附加元件」，然後選擇 `dist/firefox/manifest.json`。
4. Firefox 重啟後，暫時載入的 Extension 會卸載；請回到同一頁重新載入 `dist/firefox/manifest.json`。

Firefox 使用 WebSocket，不需要執行 `bp_cli setup firefox`。

## 啟動共用 MCP 服務

在一個獨立 Terminal 啟動一次本機服務；保持此程序執行，多個 Codex 工作階段即可共用：

```bash
bp_cli --mcp-http --browser firefox --port 9222 --mcp-port 8931 --timeout 60000
```

可用健康檢查確認服務與 Extension 狀態：

```bash
curl http://127.0.0.1:8931/healthz
```

## 從 0.1.x 升級

`0.2.0` 已移除 stdio `--mcp`。升級時先停止所有以 `bp_cli --mcp` 啟動的舊程序，再建置新版 CLI、啟動唯一的共用 HTTP 服務，並重新安裝 Plugin：

```bash
go build -o ~/.local/bin/bp_cli ./cmd/bp/
bp_cli --version
bp_cli --mcp-http --browser firefox --port 9222 --mcp-port 8931 --timeout 60000
codex plugin remove browse-pilot@browse-pilot-marketplace
codex plugin add browse-pilot@browse-pilot-marketplace
```

`bp_cli --version` 應顯示 `0.2.0`。完成後請開啟新的 Codex 工作階段；舊工作階段不會自動重新載入 MCP 設定。

## 安裝 Codex Plugin

請擇一使用本機 marketplace 或 GitHub marketplace；不要同時安裝兩個來源的同名 Plugin。

### 本機 marketplace

```bash
codex plugin marketplace add /absolute/path/to/browse-pilot-cli
codex plugin add browse-pilot@browse-pilot-marketplace
```

### GitHub marketplace

```bash
codex plugin marketplace add SDpower/browse-pilot-cli
codex plugin add browse-pilot@browse-pilot-marketplace
```

## 確認安裝

```bash
codex plugin list
codex mcp list
```

預期會看到：

```text
browse-pilot@browse-pilot-marketplace  installed, enabled
browse-pilot  http://127.0.0.1:8931/mcp
```

## 開始使用

安裝完成後開啟新的 Codex 工作階段。確認共用 MCP 服務與 Firefox Extension 已啟動後，可直接用自然語言要求瀏覽器操作；不需要為每個 Codex 工作階段另開一個 `bp_cli`。

中文範例：

```text
使用 Browse Pilot 開啟這個 Threads 網址，讀取貼文標題與正文。
```

English example:

```text
Use Browse Pilot to open this Threads URL and return the post title and body.
```

## 疑難排解

### `ConnectionError`

確認 Firefox 正在執行、Extension 已在 `about:debugging` 暫時載入，且使用連接埠 `9222`。Extension 中斷後重新載入或重新連線，再重試工具呼叫。

### `TimeoutError`

每次工具呼叫最長等待 60 秒。確認 Extension 與目標頁面仍可回應後再重試；逾時不代表操作成功。

### `address already in use`

若是 `8931`，先以 `curl http://127.0.0.1:8931/healthz` 判斷共用服務是否已在執行；已執行時不應再啟動第二份。若是 `9222`，請停止舊的 stdio MCP／CLI 程序，再啟動唯一的 HTTP MCP 服務。

### `Handler is not defined`

重新建置 Extension，並在 `about:debugging` 重新載入 `dist/firefox/manifest.json`：

```bash
bash scripts/build-extensions.sh
```

### Plugin 更新未生效

GitHub marketplace 先更新快取，再移除並重新安裝 Plugin，最後開啟新的 Codex 工作階段：

```bash
codex plugin marketplace upgrade browse-pilot-marketplace
codex plugin remove browse-pilot@browse-pilot-marketplace
codex plugin add browse-pilot@browse-pilot-marketplace
```

本機 marketplace 則先更新本機原始碼後，執行移除、重新安裝與開啟新工作階段的後兩個步驟。

## 不使用 Plugin 的直接 MCP 設定

若不使用 Plugin，可直接把同一個 HTTP endpoint 加入 Codex。Plugin 與此設定只選一種，避免出現重複工具名稱。

```bash
codex mcp add browse-pilot --url http://127.0.0.1:8931/mcp
```

移除此手動設定：

```bash
codex mcp remove browse-pilot
```

## 更新與移除

更新前先停止共用 MCP 服務，再重新建置 CLI 與 Extension，最後以相同命令重新啟動服務：

```bash
go build -o ~/.local/bin/bp_cli ./cmd/bp/
bash scripts/build-extensions.sh
bp_cli --mcp-http --browser firefox --port 9222 --mcp-port 8931 --timeout 60000
```

開發本機 Plugin 時，將 `plugins/browse-pilot/.codex-plugin/plugin.json` 的 `version` 更新為單一 `<原版本>+codex.<cachebuster>` 後重新安裝，並在 Firefox 對 `dist/firefox/manifest.json` 按「重新載入」：

```bash
codex plugin add browse-pilot@browse-pilot-marketplace
```

更新 GitHub marketplace 時，請使用「Plugin 更新未生效」的三個指令。完全移除 Plugin 與 marketplace 時：

```bash
codex plugin remove browse-pilot@browse-pilot-marketplace
codex plugin marketplace remove browse-pilot-marketplace
```

若使用本機 marketplace，也使用相同的移除指令；移除後可刪除或保留本機專案原始碼。
