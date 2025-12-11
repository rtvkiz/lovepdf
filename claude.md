# LovePDF - Local PDF & Image Tool

## Overview
A local web application for PDF and image manipulation: splitting, merging, and compressing PDFs, plus image compression. Runs entirely on the user's machine for privacy and offline use.

## Purpose
- **Split PDFs**: Extract pages, split by page ranges, or split into individual pages
- **Merge PDFs**: Combine multiple PDF files into a single document
- **Compress PDFs**: Reduce PDF file size while maintaining quality
- **Compress Images**: Reduce image file size (JPEG, PNG, WebP) with quality control and optional resizing
- **Local-first**: All processing happens locally, no cloud uploads

## Technology Stack
- **Language**: Go
- **Type**: Local web server with browser-based UI
- **Processing**: Local PDF and image manipulation (no external API calls)

## Project Structure
```
lovepdf/
├── cmd/
│   └── server/
│       └── main.go           # Main server entry point
├── internal/
│   ├── handlers/
│   │   └── handlers.go       # All HTTP handlers (PDF & Image)
│   ├── pdf/
│   │   ├── splitter.go       # PDF splitting logic
│   │   ├── merger.go         # PDF merging logic
│   │   └── compressor.go     # PDF compression logic
│   ├── image/
│   │   └── compressor.go     # Image compression logic
│   └── server/
│       └── server.go         # HTTP server setup
├── web/
│   ├── templates/
│   │   ├── index.html        # Main page with feature selection
│   │   ├── split.html        # Split interface
│   │   ├── merge.html        # Merge interface
│   │   └── compress.html     # Compress interface
│   └── static/
│       ├── css/
│       │   └── style.css
│       ├── js/
│       │   └── app.js
│       └── favicon.ico
├── tmp/                      # Temporary file storage (gitignored)
├── go.mod
├── go.sum
└── README.md
```

## Features

### PDF Splitting
- Split by page ranges (e.g., "1-5, 10-15")
- Extract specific pages
- Split into individual pages
- Split into chunks of N pages
- Download as ZIP when splitting into multiple files

### PDF Merging
- Combine multiple PDF files into one
- Reorder files before merging (drag-and-drop)
- Preview file order
- Merge 2+ PDFs with preserved formatting

### PDF Compression
- Reduce file size by optimizing images
- Remove unnecessary metadata
- Downsample images while preserving readability
- Adjustable compression levels (low, medium, high)
- Show original vs compressed file size

### Image Compression
- Support for JPEG, PNG, and WebP formats
- Adjustable quality slider (1-100%)
- Format conversion (convert between JPEG/PNG/WebP)
- Optional image resizing (set max width/height)
- Image preview before compression
- Shows size reduction percentage
- Maintains aspect ratio when resizing
- High-quality scaling algorithm (Catmull-Rom)

## Development Guidelines

### Go Conventions
- Use standard library where possible
- Follow idiomatic Go patterns
- Keep handlers thin, logic in services
- Proper error handling and logging

### PDF Library
Recommended: **pdfcpu** (pure Go, no external dependencies)
- `github.com/pdfcpu/pdfcpu` - full-featured, pure Go PDF processor
- No CGO required, easy cross-platform builds
- Supports splitting, merging, and optimization

### Local-First Design
- No external API calls or internet dependency
- Process files in memory when possible
- Clean up temporary files immediately after processing
- Fast processing for responsive UI
- All operations happen client-side (in the Go server)

### Security
- Validate PDF files before processing
- Set reasonable file size limits (e.g., 100MB per file)
- Use unique temporary filenames to prevent conflicts
- Clean up temp files on server shutdown
- No persistent storage of user files
- Prevent path traversal attacks

### Web Interface
- Simple, intuitive drag-and-drop UI
- Real-time progress feedback for large files
- Download processed files directly
- Responsive design for different screen sizes
- Clear visual feedback for multi-file operations (merge)

## Configuration
Default settings:
- Port: `8080`
- Max file size: `100MB` per file
- Max files for merge: `20` files
- Temp directory: `./tmp`
- Auto-cleanup: Delete files after download or 1 hour

## Dependencies
```go
require (
    github.com/pdfcpu/pdfcpu v0.x.x  // PDF processing
    golang.org/x/image v0.x.x        // Image processing (WebP, scaling)
    // Web server uses Go stdlib (no framework needed)
)
```

## Usage
```bash
# Run the server
go run cmd/server/main.go

# Open browser
http://localhost:8080
```

## Build
```bash
# Build for current platform
go build -o lovepdf cmd/server/main.go

# Cross-compile for other platforms
GOOS=windows GOARCH=amd64 go build -o lovepdf.exe cmd/server/main.go
GOOS=darwin GOARCH=amd64 go build -o lovepdf-mac cmd/server/main.go
GOOS=linux GOARCH=amd64 go build -o lovepdf-linux cmd/server/main.go
```

## API Endpoints (Internal)
```
GET  /                      # Main page
GET  /split                 # Split page
GET  /merge                 # Merge page
GET  /compress              # Compress page

POST /api/split             # Handle split operation
POST /api/merge             # Handle merge operation
POST /api/compress          # Handle compress operation

GET  /download/:id          # Download processed file
```

## Notes for Claude
- Keep the UI simple and focused on the three main features
- Prioritize fast local processing over advanced features
- Use pdfcpu for pure Go implementation (no external dependencies)
- Ensure proper cleanup of temporary files
- Make it easy to distribute as a single binary
- No database needed - stateless design
- Handle multiple file uploads gracefully for merge feature
- Provide clear error messages for invalid PDFs
