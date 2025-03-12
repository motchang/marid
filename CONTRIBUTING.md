# Contributing to Marid

Thank you for your interest in contributing to Marid! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please follow our [Code of Conduct](CODE_OF_CONDUCT.md) in all your interactions with the project.

## How Can I Contribute?

### Reporting Bugs

- Check if the bug has already been reported by searching on GitHub under [Issues](https://github.com/motchang/marid/issues).
- If the bug hasn't been reported, open a new issue. Be sure to include a clear title, description, steps to reproduce, expected behavior, and actual behavior.

### Suggesting Enhancements

- Open a new issue with a clear title and detailed description.
- Include examples of how the enhancement would work.
- Explain why this enhancement would be useful to most Marid users.

### Pull Requests

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development Setup

1. Install Go (version 1.24 or higher)
2. Clone the repository
3. Run `go mod download` to install dependencies
4. Build with `go build -o marid ./cmd/marid`

## Style Guidelines

- Follow standard Go formatting (`go fmt`)
- Use Go idioms and best practices
- Write meaningful commit messages
- Add comments to explain complex logic

## Testing

- Add tests for new features or bug fixes
- Run existing tests before submitting a PR: `go test ./...`
- Make sure your code passes all CI checks

## Documentation

- Update the README.md if necessary
- Document new features or changes in behavior
- Add comments to exported functions

## Additional Notes

- If you're unsure about anything, don't hesitate to ask in an issue
- For large changes, please open an issue first to discuss

Thank you for contributing to Marid!
