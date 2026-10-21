# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

No version has been tagged yet. The section below is the planned first release.
Tagging `v0.1.0` publishes the GitHub release.

## [Unreleased]

### Added
- Initial git-why CLI
- Single line blame with commit message and body
- Range history for `<file>:<start>-<end>`
- Function history (`--fn`) for Go, TypeScript/JavaScript, Python, Java, and Rust
- Compact output (`--short`)
- Commit message search: `git-why search <file> --term <text>` (text or regex)
- Watch mode for editor integration
- JSON, Markdown, plain, and LLM-oriented output
- File statistics with author breakdown and monthly activity
- MCP server (`git-why mcp`) with `git-why-context` and `git-why-stats`
- Optional GitHub pull request enrichment when `origin` is a GitHub remote
- GitHub Actions CI on `main` and `feature/ai-integration`
- GoReleaser configuration for release archives
- Shell completions via `git-why completion` (bash, zsh, fish, powershell)

### Notes
- Requires Git on `PATH`
- Go 1.23 or newer
- Release archives target linux, darwin, and windows (amd64 and arm64) once `v0.1.0` is tagged
- GitHub and GitLab API responses are cached under `~/.cache/git-why` (override with `GITWHY_CACHE_DIR`)
- The GitLab client is implemented but not called by the CLI yet
