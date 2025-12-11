package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// RemovePDFPassword removes password protection from a PDF file.
// The output PDF will have no password and can be shared freely.
func RemovePDFPassword(inputPath, outputPath, password string) error {
	// Create configuration with the current password
	conf := model.NewDefaultConfiguration()
	conf.UserPW = password

	// Decrypt the PDF - this creates a new PDF without any password protection
	if err := api.DecryptFile(inputPath, outputPath, conf); err != nil {
		return fmt.Errorf("failed to remove password: %w", err)
	}

	return nil
}
