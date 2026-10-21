# Contributing to git-why

Thanks for contributing.

## Prerequisites

- Go 1.23 or newer
- Git 2.20 or newer

## Local Setup

```bash
git clone https://github.com/Blancochuy/git-why.git
cd git-why
go mod download
go test ./...
go build -o git-why .
```

## Development Workflow

1. Create a branch from the default branch. That branch is `feature/ai-integration` until it is renamed to `main`.
2. Add or update tests for your change.
3. Run local checks.
4. Open a pull request.

## Quality Checks

CI runs these on Ubuntu:

```bash
go test -race ./...
go vet ./...
go build -o git-why .
```

## Commit Convention

This project uses Conventional Commits:

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation change
- `test:` tests only
- `refactor:` internal code change without behavior change
- `chore:` tooling/maintenance

Example:

```text
feat(search): add regex flag for history term matching
```

## Pull Request Guidelines

- Keep PR scope focused.
- Update docs if behavior changes.
- Add tests for new behavior.
- Ensure CI is green.

## Project Structure

```text
git-why/
  cmd/         # Cobra commands
  internal/    # Internal packages (git, render, mcp, cache, enricher)
  .github/     # CI workflows and templates
```

## Code of Conduct

By participating, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

By contributing, you agree that your contributions are licensed under MIT.
