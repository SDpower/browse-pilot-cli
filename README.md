# browse-pilot-cli (bp)

Cross-browser automation CLI that controls Firefox, Chrome, and Edge via WebExtension API — no CDP required.

## Features

- 🌐 Supports Firefox, Chrome, and Edge
- 🔒 Uses your real browser profile (cookies, login state, history preserved)
- 🚫 No CDP dependency — operates through WebExtension API with strong anti-detection
- 🤖 Native MCP support — integrates directly with Claude Code
- 🐍 Python session — write automation scripts with a `browser` object
- ⚡ Single Go binary — serves as CLI, WebSocket server, Native Messaging host, and MCP server

> 📖 **繁體中文版**: [README_ZH_TW.md](README_ZH_TW.md)

## Installation

### Prerequisites

- Go 1.22+
- Node.js 18+ (for extension build & lint)
- Firefox 109+ / Chrome 110+ / Edge 110+

### Build from Source

```bash
# Install via go install
go install github.com/SDpower/browse-pilot-cli/cmd/bp@latest

# Or build from source
git clone https://github.com/SDpower/browse-pilot-cli.git
cd browse-pilot-cli
make build
```

### Extension Installation

Build the extensions first:

```bash
bash scripts/build-extensions.sh
```

#### Firefox

1. Open `about:debugging`
2. Click "This Firefox"
3. Click "Load Temporary Add-on"
4. Select `dist/firefox/manifest.json`

#### Chrome

1. Open `chrome://extensions`
2. Enable "Developer mode" (top right)
3. Click "Load unpacked"
4. Select the `dist/chrome/` directory

#### Edge

1. Open `edge://extensions`
2. Enable "Developer mode" (left sidebar)
3. Click "Load unpacked"
4. Select the `dist/edge/` directory

### Native Messaging Host Setup

Chrome and Edge communicate via Native Messaging, which requires host manifest installation:

```bash
bp setup firefox
bp setup chrome
bp setup edge

# Or set up all browsers at once
bp setup --all
```

## Quick Start

### Check Environment

```bash
bp doctor
```

### Basic Operations

```bash
# Open a webpage
bp open https://example.com

# Get page state (list all interactive elements)
bp state

# Click an element (by index)
bp click 0

# Type text into a specific field
bp input 1 "hello world"

# Take a screenshot
bp screenshot output.png
```

### Wait & Extract

```bash
# Wait for an element
bp wait selector "table.result"

# Wait for text to appear
bp wait text "Loading complete"

# Execute JavaScript
bp eval "document.querySelectorAll('tr').length"

# Get page information
bp get title
bp get html --selector "table"
bp get text 2
```

### Python Automation

```bash
# Execute Python code (with access to browser object)
bp python "result = browser.state(); print(len(result['elements']))"

# Execute a Python script
bp python --file script.py

# List session variables
bp python --vars

# Reset session
bp python --reset
```

## Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--browser` | Target browser (firefox / chrome / edge / auto) | auto |
| `--port` | WebSocket port | 9222 |
| `--json` | JSON output format | false |
| `--timeout` | Timeout in milliseconds | 30000 |
| `--verbose` | Verbose logging | false |
| `--mcp` | Run as MCP server | false |
| `--session` | Session name | default |
| `--native-messaging` | Run as NM host | false |

### Navigation

| Command | Description |
|---------|-------------|
| `bp open <url>` | Navigate to URL |
| `bp back` | Go back |
| `bp forward` | Go forward |
| `bp reload` | Reload page |
| `bp scroll <up\|down>` | Scroll page (optional `--amount <px>`) |

### Page Inspection

| Command | Description |
|---------|-------------|
| `bp state` | List current URL, title, and all interactive elements |
| `bp screenshot [path]` | Take screenshot (outputs base64 if no path; `--full` for full page) |

### Interaction

| Command | Description |
|---------|-------------|
| `bp click <index>` | Click element (also `bp click <x> <y>` for coordinate click) |
| `bp dblclick <index>` | Double-click element |
| `bp rightclick <index>` | Right-click element |
| `bp hover <index>` | Hover over element |
| `bp type <text>` | Type text (focused element) |
| `bp input <index> <text>` | Click element and type text |
| `bp keys <keys>` | Send keyboard events (e.g., `Enter`, `Ctrl+a`) |
| `bp select <index> <value>` | Select dropdown option |
| `bp upload <index> <path>` | Upload file to file input |

### Tab Management

