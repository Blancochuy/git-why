# git-why

**Understand why code exists, not just who wrote it.**

`git-why` answers the question every developer asks when reading legacy code:  
*"Why does this line exist?"*

It combines `git log`, `git blame`, and semantic search to present historical context for any line of code in a readable and actionable format.

## Why not `git blame`?

`git blame` only shows **who** wrote something and **when**.  
`git-why` explains **why** it was written.

```bash
$ git blame src/auth/middleware.ts:142
# Shows: hash, author, date (cryptic and unhelpful)

$ git-why src/auth/middleware.ts:142
Line 142 — Added: 2024-08-14 by @carlos
Message: feat: add JWT refresh token rotation

PR #341 — "Security: prevent token reuse after logout"
Issue #89 — "Users getting logged out randomly after 1h"

Context: This check was added after a reported session hijacking vulnerability.
```

## Installation

### Homebrew (macOS/Linux)
```bash
brew install chuy/tap/git-why
```

### Go install
```bash
go install github.com/chuy/git-why@latest
```

### Binary download
Download from [GitHub Releases](https://github.com/chuy/git-why/releases)

## Usage

### Basic: Single line
```bash
git-why src/auth.ts:142
```

### Range of lines
```bash
git-why src/auth.ts:140-155
```

### Compact output
```bash
git-why src/auth.ts:142 --short
a3f92c1 · 2024-08-14 · @carlos · feat: JWT refresh token rotation
```

### More context
```bash
git-why src/auth.ts:142 --context 5
```

### Output formats
```bash
git-why src/auth.ts:142 --json    # JSON output
git-why src/auth.ts:142 --md      # Markdown output
git-why src/auth.ts:142 --plain   # No colors (for pipes)
git-why src/auth.ts:142 --llm     # LLM-optimized "Context Pack"
git-why src/auth.ts:142 --include-diff # Include commit patch
```

### Statistics
```bash
git-why stats src/auth.ts
```

## Roadmap / Features

### Core Features (MVP)
- [x] Single line blame with full commit context
- [x] Range of lines history
- [x] Compact one-line output (`--short`)
- [x] Configurable context lines (`--context`)
- [x] Multiple output formats (JSON, Markdown, plain)

### Advanced Features
- [x] Function/block detection (`--fn`) — Go, TS/JS, Python, Java, Rust
- [x] Search in commit history (`--search`) — Text + regex support
- [x] AI context optimization (`--llm`) — Optimized for Agents/LLMs
- [ ] Watch mode for editor integration (`watch`)
- [ ] AI-powered summaries (local via Ollama) — `--ai`
- [x] Git diff extraction (`--include-diff`)
- [ ] GitHub/GitLab PR & Issue integration — Auto-link commits to PRs
- [ ] Shell completions (Bash, Zsh, Fish)
- [ ] Author statistics and heatmap (`stats`)

## Building from source

```bash
git clone https://github.com/chuy/git-why.git
cd git-why
go build -o git-why .
```

## Development

```bash
# Run tests
go test ./...

# Build
go build -o git-why .

# Install locally
go install .
```

## AI Agent Integration (MCP)

`git-why` can be used as a native tool by AI agents (Claude Desktop, Cursor, etc.) via the Model Context Protocol.

### Configuration

Add this to your MCP settings (e.g., `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "git-why": {
      "command": "git-why",
      "args": ["mcp"]
    }
  }
}
```

### Tools exposed:
- `git-why-context`: Explains the context of a line or range.
- `git-why-stats`: Shows author and history statistics for a file.

## License

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
