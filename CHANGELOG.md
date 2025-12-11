# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-11

### Added
- PDF splitting with page range selection
- PDF merging with drag-and-drop file reordering
- PDF compression with configurable quality levels
- Image compression for JPEG, PNG, and WebP formats
- Image quality control and format conversion
- Optional image resizing with aspect ratio preservation
- Basic authentication for shared access
- Rate limiting (30 requests per minute per IP)
- Automatic cleanup of temporary files
- Drag-and-drop file upload support
- Real-time file size reduction statistics
- Responsive web interface

### Security
- File type validation
- Path traversal protection
- Configurable file size limits
- No persistent storage of user files
- Optional basic authentication
- Rate limiting to prevent abuse
