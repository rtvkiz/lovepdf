package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type CompressionLevel string

const (
	CompressionLow    CompressionLevel = "low"
	CompressionMedium CompressionLevel = "medium"
	CompressionHigh   CompressionLevel = "high"
)

// CompressPDF reduces the size of a PDF file
func CompressPDF(inputPath, outputPath string, level CompressionLevel) error {
	// Use default configuration for optimization
	// pdfcpu's OptimizeFile will compress the PDF by removing unnecessary elements
	// and optimizing internal structures
	if err := api.OptimizeFile(inputPath, outputPath, nil); err != nil {
		return fmt.Errorf("failed to compress PDF: %w", err)
	}

	return nil
}
