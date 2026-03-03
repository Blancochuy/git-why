# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-02

### Added
- Initial release of git-why
- **RF-01**: Single line blame with full commit context
- **RF-02**: Range of lines history
- **RF-03**: Function/block detection (`--fn`) for Go, TS/JS, Python, Java, Rust
- **RF-04**: Compact one-line output (`--short`)
- **RF-05**: Search in commit history (`--search`) with regex support
- **RF-08**: Watch mode for editor integration
- **RF-09**: Multiple output formats (JSON, Markdown, plain)
- **RF-10**: File statistics with author breakdown and activity heatmap
- GitHub Actions CI/CD pipeline
- goreleaser configuration for multi-platform releases
- Shell completions (bash, zsh, fish, powershell)

### Features
- `git-why <file>:<line>` - Show why a line exists
- `git-why <file>:<start>-<end>` - Show history for a range
- `git-why <file> --fn <function>` - Show function history
- `git-why search <file> --term <text>` - Search commit history
- `git-why stats <file>` - Show file statistics
- `git-why watch <file>` - Interactive mode for editors

### Technical
- Go 1.22+ support
- Cross-platform: macOS, Linux, Windows
- Zero external runtime dependencies
- < 500ms response time for typical queries
- Local caching for API responses
