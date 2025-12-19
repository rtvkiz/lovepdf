package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ConvertImagesToPDF converts one or more image files to a single PDF.
// Supported formats: JPEG, PNG, TIFF, WebP
func ConvertImagesToPDF(imagePaths []string, outputPath string) error {
	// Create default configuration
	conf := model.NewDefaultConfiguration()

	// Import images into a PDF
	// This will create a PDF with one page per image
	if err := api.ImportImagesFile(imagePaths, outputPath, nil, conf); err != nil {
		return fmt.Errorf("failed to convert images to PDF: %w", err)
	}

	return nil
}
