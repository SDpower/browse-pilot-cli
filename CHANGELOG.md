# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.3] - 2026-03-23

### Added

- **i18n**: 12 languages support (en, de, es, fr, hi, id, it, ja, ko, pt-BR, zh-Hans, zh-Hant) with auto system locale detection (macOS AppleLocale / LANG)
- **Plugin Marketplace**: Claude Code plugin marketplace support — install via `/plugin marketplace add SDpower/browse-pilot-cli`

## [0.1.2] - 2026-03-23

### Added

- **CI**: GitHub Actions CI workflow (Go build/test/vet, golangci-lint, ESLint, Extension build)
- **CD**: Release workflow — push tag auto-builds multi-platform binaries (linux/darwin/windows × amd64/arm64) + Extension archives

### Fixed

- golangci-lint-action v6 → v7 for lint v2 support
- `package-lock.json` added to version control for CI `npm ci`

## [0.1.1] - 2026-03-23

### Changed

- **CLI binary renamed**: `bp` → `bp_cli` (all docs, examples, configs updated)

### Added

- **LICENSE**: MIT license file
- **CHANGELOG.md**: version history
- **.golangci.yml**: golangci-lint v2 configuration
- All golangci-lint issues fixed (octalLiteral, paramTypeCombine, bodyclose, errcheck, etc.)

## [0.1.0] - 2026-03-23

### Added

- **CLI (`bp_cli`)**: 33 commands covering navigation, inspection, interaction, tabs, cookies, wait, data extraction, eval, python, and system operations
- **Transport layer**: WebSocket server (Firefox) + Native Messaging host (Chrome/Edge) with unified JSON-RPC 2.0 protocol
- **Browser auto-detection**: automatically selects the best available browser and transport
- **WebExtension**: Firefox MV2 + Chrome/Edge MV3 with 100% shared content scripts
- **MCP server**: 20 tools + 2 resources for Claude Code integration via `bp_cli --mcp` flag
- **Python session**: persistent subprocess with `browser` object for scripting automation
- **Output formatting**: human-readable and JSON dual-mode output (`--json`)
- **Full page screenshot**: segment-based capture with scroll stitching
- **Cookie management**: get/set/clear/export/import
- **Native Messaging host setup**: `bp_cli setup firefox/chrome/edge/--all`
- **Extension build pipeline**: `scripts/build-extensions.sh` produces dist/firefox, dist/chrome, dist/edge
- **Documentation**: README (English + 繁體中文), 6 docs (INSTALL, COMMANDS, PROTOCOL, BROWSERS, MCP, EXAMPLES)
- **SKILL.md**: Claude Code skill reference for MCP integration

[0.1.3]: https://github.com/SDpower/browse-pilot-cli/releases/tag/v0.1.3
[0.1.2]: https://github.com/SDpower/browse-pilot-cli/releases/tag/v0.1.2
[0.1.1]: https://github.com/SDpower/browse-pilot-cli/releases/tag/v0.1.1
[0.1.0]: https://github.com/SDpower/browse-pilot-cli/releases/tag/v0.1.0
