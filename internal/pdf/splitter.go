package pdf

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// generateID creates a unique identifier for file naming
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

type SplitMode string

const (
	SplitAll   SplitMode = "all"
	SplitRange SplitMode = "range"
)

// SplitPDF splits a PDF file based on the specified mode
func SplitPDF(inputPath, outputDir string, mode SplitMode, pageRange string) (string, error) {
	if mode == SplitAll {
		return splitAllPages(inputPath, outputDir)
	}
	return splitByRange(inputPath, outputDir, pageRange)
}

// splitAllPages splits PDF into individual pages
func splitAllPages(inputPath, outputDir string) (string, error) {
	// Create unique output directory for this split operation using MkdirTemp
	// This guarantees a unique directory name and avoids collisions
	splitDir, err := os.MkdirTemp(outputDir, "split_")
	if err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Split into individual pages
	if err := api.SplitFile(inputPath, splitDir, 1, nil); err != nil {
		os.RemoveAll(splitDir) // Clean up on error
		return "", fmt.Errorf("failed to split PDF: %w", err)
	}

	// Create ZIP file with unique name based on the split directory name
	zipName := filepath.Base(splitDir) + ".zip"
	zipPath := filepath.Join(outputDir, zipName)

	if err := zipDirectory(splitDir, zipPath); err != nil {
		os.RemoveAll(splitDir) // Clean up on error
		return "", fmt.Errorf("failed to create ZIP: %w", err)
	}

	// Clean up the temporary split directory after zipping
	os.RemoveAll(splitDir)

	return zipPath, nil
}

// splitByRange extracts specific pages from PDF
func splitByRange(inputPath, outputDir, pageRange string) (string, error) {
	// Parse page ranges (e.g., "1-3,5,7-9")
	pages, err := parsePageRanges(pageRange)
	if err != nil {
		return "", fmt.Errorf("invalid page range: %w", err)
	}

	if len(pages) == 0 {
		return "", fmt.Errorf("no valid pages specified")
	}

	// Create temp file for output to ensure uniqueness
	tmpFile, err := os.CreateTemp(outputDir, "extracted_*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	outputPath := tmpFile.Name()
	tmpFile.Close()
	// We only needed the name, ExtractPagesFile will overwrite/write to it
	// But ExtractPagesFile expects the file to NOT be open.

	// Extract specified pages
	if err := api.ExtractPagesFile(inputPath, outputPath, pages, nil); err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("failed to extract pages: %w", err)
	}

	return outputPath, nil
}

// parsePageRanges parses page range strings like "1-3,5,7-9"
func parsePageRanges(rangeStr string) ([]string, error) {
	var pages []string
	parts := strings.Split(rangeStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Range like "1-3"
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start page: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end page: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("invalid range: start > end (%d > %d)", start, end)
			}

			for i := start; i <= end; i++ {
				pages = append(pages, strconv.Itoa(i))
			}
		} else {
			// Single page
			pageNum, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid page number: %s", part)
			}
			pages = append(pages, strconv.Itoa(pageNum))
		}
	}

	return pages, nil
}

// zipDirectory creates a ZIP file from a directory
func zipDirectory(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipEntry, file)
		return err
	})
}
