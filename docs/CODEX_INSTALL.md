# Codex Installation and Usage Guide

This guide explains how to use Browse Pilot with the Codex Plugin to control local Firefox. The Plugin starts the local STDIO MCP Server automatically; you do not need to run the MCP command manually during normal use.

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
browse-pilot  bp_cli  --mcp --browser firefox --port 9222 --timeout 60000
```

## Start using it

Open a new Codex session after installation. The Plugin starts the MCP Server automatically, so do not start the MCP command separately. Once the Firefox Extension is loaded, ask for browser operations in natural language.

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

Port `9222` is occupied by another process. Stop the old MCP/CLI process using that port, or keep only one Browse Pilot MCP configuration, then open a new Codex session.

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

## Manual MCP fallback

Add a manual MCP Server only when you are not using the Plugin. Do not use it with the Plugin, because two processes would compete for port `9222`.

```bash
codex mcp add browse-pilot -- bp_cli --mcp --browser firefox --port 9222 --timeout 60000
```

To remove this manual configuration:

```bash
codex mcp remove browse-pilot
```

## Update and remove

Rebuild the CLI and Extension before updating:

```bash
go build -o ~/.local/bin/bp_cli ./cmd/bp/
bash scripts/build-extensions.sh
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
