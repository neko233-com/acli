# Contributing to acli

Thank you for your interest in contributing to unicli!

## How to Contribute

### Reporting Bugs

- Search existing issues before creating a new one
- Use a clear, descriptive title
- Include steps to reproduce the issue
- Mention your OS, Go version, and acli version

### Suggesting Features

- Search existing issues before suggesting
- Explain the use case and motivation
- Provide examples of expected behavior

### Pull Requests

1. **Fork** the repository
2. **Clone** your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/unicli.git
   cd acli
   ```
3. **Create a branch**:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```
4. **Make your changes**
5. **Test** your changes:
   ```bash
   go test -v ./...
   ```
6. **Commit** using conventional commits:
   ```bash
   git commit -m "feat: add new command"
   git commit -m "fix: resolve issue"
   git commit -m "docs: update README"
   git commit -m "refactor: improve code"
   git commit -m "test: add tests for feature"
   ```
7. **Push** and create a Pull Request

## Commit Message Format

```
<type>(<scope>): <description>

[optional body]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting)
- `refactor` - Code refactoring
- `test` - Adding tests
- `chore` - Build process or tool changes

**Examples:**
```
feat(scan): add port scanning feature
fix(ip): resolve IPv6 display issue
docs: update installation instructions
```

## Code Style

- Run `go fmt ./...` before committing
- Run `goimports -l -w .` to organize imports
- Follow standard Go naming conventions
- Add comments for non-obvious code

## Project Structure

```
unicli/
├── cmd/              # Command implementations
│   ├── root.go      # Root command
│   ├── commands.go  # Command registration
│   └── *.go         # Individual commands
├── scripts/          # Build and install scripts
├── docs/             # Documentation
└── .github/
    └── workflows/    # GitHub Actions
```

## Questions?

- Open an issue for discussion
- Check the existing documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.