| Command | Description |
|---------|-------------|
| `bp tabs` | List all tabs |
| `bp tab <index>` | Switch to tab |
| `bp close-tab [index]` | Close tab (defaults to current) |

### Cookies

| Command | Description |
|---------|-------------|
| `bp cookies get [--url <url>]` | Get cookies |
| `bp cookies set <name> <value>` | Set cookie (`--domain`, `--secure`, `--same-site`) |
| `bp cookies clear [--url <url>]` | Clear cookies |
| `bp cookies export <file>` | Export cookies to JSON file |
| `bp cookies import <file>` | Import cookies from JSON file |

### Wait

| Command | Description |
|---------|-------------|
| `bp wait selector <css>` | Wait for element (`--hidden` to wait for removal) |
| `bp wait text <text>` | Wait for text to appear |
| `bp wait url <pattern>` | Wait for URL to match pattern |

All wait commands support `--timeout <ms>` (default 30000).

### Data Extraction

| Command | Description |
|---------|-------------|
| `bp get title` | Get page title |
| `bp get html [--selector <css>]` | Get HTML content |
| `bp get text <index>` | Get element text |
| `bp get value <index>` | Get form field value |
| `bp get attributes <index>` | Get all element attributes |
| `bp get bbox <index>` | Get element bounding box |

### Execution

| Command | Description |
|---------|-------------|
| `bp eval <code>` | Execute JavaScript in page context |
| `bp python <code>` | Execute Python (with `browser` object access) |
| `bp python --file <path>` | Execute Python script |
| `bp python --vars` | List session variables |
| `bp python --reset` | Reset Python session |

### System

| Command | Description |
|---------|-------------|
| `bp doctor` | Check browser and connection status |
| `bp status` | Show current connection info |
| `bp sessions` | List active sessions |
| `bp close [--all]` | Close connection |
| `bp setup <browser>` | Install Native Messaging Host (`--all` for all browsers) |

## MCP Integration (Claude Code)

`bp` supports the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) for direct integration with Claude Code and other AI agents.

### Configuration

Add to `.claude/mcp.json` or `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "browse-pilot": {
      "command": "bp",
      "args": ["--mcp"],
      "env": {
        "BP_BROWSER": "firefox",
        "BP_PORT": "9222"
      }
    }
  }
}
```

### Available MCP Tools

| Tool | Description |
|------|-------------|
| `bp_navigate` | Navigate to URL |
| `bp_state` | Get page state and interactive elements |
| `bp_click` | Click element |
| `bp_input` | Type text into element |
| `bp_type` | Type text (focused element) |
| `bp_screenshot` | Take screenshot |
| `bp_eval` | Execute JavaScript |
| `bp_scroll` | Scroll page |
| `bp_keys` | Send keyboard events |
| `bp_wait` | Wait for condition |
| `bp_get` | Get element information |
| `bp_tabs` | Manage tabs |
| `bp_cookies` | Manage cookies |
| `bp_upload` | Upload file |
| `bp_hover` | Hover over element |
| `bp_dblclick` | Double-click |
| `bp_rightclick` | Right-click |
| `bp_select` | Select dropdown option |
| `bp_back` / `bp_forward` / `bp_reload` | Navigation controls |

## Architecture

```
Claude Code / AI Agent
    │ MCP protocol (stdio)
    ▼
browse-pilot-cli (Go binary)
    ├── WebSocket Server ──→ Firefox Extension (MV2, persistent background)
    └── Native Messaging ──→ Chrome/Edge Extension (MV3, service worker)
                                    │
                                    ▼
                              Content Script → Target Web Page
```

### Dual-Channel Communication

| Browser | Channel | Reason |
|---------|---------|--------|
| Firefox | WebSocket | MV2 persistent background supports long-lived connections |
| Chrome | Native Messaging | MV3 service worker has idle timeout limitations |
| Edge | Native Messaging | Same as Chrome (Chromium-based) |

Both channels use the same JSON-RPC 2.0 message format. Upper-level command logic is channel-agnostic.

## Development

```bash
# Build Go binary
make build

# Run tests
make test

# Run lint
make lint

# Build extensions
make extension-build

# Extension lint
make extension-lint
```

## Documentation

- [Installation Guide](docs/INSTALL.md)
- [Command Reference](docs/COMMANDS.md)
- [Communication Protocol](docs/PROTOCOL.md)
- [Cross-Browser Compatibility](docs/BROWSERS.md)
- [MCP Integration Guide](docs/MCP.md)
- [Usage Examples](docs/EXAMPLES.md)

## License

MIT © [@SteveLuo](https://github.com/sdpower)
