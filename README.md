# LovePDF

A local web application for PDF and image manipulation. All processing happens on your machine - no cloud uploads, completely private.

## Features

- Split PDFs by page ranges or extract individual pages
- Merge multiple PDF files into a single document
- Compress PDFs to reduce file size
- Compress images (JPEG, PNG, WebP) with quality control and dimension presets (passport photos, ID photos, etc.)
- All processing happens locally on your machine
- No internet connection required
- Privacy-focused - your files never leave your computer

## Requirements

- Go 1.24 or higher (for building from source)
- Linux, macOS, or Windows

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/lovepdf.git
cd lovepdf

# Install dependencies
go mod download

# Build the application
go build -o lovepdf cmd/server/main.go

# Run it
./lovepdf
```

The server will start on port 8080. Open your browser to `http://localhost:8080`

### Option 2: Download Pre-built Binary

Download the latest release for your platform from the releases page and run it:

```bash
./lovepdf
```

## Usage

### Starting the Server

Basic usage:
```bash
./lovepdf
```

With custom options:
```bash
./lovepdf -port :8081 -tmp ./temp -max-memory 209715200
```

Available flags:
- `-port` - Server port (default: :8080)
- `-tmp` - Temporary files directory (default: ./tmp)
- `-max-memory` - Maximum file upload size in bytes (default: 104857600 / 100MB)

### Enabling Authentication

For shared access, enable basic authentication using environment variables:

```bash
AUTH_USERNAME=admin AUTH_PASSWORD=secretpass ./lovepdf
```

### Split PDF

1. Navigate to Split PDF from the home page
2. Upload a PDF file
3. Choose splitting mode:
   - Extract all pages as separate files (downloads as ZIP)
   - Extract specific pages using ranges (e.g., "1-3,5,7-9")
4. Process and download

### Merge PDFs

1. Navigate to Merge PDFs from the home page
2. Upload multiple PDF files (minimum 2, maximum 20)
3. Drag and drop to reorder files if needed
4. Process and download the merged PDF

### Compress PDF

1. Navigate to Compress PDF from the home page
2. Upload a PDF file
3. Choose compression level (low/medium/high)
4. Process and download the compressed file

### Compress Image

1. Navigate to Compress Image from the home page
2. Upload an image (JPEG, PNG, or WebP)
3. Adjust settings:
   - Quality: 1-100% (lower = smaller file)
   - Output format: Keep original or convert to JPEG/PNG/WebP
   - Resize options:
     - Fit within maximum dimensions (maintains aspect ratio)
     - Exact dimensions (stretches to fit)
   - Dimension presets:
     - Passport Photo (2x2 inches / 600x600px at 300 DPI)
     - ID Photo (1.5x2 inches / 450x600px at 300 DPI)
     - HD (1920x1080px)
     - Square sizes (1024x1024px or 512x512px)
     - Custom dimensions
4. Preview the image
5. Process and download the compressed file

### Password Remove from PDF

Note: Uses high-quality Catmull-Rom interpolation to maintain image quality during resizing.

## Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o lovepdf-linux cmd/server/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o lovepdf-mac cmd/server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o lovepdf.exe cmd/server/main.go
```

## Project Structure

```
lovepdf/
├── cmd/server/          # Application entry point
├── internal/
│   ├── handlers/        # HTTP request handlers
│   ├── pdf/            # PDF processing logic
│   ├── image/          # Image processing logic
│   ├── middleware/     # Authentication and rate limiting
│   └── server/         # Server configuration
├── web/
│   ├── templates/      # HTML templates
│   └── static/         # CSS and JavaScript
├── tmp/                # Temporary file storage (auto-cleaned)
├── go.mod              # Go module definition
└── README.md           # This file
```

## Technical Details

- **Language**: Go 1.24+
- **PDF Processing**: pdfcpu (pure Go, no external dependencies)
- **Image Processing**: Go standard library + golang.org/x/image for WebP and scaling
- **Web Server**: Go net/http standard library
- **Frontend**: Vanilla JavaScript and CSS

## Security Features

- File type validation
- Configurable file size limits (default 100MB)
- Path traversal protection
- Automatic cleanup of temporary files (after 1 hour)
- Optional basic authentication
- Rate limiting (30 requests/minute per IP)
- No persistent storage of user files

## Configuration

### Environment Variables

- `AUTH_USERNAME` - Username for basic authentication (optional)
- `AUTH_PASSWORD` - Password for basic authentication (optional)

### Command Line Flags

- `-port` - Server listening port
- `-tmp` - Temporary files directory
- `-max-memory` - Maximum memory for file uploads in bytes

## Development

### Running in Development

```bash
go run cmd/server/main.go
```

### Running Tests

```bash
go test ./...
```

### Adding New Features

The codebase follows a simple structure:
- Add processing logic to `internal/pdf/` or `internal/image/`
- Add HTTP handlers to `internal/handlers/`
- Add routes in `internal/server/server.go`
- Add UI templates to `web/templates/`
- Add frontend code to `web/static/`

## Troubleshooting

### Port Already in Use

If port 8080 is in use, specify a different port:
```bash
./lovepdf -port :8081
```

### File Upload Fails

Check that your file size is within limits. Increase if needed:
```bash
./lovepdf -max-memory 209715200  # 200MB
```

### Permission Errors

Ensure the application has write permissions for the tmp directory:
```bash
mkdir -p tmp
chmod 755 tmp
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome. Please open an issue or pull request.

## Acknowledgments

- PDF processing powered by [pdfcpu](https://github.com/pdfcpu/pdfcpu)
- Image processing using Go standard library and golang.org/x/image
