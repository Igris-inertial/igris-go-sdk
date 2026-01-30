# Contributing to Igris SDK

Thank you for your interest in contributing to the Igris SDK project. This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with the following information:

- Clear description of the problem
- Steps to reproduce the issue
- Expected behavior vs actual behavior
- Your environment (OS, language version, SDK version)
- Any relevant code snippets or error messages

### Suggesting Features

Feature requests are welcome. Please create an issue describing:

- The problem you're trying to solve
- Your proposed solution
- Any alternative solutions you've considered
- How this feature would benefit other users

### Submitting Pull Requests

1. Fork the repository
2. Create a new branch for your feature (`git checkout -b feature/amazing-feature`)
3. Make your changes following our coding standards (see below)
4. Write or update tests as needed
5. Ensure all tests pass
6. Commit your changes with clear, descriptive commit messages
7. Push to your fork (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Pull Request Guidelines

- **One feature per PR**: Keep pull requests focused on a single feature or bug fix
- **Write tests**: All new features and bug fixes should include tests
- **Update documentation**: Update README and inline documentation as needed
- **Follow code style**: Adhere to the language-specific coding standards below
- **Clean commit history**: Squash commits if necessary before submitting

## Coding Standards

### Python
- Follow PEP 8 style guide
- Use type hints where appropriate
- Format code with `black` (line length: 100)
- Run `flake8` for linting
- Run `mypy` for type checking

### Rust
- Follow official Rust style guidelines
- Format code with `cargo fmt`
- Run `cargo clippy` for linting
- Write documentation comments for public APIs
- Use meaningful variable and function names

### JavaScript/TypeScript
- Use TypeScript for new code
- Follow Airbnb JavaScript Style Guide
- Format code with `prettier`
- Run `eslint` for linting
- Provide type definitions for all public APIs

### Go
- Follow official Go style guidelines
- Format code with `gofmt`
- Run `go vet` for linting
- Write clear, concise documentation comments
- Use meaningful package and function names

### Java
- Follow Google Java Style Guide
- Use meaningful variable and method names
- Write JavaDoc comments for public APIs
- Keep methods focused and concise

### C#
- Follow Microsoft C# Coding Conventions
- Use PascalCase for public members
- Use camelCase for private members
- Write XML documentation comments

### Ruby
- Follow Ruby Style Guide
- Use meaningful variable and method names
- Write clear documentation
- Keep methods short and focused

## Testing

All contributions must include appropriate tests:

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test component interactions
- **Example tests**: Ensure examples remain functional

Run the test suite before submitting:

```bash
# Python
pytest

# Rust
cargo test

# JavaScript/TypeScript
npm test

# Go
go test ./...

# Java
mvn test

# C#
dotnet test

# Ruby
rspec
```

## Documentation

- Update README.md if you change functionality
- Add inline code comments for complex logic
- Update API documentation for public interfaces
- Include usage examples for new features

## Commit Message Guidelines

Write clear, descriptive commit messages:

```
feat: add support for streaming responses
fix: resolve timeout issue in retry logic
docs: update installation instructions
test: add unit tests for error handling
refactor: simplify authentication flow
```

Prefix types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

## Development Setup

### Prerequisites

Install the required development tools for your language:

- **Python**: Python 3.8+, pip, virtualenv
- **Rust**: Rust 1.70+, cargo
- **JavaScript**: Node.js 14+, npm or yarn
- **Go**: Go 1.21+
- **Java**: JDK 11+, Maven
- **C#**: .NET 6+
- **Ruby**: Ruby 2.7+, bundler

### Local Development

1. Clone the repository
2. Install dependencies
3. Run tests to verify setup
4. Make your changes
5. Run tests again
6. Submit pull request

## Release Process

Releases are managed by maintainers. Version numbers follow Semantic Versioning (SemVer):

- MAJOR version for incompatible API changes
- MINOR version for new functionality (backwards-compatible)
- PATCH version for backwards-compatible bug fixes

## Getting Help

If you need help:

- Check existing issues and discussions
- Read the documentation in the README
- Ask questions in GitHub Discussions
- Contact maintainers via issues

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.

## Attribution

Contributors will be recognized in the project's contributor list. Significant contributions may be highlighted in release notes.

Thank you for contributing to Igris SDK!
