---
title: "Contributing"
weight: 4
---

We welcome contributions to Cheetah! Here's how you can help.

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
├── documents/      # Document parsers
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

- Go: Follow standard Go formatting (`gofmt`)
- JavaScript: Use Prettier for formatting
- All files must have SPDX license headers

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request with a clear description

## License

Cheetah is licensed under the MIT License. All contributions must be compatible with this license.

---

Made with ❤️ by [Kartoza](https://kartoza.com) | [Donate!](https://github.com/sponsors/timlinux) | [GitHub](https://github.com/timlinux/cheetah)
