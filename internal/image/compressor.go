package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

type CompressionOptions struct {
	Quality   int    // 1-100
	Format    string // "same", "jpeg", "png", "webp"
	MaxWidth  int    // 0 means no resize
	MaxHeight int    // 0 means no resize
}

// CompressImage compresses an image file with the given options
func CompressImage(inputPath, outputPath string, opts CompressionOptions) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Detect input format
	inputFormat, err := detectImageFormat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to detect image format: %w", err)
	}

	// Decode image
	img, err := decodeImage(inputFile, inputFormat)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if needed
	if opts.MaxWidth > 0 || opts.MaxHeight > 0 {
		img = resizeImage(img, opts.MaxWidth, opts.MaxHeight)
	}

	// Determine output format
	outputFormat := opts.Format
	if outputFormat == "same" || outputFormat == "" {
		outputFormat = inputFormat
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Encode with compression
	if err := encodeImage(outputFile, img, outputFormat, opts.Quality); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

// detectImageFormat detects the image format from file extension
func detectImageFormat(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg", nil
	case ".png":
		return "png", nil
	case ".webp":
		return "webp", nil
	default:
		return "", fmt.Errorf("unsupported image format: %s", ext)
	}
}

// decodeImage decodes an image based on its format
func decodeImage(r io.Reader, format string) (image.Image, error) {
	switch format {
	case "jpeg":
		return jpeg.Decode(r)
	case "png":
		return png.Decode(r)
	case "webp":
		return webp.Decode(r)
	default:
		return nil, fmt.Errorf("unsupported decode format: %s", format)
	}
}

// encodeImage encodes an image in the specified format with quality
func encodeImage(w io.Writer, img image.Image, format string, quality int) error {
	switch format {
	case "jpeg":
		opts := &jpeg.Options{Quality: quality}
		return jpeg.Encode(w, img, opts)
	case "png":
		// PNG uses compression level, not quality
		// We'll use default encoder which provides good compression
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		return encoder.Encode(w, img)
	case "webp":
		// For WebP, we'll convert to JPEG as a fallback
		// since encoding WebP requires external library
		// For now, convert to JPEG with quality
		opts := &jpeg.Options{Quality: quality}
		return jpeg.Encode(w, img, opts)
	default:
		return fmt.Errorf("unsupported encode format: %s", format)
	}
}

// resizeImage resizes an image to fit within maxWidth and maxHeight while preserving aspect ratio
func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If no resize needed, return original
	if (maxWidth == 0 || width <= maxWidth) && (maxHeight == 0 || height <= maxHeight) {
		return img
	}

	// Calculate new dimensions while preserving aspect ratio
	newWidth := width
	newHeight := height

	if maxWidth > 0 && width > maxWidth {
		newWidth = maxWidth
		newHeight = int(float64(height) * float64(maxWidth) / float64(width))
	}

	if maxHeight > 0 && newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = int(float64(width) * float64(maxHeight) / float64(height))
	}

	// Create new image with calculated dimensions
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Use high-quality scaling
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}
