# git-why

Understand why code exists, not just who wrote it.

`git-why` adds historical context to `git blame` by combining blame, commit details, and targeted history queries.
It helps answer:

- Why was this line introduced?
- What change or incident motivated it?
- Who has changed this area over time?

## Features

- Line-level context: `git-why <file>:<line>`
- Range history: `git-why <file>:<start>-<end>`
- Function history: `git-why <file> --fn <function>`
- Commit search: `git-why search <file> --term <text>` (text or regex)
- File statistics: `git-why stats <file>`
- Machine-readable output: `--json`, `--md`, `--plain`, `--llm`
- Optional patch context: `--include-diff`
- MCP mode for AI agents: `git-why mcp`

## Installation

### Homebrew (macOS/Linux)

```bash
brew install blancochuy/tap/git-why
```

### Go install

```bash
go install github.com/blancochuy/git-why@latest
```

### Binary download

Download prebuilt binaries from [GitHub Releases](https://github.com/blancochuy/git-why/releases).

## Quick Start

```bash
# Analyze one line
git-why src/auth.ts:142

# Analyze a range
git-why src/auth.ts:140-155

# Analyze a function
git-why src/auth.ts --fn authenticateUser

# Search history
git-why search src/auth.ts --term "refresh token"

# Show stats
git-why stats src/auth.ts
```

## Output Modes

```bash
git-why src/auth.ts:142 --short
git-why src/auth.ts:142 --json
git-why src/auth.ts:142 --md
git-why src/auth.ts:142 --plain
git-why src/auth.ts:142 --llm --include-diff
```

## MCP Integration

`git-why` can run as an MCP server over stdio.

Example configuration:

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

Exposed tools:

- `git-why-context`: explain the context of a line or range
- `git-why-stats`: show file authorship and activity stats

## Development

```bash
# Run tests
go test ./...

# Build
go build -o git-why .

# Install locally
go install .
```

## Project Status

Current version: `v0.1.0`

This project is early-stage and improving quickly. Breaking changes may happen before `v1.0.0`.

## Roadmap

- Improve watch mode editor integrations
- Add richer PR/Issue enrichment for GitHub and GitLab
- Improve AI summaries for local LLM workflows
- Run multi-model AI/MCP benchmark (with vs without git-why); see [docs/AI_MCP_BENCHMARK_TODO.md](docs/AI_MCP_BENCHMARK_TODO.md)
- Expand shell completion coverage and docs

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## Security

To report vulnerabilities, see [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).


