# git-why

[![CI](https://github.com/Blancochuy/git-why/actions/workflows/ci.yml/badge.svg)](https://github.com/Blancochuy/git-why/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Understand why code exists, not just who wrote it.

`git-why` adds historical context to `git blame`: the commit that introduced a line, the message and body that explain it, and the commits that touched a range or a function. It also runs as an MCP server so an agent can ask the same questions.

It needs Git on `PATH`.

## Demo

These commands were run from this repository with a locally built binary (`go build -o git-why .`). Output below is copied from that run. Dates use the local timezone of that machine, which was UTC.

### Line

```bash
git-why internal/render/ai_renderer.go:20 --plain
```

```text

Line 20 - 1e08ea65b536 - Added: 2026-03-03 by @jesus

Message:
  chore: optimize AI output for token efficiency and clean up diff headers
```

### Function

```bash
git-why internal/render/ai_renderer.go --fn GetContext --plain
```

```text
Function GetContext() in internal/render/ai_renderer.go (lines 20-44):

Commits that modified internal/render/ai_renderer.go:20-44:

  ea3c55e 2026-03-04 - Chuy
    feat: Implement Git blame and commit retrieval, integrate GitHub API for commit enrichment, and add project badges to README.

  1e08ea6 2026-03-03 - jesus
    chore: optimize AI output for token efficiency and clean up diff headers

  14ffd42 2026-03-03 - jesus
    feat: implement AI context optimized flags and MCP server
```

### MCP

`git-why mcp` speaks JSON-RPC on stdin/stdout. One request per line:

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"git-why-context","arguments":{"file":"internal/render/ai_renderer.go","line":20}}}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"git-why-stats","arguments":{"file":"internal/render/ai_renderer.go"}}}' \
  | git-why mcp
```

`tools/list` responded with this object (pretty-printed here; on the wire it is a single line):

```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "tools": [
            {
                "name": "git-why-context",
                "description": "Explain the historical intent of a line or range of lines using Git context.",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "file": {
                            "type": "string",
                            "description": "Path to the file (relative to the repository root)"
                        },
                        "line": {
                            "type": "integer",
                            "description": "Specific line number (1-indexed)"
                        },
                        "range": {
                            "type": "string",
                            "description": "Line range such as 10-20. When set, this is used instead of line."
                        }
                    },
                    "required": [
                        "file"
                    ]
                }
            },
            {
                "name": "git-why-stats",
                "description": "Show authorship and historical change statistics for a specific file.",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "file": {
                            "type": "string",
                            "description": "Path to the file"
                        }
                    },
                    "required": [
                        "file"
                    ]
                }
            }
        ]
    }
}
```

`git-why-context` for that line started with:

```text
ID: 1e08ea65b536 | 2026-03-03 | @jesus
FILE: internal/render/ai_renderer.go:20
MSG: chore: optimize AI output for token efficiency and clean up diff headers

--- DIFF ---
```

The rest of that result is the commit diff (187 lines, including a `git-why.exe` binary delta from that commit). It is omitted here.

`git-why-stats` returned:

```text
Stats for internal/render/ai_renderer.go
Total Lines: 76
Total Commits: 4

Contributors:
- jesus: 51 lines (67.1%)
- Chuy: 25 lines (32.9%)
```

Point an MCP client at the binary:

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

## Features

- Line context: `git-why <file>:<line>`
- Range history: `git-why <file>:<start>-<end>`
- Function history: `git-why <file> --fn <function>` (Go, TypeScript/JavaScript, Python, Java, Rust)
- Commit search: `git-why search <file> --term <text>` (subject and body; not the diff)
- File statistics: `git-why stats <file>`
- Output formats: `--json`, `--md`, `--plain`, `--short`, `--llm`
- Optional patch on a single line: `--include-diff` (also included by `--llm`)
- MCP server: `git-why mcp`

`--json`, `--md`, and `--llm` work for a line and for a range. `--llm` on a range lists the commits instead of inlining a diff.

## Installation

There is no tagged release yet, so there are no published binaries and `brew install` does not work.

### Go install

The module path is lowercase and must match `go.mod`. The GitHub owner is `Blancochuy`; `go install` with that capital B fails because Go module paths are case-sensitive.

```bash
go install github.com/blancochuy/git-why@latest
```

With no semver tags, `@latest` is the tip of the default branch (`feature/ai-integration` until it is renamed to `main`).

### From a clone

```bash
git clone https://github.com/Blancochuy/git-why.git
cd git-why
go install .
```

### Binaries and Homebrew, after the first tag

Tagging `v0.1.0` (or any `v*`) runs [`.github/workflows/release.yml`](.github/workflows/release.yml). GoReleaser builds linux, darwin, and windows archives (amd64 and arm64) and attaches them to the GitHub release. The config is checked against GoReleaser v2.18.3.

Homebrew (`brew install blancochuy/tap/git-why`) also needs:

1. A repository `Blancochuy/homebrew-tap` (it does not exist yet).
2. A `HOMEBREW_TAP_GITHUB_TOKEN` secret that can push to that repository.

Until that secret is set, the release workflow still publishes the binaries and skips the tap. Do not run the `brew` command before then.

## Quick start

```bash
git-why internal/render/ai_renderer.go:20
git-why internal/render/ai_renderer.go:20-44
git-why internal/render/ai_renderer.go --fn GetContext
git-why search internal/render/ai_renderer.go --term "token"
git-why stats internal/render/ai_renderer.go
```

## Output

```bash
git-why internal/render/ai_renderer.go:20 --short
git-why internal/render/ai_renderer.go:20 --json
git-why internal/render/ai_renderer.go:20 --md
git-why internal/render/ai_renderer.go:20 --plain
git-why internal/render/ai_renderer.go:20 --llm --include-diff
```

Shell completions are the standard Cobra ones: `git-why completion bash` (also `zsh`, `fish`, `powershell`).

## MCP tools

- `git-why-context`: commit context for a line, or the commits in a `range` such as `10-20`. If both are set, `range` is used. When `origin` is a GitHub remote, a linked pull request is included when the API returns one.
- `git-why-stats`: authorship and commit counts for a file.

GitHub enrichment uses `GITHUB_TOKEN` or `GITWHY_TOKEN` when set. The GitLab client in `internal/enricher` is not wired up yet.

## Development

```bash
go test -race ./...
go vet ./...
go build -o git-why .
```

CI runs those three commands on Ubuntu for pushes and pull requests to `main` and `feature/ai-integration`.

## Project status

Not released. The first tag will be `v0.1.0`. The default branch is `feature/ai-integration`.

`git-why --version` prints `dev` until a release build sets the version with `-ldflags`.

## Roadmap

- Editor integrations for watch mode
- Call the GitLab enricher the same way GitHub pull requests are attached
- Publish the Homebrew cask once the tap exists

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
