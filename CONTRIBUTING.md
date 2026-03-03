# Contributing to git-why

Thank you for your interest in contributing to git-why!

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git 2.20 or later
- Make (optional, for build tasks)

### Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/git-why.git
   cd git-why
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Build the project:
   ```bash
   go build -o git-why .
   ```

5. Run tests:
   ```bash
   go test ./... -v
   ```

## Development Workflow

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/git -v
```

### Linting

We use golangci-lint for code quality:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Building

```bash
# Development build
go build -o git-why .

# Production build (optimized)
go build -ldflags="-s -w" -o git-why .
```

## Project Structure

```
git-why/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and global flags
│   ├── why.go             # Main why command
│   ├── stats.go           # Statistics command
│   ├── search.go          # Search command
│   └── watch.go           # Watch mode command
├── internal/
│   ├── git/               # Git operations wrapper
│   ├── enricher/          # GitHub/GitLab API clients
│   ├── cache/             # Local caching
│   └── render/            # Terminal output formatting
├── .github/
│   └── workflows/         # GitHub Actions CI/CD
└── test/
    └── fixtures/          # Test repositories
```

## Adding a New Feature

1. Create a new branch:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. Write tests first (TDD approach)

3. Implement the feature

4. Ensure all tests pass:
   ```bash
   go test ./...
   ```

5. Update documentation if needed

6. Submit a pull request

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Add comments for exported functions
- Write meaningful commit messages (Conventional Commits)

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
