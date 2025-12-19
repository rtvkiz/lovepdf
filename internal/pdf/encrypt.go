package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// AddPDFPassword adds password protection to a PDF file.
// The output PDF will require a password to open.
func AddPDFPassword(inputPath, outputPath, password string) error {
	// Create configuration with the new password
	conf := model.NewDefaultConfiguration()
	conf.UserPW = password
	conf.OwnerPW = password // Set both user and owner passwords

	// Encrypt the PDF with the password
	if err := api.EncryptFile(inputPath, outputPath, conf); err != nil {
		return fmt.Errorf("failed to add password: %w", err)
	}

	return nil
}
