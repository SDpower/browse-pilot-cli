# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-03-23

### Added

- **CLI (`bp`)**: 33 commands covering navigation, inspection, interaction, tabs, cookies, wait, data extraction, eval, python, and system operations
- **Transport layer**: WebSocket server (Firefox) + Native Messaging host (Chrome/Edge) with unified JSON-RPC 2.0 protocol
- **Browser auto-detection**: automatically selects the best available browser and transport
- **WebExtension**: Firefox MV2 + Chrome/Edge MV3 with 100% shared content scripts
- **MCP server**: 20 tools + 2 resources for Claude Code integration via `--mcp` flag
- **Python session**: persistent subprocess with `browser` object for scripting automation
- **Output formatting**: human-readable and JSON dual-mode output (`--json`)
- **Full page screenshot**: segment-based capture with scroll stitching
- **Cookie management**: get/set/clear/export/import
- **Native Messaging host setup**: `bp setup firefox/chrome/edge/--all`
- **Extension build pipeline**: `scripts/build-extensions.sh` produces dist/firefox, dist/chrome, dist/edge
- **Documentation**: README (English + 繁體中文), 6 docs (INSTALL, COMMANDS, PROTOCOL, BROWSERS, MCP, EXAMPLES)
- **SKILL.md**: Claude Code skill reference for MCP integration

[0.1.0]: https://github.com/SDpower/browse-pilot-cli/releases/tag/v0.1.0
