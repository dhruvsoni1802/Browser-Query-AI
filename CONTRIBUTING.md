# Contributing to Browser Query AI

Thank you for your interest in contributing to Browser Query AI! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Submitting Changes](#submitting-changes)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow. Please be respectful, inclusive, and professional in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
```bash
   git clone https://github.com/YOUR_USERNAME/browser-query-ai.git
   cd browser-query-ai
```
3. **Add upstream remote**:
```bash
   git remote add upstream https://github.com/dhruvsoni1802/browser-query-ai.git
```

## Development Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://go.dev/doc/install)
- **Chromium or Google Chrome** - Required for browser automation
  - macOS: `brew install chromium`
  - Linux: `sudo apt install chromium-browser`

### Install Dependencies
```bash
go mod download
```

### Environment Variables

Create a `.env` file (optional):
```bash
# Development environment
ENV=development

# Browser configuration (optional overrides)
CHROMIUM_PATH=/path/to/chromium
SERVER_PORT=8080
MAX_BROWSERS=5
```

### Running the Service
```bash
# Development mode
go run cmd/server/main.go

# With custom environment
ENV=production go run cmd/server/main.go
```

## Project Structure
```
browser-query-ai/
├── cmd/
│   └── server/           # Application entry point
│       └── main.go
├── internal/
│   ├── browser/          # Browser process management
│   │   ├── process.go    # Process lifecycle
│   │   └── ports.go      # Port allocation pool
│   ├── cdp/              # Chrome DevTools Protocol client
│   │   ├── client.go     # WebSocket client
│   │   ├── client_helpers.go
│   │   ├── commands.go   # High-level CDP commands
│   │   ├── discovery.go  # Browser target discovery
│   │   └── types.go      # Type definitions
│   └── config/           # Configuration management
│       ├── config.go     # Config loading
│       └── helpers.go    # Config utilities
├── CONTRIBUTING.md
├── SECURITY.md
├── README.md
└── go.mod
```

## Making Changes

### Branch Naming Convention

Use descriptive branch names:
- `feature/add-screenshot-api` - New features
- `fix/port-allocation-race` - Bug fixes
- `docs/update-readme` - Documentation
- `refactor/cleanup-cdp-client` - Code refactoring

### Workflow

1. **Create a branch**:
```bash
   git checkout -b feature/your-feature-name
```

2. **Make your changes** following our coding standards

3. **Commit your changes**:
```bash
   git add .
   git commit -m "feat: add screenshot capture functionality"
```

4. **Keep your branch updated**:
```bash
   git fetch upstream
   git rebase upstream/main
```

5. **Push to your fork**:
```bash
   git push origin feature/your-feature-name
```

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` to format code:
```bash
  gofmt -w .
```
- Run linter before submitting:
```bash
  golangci-lint run
```

### Code Organization

- **Package-level comments**: Every package should have a doc comment
- **Exported functions**: Must have documentation comments
- **Error handling**: Always wrap errors with context using `fmt.Errorf("context: %w", err)`
- **Logging**: Use structured logging with `slog`
```go
  slog.Info("operation completed", "key", value)
  slog.Error("operation failed", "error", err)
```

### Naming Conventions

- **Variables**: camelCase for local, PascalCase for exported
- **Constants**: PascalCase or ALL_CAPS for package-level
- **Interfaces**: Descriptive names ending in `-er` when possible (e.g., `Handler`, `Manager`)
- **Files**: lowercase with underscores (e.g., `client_helpers.go`)

### Example Code Style
```go
// CreateBrowserContext creates a new isolated browser context.
// It returns the context ID or an error if creation fails.
func (c *Client) CreateBrowserContext() (string, error) {
    result, err := c.SendCommand("Target.createBrowserContext", nil)
    if err != nil {
        return "", fmt.Errorf("failed to create browser context: %w", err)
    }

    var response struct {
        BrowserContextID string `json:"browserContextId"`
    }

    if err := json.Unmarshal(result, &response); err != nil {
        return "", fmt.Errorf("failed to parse response: %w", err)
    }

    return response.BrowserContextID, nil
}
```

## Testing Guidelines

### Writing Tests

- Place tests in `*_test.go` files alongside source code
- Use table-driven tests where appropriate
- Mock external dependencies (WebSocket, HTTP, browser processes)

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/cdp/...
```

### Test Example
```go
func TestCreateBrowserContext(t *testing.T) {
    tests := []struct {
        name    string
        mock    func(*MockClient)
        want    string
        wantErr bool
    }{
        {
            name: "successful creation",
            mock: func(m *MockClient) {
                m.EXPECT().SendCommand("Target.createBrowserContext", nil).
                    Return([]byte(`{"browserContextId":"ctx123"}`), nil)
            },
            want:    "ctx123",
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Update documentation** if you changed APIs or added features
2. **Add tests** for new functionality
3. **Ensure all tests pass**: `go test ./...`
4. **Update CHANGELOG.md** with your changes (if applicable)
5. **Create a pull request** with a clear title and description

### Pull Request Template

When creating a PR, include:
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing performed
- [ ] All tests passing

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings introduced
```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(cdp): add screenshot capture command

Implements Page.captureScreenshot CDP command with support for
full page screenshots and custom viewport dimensions.

Closes #123
```
```
fix(browser): prevent port pool exhaustion

Added validation to ensure ports are returned even when process
crashes unexpectedly.

Fixes #456
```

## Reporting Bugs

### Before Submitting

- Check existing issues to avoid duplicates
- Verify the bug exists in the latest version
- Collect relevant information (OS, Go version, error logs)

### Bug Report Template
```markdown
## Bug Description
Clear and concise description of the bug

## Steps to Reproduce
1. Start service with '...'
2. Send request to '...'
3. Observe error '...'

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: macOS 14.2
- Go Version: 1.21.5
- Chrome Version: 144.0.7559.133

## Logs
```
Paste relevant logs here
```

## Additional Context
Screenshots, configuration files, etc.
```

## Feature Requests

We welcome feature requests! Please:

1. **Check existing issues** to see if already requested
2. **Describe the problem** you're trying to solve
3. **Propose a solution** if you have one in mind
4. **Explain use cases** and benefits

### Feature Request Template
```markdown
## Feature Description
Clear description of the feature

## Problem Statement
What problem does this solve?

## Proposed Solution
How should this work?

## Alternatives Considered
Other approaches you've thought about

## Additional Context
Examples, mockups, references
```

## Questions?

- **General questions**: Open a [Discussion](https://github.com/dhruvsoni1802/browser-query-ai/discussions)
- **Bug reports**: Open an [Issue](https://github.com/dhruvsoni1802/browser-query-ai/issues)
- **Security concerns**: See [SECURITY.md](SECURITY.md)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

---

Thank you for contributing to Browser Query AI! 🚀