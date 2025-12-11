package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// MergePDFs combines multiple PDF files into one
func MergePDFs(inputPaths []string, outputPath string) error {
	if len(inputPaths) < 2 {
		return fmt.Errorf("at least 2 PDF files are required for merging")
	}

	// Validate all input files exist
	for _, path := range inputPaths {
		if path == "" {
			return fmt.Errorf("invalid file path")
		}
	}

	// Merge PDFs using pdfcpu
	if err := api.MergeCreateFile(inputPaths, outputPath, false, nil); err != nil {
		return fmt.Errorf("failed to merge PDFs: %w", err)
	}

	return nil
}
