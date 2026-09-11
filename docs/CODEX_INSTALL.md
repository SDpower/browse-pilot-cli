# Codex Installation and Usage Guide

This guide explains how to use Browse Pilot with the Codex Plugin to control local Firefox. One Streamable HTTP MCP service can be shared by multiple Codex sessions, avoiding duplicate processes competing for the Firefox WebSocket port.

## System requirements

- A working Go installation and either Codex CLI or the Codex App.
- Firefox, with access to `about:debugging` for temporarily loading an extension.
- `bp_cli` available on `PATH`.
- A checkout of this project with `dist/firefox/manifest.json` available. The examples below use `/absolute/path/to/browse-pilot-cli` as a replaceable project-directory placeholder.

## Build and install the CLI

From the project root, build the CLI and make sure `~/.local/bin` is on `PATH`:

```bash
go build -o ~/.local/bin/bp_cli ./cmd/bp/
bp_cli --version
```

## Temporarily load the Firefox Extension

1. Run `bash scripts/build-extensions.sh` from the project root.
2. Open `about:debugging` in Firefox and choose “This Firefox.”
3. Choose “Load Temporary Add-on” and select `dist/firefox/manifest.json`.
4. Firefox unloads temporary extensions after a restart; return to the same page and load `dist/firefox/manifest.json` again.

Firefox uses WebSocket, so do not run `bp_cli setup firefox`.

## Start the shared MCP service

Start the local service once in a separate terminal and keep it running. Multiple Codex sessions can share it:

```bash
bp_cli --mcp-http --browser firefox --port 9222 --mcp-port 8931 --timeout 60000
```

Use the health endpoint to verify both the service and Extension state:

```bash
curl http://127.0.0.1:8931/healthz
```

## Upgrade from 0.1.x

Version `0.2.0` removes the stdio `--mcp` mode. Stop every old process started with `bp_cli --mcp`, then build the new CLI, start the single shared HTTP service, and reinstall the Plugin:

```bash
go build -o ~/.local/bin/bp_cli ./cmd/bp/
bp_cli --version
bp_cli --mcp-http --browser firefox --port 9222 --mcp-port 8931 --timeout 60000
codex plugin remove browse-pilot@browse-pilot-marketplace
codex plugin add browse-pilot@browse-pilot-marketplace
```

`bp_cli --version` should print `0.2.0`. Open a new Codex session afterward; existing sessions do not automatically reload the MCP configuration.

## Install the Codex Plugin

Choose either the local marketplace or the GitHub marketplace; do not install the same Plugin from both sources at once.

### Local marketplace

```bash
codex plugin marketplace add /absolute/path/to/browse-pilot-cli
codex plugin add browse-pilot@browse-pilot-marketplace
```

### GitHub marketplace

```bash
codex plugin marketplace add SDpower/browse-pilot-cli
codex plugin add browse-pilot@browse-pilot-marketplace
```

## Verify the installation

```bash
codex plugin list
codex mcp list
```

Expect to see:

```text
browse-pilot@browse-pilot-marketplace  installed, enabled
browse-pilot  http://127.0.0.1:8931/mcp
```

## Start using it

Open a new Codex session after installation. Once the shared MCP service and Firefox Extension are running, ask for browser operations in natural language. Do not start one `bp_cli` process per Codex session.

中文範例：

```text
使用 Browse Pilot 開啟這個 Threads 網址，讀取貼文標題與正文。
```

English example:

```text
Use Browse Pilot to open this Threads URL and return the post title and body.
```

## Troubleshooting

### `ConnectionError`

Confirm that Firefox is running, the Extension is temporarily loaded in `about:debugging`, and the Extension uses port `9222`. After the Extension reconnects or is reloaded, retry the tool call.

### `TimeoutError`

Each tool call waits for at most 60 seconds. Confirm that the Extension and target page can respond, then retry; a timeout does not mean that the operation succeeded.

### `address already in use`

For port `8931`, first run `curl http://127.0.0.1:8931/healthz`; if it succeeds, the shared service is already running and a second copy is unnecessary. For port `9222`, stop the old stdio MCP/CLI process, then start the single HTTP MCP service.

### `Handler is not defined`

Rebuild the Extension, then reload `dist/firefox/manifest.json` in `about:debugging`:

```bash
bash scripts/build-extensions.sh
```

### Plugin update does not take effect

For a GitHub marketplace, refresh its cache, remove and reinstall the Plugin, then open a new Codex session:

```bash
codex plugin marketplace upgrade browse-pilot-marketplace
codex plugin remove browse-pilot@browse-pilot-marketplace
codex plugin add browse-pilot@browse-pilot-marketplace
```

For a local marketplace, update the local checkout first, then perform the latter two commands and open a new session.

## Direct MCP configuration without the Plugin

When not using the Plugin, add the same HTTP endpoint directly. Choose either the Plugin or this configuration to avoid duplicate tool names.

```bash
codex mcp add browse-pilot --url http://127.0.0.1:8931/mcp
```

To remove this manual configuration:

```bash
codex mcp remove browse-pilot
```

## Update and remove

Stop the shared MCP service before updating, then rebuild the CLI and Extension and restart the service with the same command:

```bash
go build -o ~/.local/bin/bp_cli ./cmd/bp/
bash scripts/build-extensions.sh
bp_cli --mcp-http --browser firefox --port 9222 --mcp-port 8931 --timeout 60000
```

When developing the local Plugin, update `version` in `plugins/browse-pilot/.codex-plugin/plugin.json` to one `<base-version>+codex.<cachebuster>` suffix, reinstall, then click “Reload” for `dist/firefox/manifest.json` in Firefox:

```bash
codex plugin add browse-pilot@browse-pilot-marketplace
```

To update the GitHub marketplace, use the three commands in “Plugin update does not take effect.” To completely remove the Plugin and marketplace:

```bash
codex plugin remove browse-pilot@browse-pilot-marketplace
codex plugin marketplace remove browse-pilot-marketplace
```

For a local marketplace, use the same removal commands; you may then keep or delete the local project checkout.
