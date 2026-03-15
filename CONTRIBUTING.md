# Contributing to Cheetah

Contributions are welcome and appreciated! Here's how you can help.

## Development Setup

```bash
# Clone the repository
git clone https://github.com/timlinux/cheetah.git
cd cheetah

# Enter the Nix development environment
nix develop

# Build the project
make build

# Run tests
make test
```

## Project Structure

```
cheetah/
├── backend/        # Go REST API server
├── documents/      # Document parsers (PDF, DOCX, EPUB, ODT, TXT, MD)
├── frontend/       # Bubble Tea TUI
├── (uses: github.com/timlinux/blockfont)  # Block letter rendering
├── sessions/       # Position persistence
├── settings/       # User preferences
├── web/            # React web frontend
│   ├── src/
│   │   ├── components/
│   │   └── parsers/
│   └── tests/      # Playwright E2E tests
└── hugo/           # Documentation site
```

## Running Tests

```bash
# Run Go tests
make test

# Run Go tests with coverage
make test-coverage

# Run Playwright E2E tests
make test-web

# Run all tests
make test-all
```

## Code Style

- **Go**: Follow standard Go formatting (`gofmt`)
- **JavaScript/React**: Use Prettier for formatting
- **All files**: Must have SPDX license headers
- **Commits**: Must conform to [Conventional Commits](https://www.conventionalcommits.org/)

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for new functionality
4. Ensure all tests pass (`make test-all`)
5. Commit your changes with a clear, conventional commit message
6. Push to your fork and submit a pull request

## Reporting Bugs

- Use the [bug report template](../../issues/new?template=bug_report.yml)
- Include steps to reproduce
- Include expected vs actual behavior
- Include screenshots if applicable

## Requesting Features

- Use the [feature request template](../../issues/new?template=feature_request.yml)
- Describe the use case
- Explain the expected behavior

## License

Cheetah is licensed under the MIT License. All contributions must be compatible with this license.

---

Made with ❤️ by [Kartoza](https://kartoza.com) | [Donate!](https://github.com/sponsors/timlinux) | [GitHub](https://github.com/timlinux/cheetah)
