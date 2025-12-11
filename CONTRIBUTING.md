# Contributing to LovePDF

Thank you for considering contributing to LovePDF. This document outlines the process for contributing to this project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/lovepdf.git`
3. Create a feature branch: `git checkout -b feature-name`
4. Make your changes
5. Test your changes locally
6. Commit your changes: `git commit -m "Description of changes"`
7. Push to your fork: `git push origin feature-name`
8. Open a pull request

## Development Setup

```bash
# Install Go 1.24 or higher
# Clone the repository
git clone https://github.com/yourusername/lovepdf.git
cd lovepdf

# Install dependencies
go mod download

# Run in development mode
go run cmd/server/main.go

# Build
go build -o lovepdf cmd/server/main.go
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions focused and single-purpose
- Add comments for complex logic
- Use meaningful variable names

## Project Structure

```
lovepdf/
├── cmd/server/          # Application entry point
├── internal/            # Internal packages
│   ├── handlers/        # HTTP handlers
│   ├── pdf/            # PDF processing
│   ├── image/          # Image processing
│   ├── middleware/     # HTTP middleware
│   └── server/         # Server setup
└── web/                # Frontend assets
    ├── templates/      # HTML templates
    └── static/         # CSS and JavaScript
```

## Adding New Features

### Backend Changes

1. Add processing logic to appropriate package (`internal/pdf/` or `internal/image/`)
2. Add HTTP handler to `internal/handlers/`
3. Register route in `internal/server/server.go`
4. Test the endpoint

### Frontend Changes

1. Add HTML template to `web/templates/`
2. Add styles to `web/static/css/style.css`
3. Add JavaScript to `web/static/js/app.js`
4. Test in browser

## Testing

Before submitting a pull request:

1. Test all features manually
2. Test with different file sizes
3. Test with various file types
4. Ensure no errors in browser console
5. Check for memory leaks with large files

## Commit Messages

Write clear, descriptive commit messages:

```
Add feature: Brief description

- Detailed point 1
- Detailed point 2
```

Good examples:
- `Add WebP image support for compression`
- `Fix memory leak in PDF splitting`
- `Improve error handling for large files`

Bad examples:
- `Fix bug`
- `Update code`
- `Changes`

## Pull Request Process

1. Update README.md if adding new features
2. Ensure all tests pass
3. Update documentation as needed
4. Request review from maintainers
5. Address any feedback

## Code of Conduct

- Be respectful and professional
- Focus on constructive feedback
- Help others learn and improve
- Welcome newcomers

## Questions?

Open an issue for:
- Bug reports
- Feature requests
- Questions about the code
- Documentation improvements

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
