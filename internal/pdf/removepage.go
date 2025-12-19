package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// RemovePDFPages removes specific pages from a PDF file.
// pageRange should be in format like "1,3,5" or "1-3,5,7-9"
func RemovePDFPages(inputPath, outputPath, pageRange string) error {
	// Create default configuration
	conf := model.NewDefaultConfiguration()

	// Remove pages from PDF
	if err := api.RemovePagesFile(inputPath, outputPath, []string{pageRange}, conf); err != nil {
		return fmt.Errorf("failed to remove pages: %w", err)
	}

	return nil
}
