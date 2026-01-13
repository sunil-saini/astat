# Contributing to astat

First off, thank you for considering contributing to astat! It's people like you that make astat such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** to demonstrate the steps
- **Describe the behavior you observed** and what you expected
- **Include screenshots** if relevant
- **Include your environment details** (OS, Go version, astat version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful**
- **List any similar features** in other tools

### Pull Requests

1. **Fork the repo** and create your branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** if you've added code that should be tested
4. **Ensure the test suite passes** (`go test ./...`)
5. **Format your code** (`go fmt ./...`)
6. **Commit your changes** using clear commit messages
7. **Push to your fork** and submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- AWS credentials configured
- Git

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/astat.git
cd astat

# Install dependencies
go mod download

# Build
go build -o astat main.go

# Run tests
go test ./...
```

### Project Structure

```
astat/
├── cmd/              # Command implementations
│   ├── ec2/         # EC2 commands
│   ├── s3/          # S3 commands
│   ├── lambda/      # Lambda commands
│   └── ...
├── internal/        # Internal packages
│   ├── aws/        # AWS API clients
│   ├── cache/      # Cache management
│   ├── logger/     # Logging utilities
│   └── ...
└── main.go         # Entry point
```

## Coding Standards

### Go Code Style

- Follow standard Go conventions and idioms
- Use `go fmt` for formatting
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise

### Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters
- Reference issues and pull requests after the first line

Example:
```
feat: add support for ECS service listing

- Implement ECS client
- Add cache support for ECS
- Update documentation

Closes #123
```

### Commit Prefixes

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

## Testing

- Write tests for new features
- Ensure existing tests pass
- Aim for good test coverage
- Use table-driven tests where appropriate

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/cache
```

## Documentation

- Update README.md for user-facing changes
- Add inline code comments for complex logic
- Update command help text when adding/modifying commands
- Include examples in help text

## Questions?

Feel free to open an issue with your question or reach out to the maintainers.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
