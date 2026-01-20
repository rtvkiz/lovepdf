package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lovepdf/internal/image"
	"lovepdf/internal/pdf"
)

// Response represents the JSON response structure
type Response struct {
	Success        bool   `json:"success,omitempty"`
	Message        string `json:"message,omitempty"`
	Error          string `json:"error,omitempty"`
	DownloadURL    string `json:"downloadUrl,omitempty"`
	OriginalSize   int64  `json:"originalSize,omitempty"`
	CompressedSize int64  `json:"compressedSize,omitempty"`
}

// Home renders the home page
func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "index.html")
}

// SplitPage renders the split page
func SplitPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "split.html")
}

// MergePage renders the merge page
func MergePage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "merge.html")
}

// CompressPage renders the compress page
func CompressPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "compress.html")
}

// CompressImagePage renders the image compression page
func CompressImagePage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "compress-image.html")
}

// CompressGIFPage renders the GIF compression page
func CompressGIFPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "compress-gif.html")
}

// RemovePasswordPage renders the remove password page
func RemovePasswordPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "remove-password.html")
}

// AddPasswordPage renders the add password page
func AddPasswordPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "add-password.html")
}

// RemovePagePage renders the remove page page
func RemovePagePage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "remove-page.html")
}

// ImageToPDFPage renders the image to PDF page
func ImageToPDFPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "image-to-pdf.html")
}

// HandleSplit handles PDF splitting requests
func HandleSplit(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate PDF
		if filepath.Ext(header.Filename) != ".pdf" {
			writeJSONError(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		// Save uploaded file
		inputPath := filepath.Join(tmpDir, generateID()+"_input.pdf")
		if err := saveUploadedFile(file, inputPath); err != nil {
			log.Printf("Error saving file: %v", err)
			writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(inputPath)

		// Get split options
		splitMode := r.FormValue("splitMode")
		pageRange := r.FormValue("pageRange")

		// Split PDF
		var outputPath string
		if splitMode == "range" {
			outputPath, err = pdf.SplitPDF(inputPath, tmpDir, pdf.SplitRange, pageRange)
		} else {
			outputPath, err = pdf.SplitPDF(inputPath, tmpDir, pdf.SplitAll, "")
		}

		if err != nil {
			log.Printf("Error splitting PDF: %v", err)
			writeJSONError(w, fmt.Sprintf("Failed to split PDF: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "PDF split successfully", downloadURL, 0, 0)
	}
}

// HandleMerge handles PDF merging requests
func HandleMerge(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "Files too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded files
		files := r.MultipartForm.File["files"]
		if len(files) < 2 {
			writeJSONError(w, "At least 2 PDF files are required", http.StatusBadRequest)
			return
		}

		if len(files) > 20 {
			writeJSONError(w, "Maximum 20 files allowed", http.StatusBadRequest)
			return
		}

		// Save all uploaded files
		var inputPaths []string
		for i, fileHeader := range files {
			if filepath.Ext(fileHeader.Filename) != ".pdf" {
				writeJSONError(w, "Only PDF files are allowed", http.StatusBadRequest)
				// Clean up previously saved files
				for _, path := range inputPaths {
					os.Remove(path)
				}
				return
			}

			file, err := fileHeader.Open()
			if err != nil {
				writeJSONError(w, "Failed to read uploaded file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			inputPath := filepath.Join(tmpDir, fmt.Sprintf("%s_input_%d.pdf", generateID(), i))
			if err := saveUploadedFile(file, inputPath); err != nil {
				log.Printf("Error saving file: %v", err)
				writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
				return
			}
			inputPaths = append(inputPaths, inputPath)
		}

		// Clean up input files after processing
		defer func() {
			for _, path := range inputPaths {
				os.Remove(path)
			}
		}()

		// Merge PDFs
		outputPath := filepath.Join(tmpDir, generateID()+"_merged.pdf")
		if err := pdf.MergePDFs(inputPaths, outputPath); err != nil {
			log.Printf("Error merging PDFs: %v", err)
			writeJSONError(w, fmt.Sprintf("Failed to merge PDFs: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "PDFs merged successfully", downloadURL, 0, 0)
	}
}

// HandleCompress handles PDF compression requests
func HandleCompress(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate PDF
		if filepath.Ext(header.Filename) != ".pdf" {
			writeJSONError(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		// Save uploaded file
		inputPath := filepath.Join(tmpDir, generateID()+"_input.pdf")
		if err := saveUploadedFile(file, inputPath); err != nil {
			log.Printf("Error saving file: %v", err)
			writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(inputPath)

		// Get original file size
		originalInfo, _ := os.Stat(inputPath)
		originalSize := originalInfo.Size()

		// Get compression level
		compressionLevel := r.FormValue("compression")
		if compressionLevel == "" {
			compressionLevel = "medium"
		}

		// Compress PDF
		outputPath := filepath.Join(tmpDir, generateID()+"_compressed.pdf")
		if err := pdf.CompressPDF(inputPath, outputPath, pdf.CompressionLevel(compressionLevel)); err != nil {
			log.Printf("Error compressing PDF: %v", err)
			writeJSONError(w, fmt.Sprintf("Failed to compress PDF: %v", err), http.StatusInternalServerError)
			return
		}

		// Get compressed file size
		compressedInfo, _ := os.Stat(outputPath)
		compressedSize := compressedInfo.Size()

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "PDF compressed successfully", downloadURL, originalSize, compressedSize)
	}
}

// HandleCompressImage handles image compression requests
func HandleCompressImage(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate image file
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			writeJSONError(w, "Only JPEG, PNG, and WebP images are allowed", http.StatusBadRequest)
			return
		}

		// Save uploaded file
		inputPath := filepath.Join(tmpDir, generateID()+"_input"+ext)
		if err := saveUploadedFile(file, inputPath); err != nil {
			log.Printf("Error saving file: %v", err)
			writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(inputPath)

		// Get original file size
		originalInfo, _ := os.Stat(inputPath)
		originalSize := originalInfo.Size()

		// Parse compression options
		quality := 80
		if q := r.FormValue("quality"); q != "" {
			if parsed, err := strconv.Atoi(q); err == nil && parsed >= 1 && parsed <= 100 {
				quality = parsed
			}
		}

		format := r.FormValue("format")
		if format == "" {
			format = "same"
		}

		resizeMode := r.FormValue("resizeMode")
		if resizeMode == "" {
			resizeMode = "max"
		}

		targetWidth := 0
		if w := r.FormValue("targetWidth"); w != "" {
			if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 {
				targetWidth = parsed
			}
		}

		targetHeight := 0
		if h := r.FormValue("targetHeight"); h != "" {
			if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 {
				targetHeight = parsed
			}
		}

		// Determine output extension
		var outputExt string
		if format == "same" {
			outputExt = ext
		} else {
			outputExt = "." + format
		}

		// Compress image
		outputPath := filepath.Join(tmpDir, generateID()+"_compressed"+outputExt)
		opts := image.CompressionOptions{
			Quality:      quality,
			Format:       format,
			TargetWidth:  targetWidth,
			TargetHeight: targetHeight,
			ResizeMode:   resizeMode,
		}

		if err := image.CompressImage(inputPath, outputPath, opts); err != nil {
			log.Printf("Error compressing image: %v", err)
			writeJSONError(w, fmt.Sprintf("Failed to compress image: %v", err), http.StatusInternalServerError)
			return
		}

		// Get compressed file size
		compressedInfo, _ := os.Stat(outputPath)
		compressedSize := compressedInfo.Size()

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "Image compressed successfully", downloadURL, originalSize, compressedSize)
	}
}

// HandleCompressGIF handles GIF compression requests
func HandleCompressGIF(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate GIF file
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".gif" {
			writeJSONError(w, "Only GIF files are allowed", http.StatusBadRequest)
			return
		}

		// Save uploaded file
		inputPath := filepath.Join(tmpDir, generateID()+"_input.gif")
		if err := saveUploadedFile(file, inputPath); err != nil {
			log.Printf("Error saving file: %v", err)
			writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(inputPath)

		// Get original file size
		originalInfo, _ := os.Stat(inputPath)
		originalSize := originalInfo.Size()

		// Parse compression options
		opts := image.GIFCompressionOptions{
			ColorCount:     parseIntWithDefault(r.FormValue("colorCount"), 256, 2, 256),
			ResizePercent:  parseIntWithDefault(r.FormValue("resizePercent"), 100, 10, 100),
			LossyLevel:     parseIntWithDefault(r.FormValue("lossyLevel"), 0, 0, 100),
			OptimizeFrames: r.FormValue("optimizeFrames") == "true",
			FrameSkip:      parseIntWithDefault(r.FormValue("frameSkip"), 0, 0, 10),
		}

		// Handle presets
		if preset := r.FormValue("preset"); preset != "" {
			opts = applyGIFPreset(preset, opts)
		}

		// Compress GIF
		outputPath := filepath.Join(tmpDir, generateID()+"_compressed.gif")
		if err := image.CompressGIF(inputPath, outputPath, opts); err != nil {
			log.Printf("Error compressing GIF: %v", err)
			writeJSONError(w, fmt.Sprintf("Failed to compress GIF: %v", err), http.StatusInternalServerError)
			return
		}

		// Get compressed file size
		compressedInfo, _ := os.Stat(outputPath)
		compressedSize := compressedInfo.Size()

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "GIF compressed successfully", downloadURL, originalSize, compressedSize)
	}
}

// parseIntWithDefault parses an integer from string with bounds checking
func parseIntWithDefault(value string, defaultVal, min, max int) int {
	if value == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min || parsed > max {
		return defaultVal
	}
	return parsed
}

// applyGIFPreset applies compression presets
func applyGIFPreset(preset string, opts image.GIFCompressionOptions) image.GIFCompressionOptions {
	switch preset {
	case "light":
		opts.ColorCount = 128
		opts.ResizePercent = 90
		opts.OptimizeFrames = true
		opts.FrameSkip = 0
		opts.LossyLevel = 0
	case "medium":
		opts.ColorCount = 64
		opts.ResizePercent = 75
		opts.OptimizeFrames = true
		opts.FrameSkip = 0
		opts.LossyLevel = 0
	case "high":
		opts.ColorCount = 32
		opts.ResizePercent = 50
		opts.OptimizeFrames = true
		opts.FrameSkip = 1
		opts.LossyLevel = 0
	case "maximum":
		opts.ColorCount = 16
		opts.ResizePercent = 40
		opts.OptimizeFrames = true
		opts.FrameSkip = 2
		opts.LossyLevel = 50
	}
	return opts
}

// HandleRemovePassword handles PDF password removal requests
func HandleRemovePassword(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate PDF
		if filepath.Ext(header.Filename) != ".pdf" {
			writeJSONError(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		// Get password
		password := r.FormValue("password")
		if password == "" {
			writeJSONError(w, "Password is required", http.StatusBadRequest)
			return
		}

		// Save uploaded file
		inputPath := filepath.Join(tmpDir, generateID()+"_input.pdf")
		if err := saveUploadedFile(file, inputPath); err != nil {
			log.Printf("Error saving file: %v", err)
			writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(inputPath)

		// Remove password from PDF
		outputPath := filepath.Join(tmpDir, generateID()+"_unlocked.pdf")
		if err := pdf.RemovePDFPassword(inputPath, outputPath, password); err != nil {
			log.Printf("Error removing password: %v", err)
			writeJSONError(w, "Failed to remove password. Please check if the password is correct.", http.StatusBadRequest)
			return
		}

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "Password removed successfully. PDF is now unlocked and can be shared freely.", downloadURL, 0, 0)
	}
}

// HandleAddPassword handles PDF password addition requests
func HandleAddPassword(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate PDF
		if filepath.Ext(header.Filename) != ".pdf" {
			writeJSONError(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		// Get password
		password := r.FormValue("password")
		if password == "" {
			writeJSONError(w, "Password is required", http.StatusBadRequest)
			return
		}

		// Save uploaded file
		inputPath := filepath.Join(tmpDir, generateID()+"_input.pdf")
		if err := saveUploadedFile(file, inputPath); err != nil {
			log.Printf("Error saving file: %v", err)
			writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(inputPath)

		// Add password to PDF
		outputPath := filepath.Join(tmpDir, generateID()+"_protected.pdf")
		if err := pdf.AddPDFPassword(inputPath, outputPath, password); err != nil {
			log.Printf("Error adding password: %v", err)
			writeJSONError(w, "Failed to add password to PDF.", http.StatusInternalServerError)
			return
		}

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "Password added successfully. PDF is now protected.", downloadURL, 0, 0)
	}
}

// HandleRemovePage handles PDF page removal requests
func HandleRemovePage(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSONError(w, "No file uploaded", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate PDF
		if filepath.Ext(header.Filename) != ".pdf" {
			writeJSONError(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		// Get page range
		pageRange := r.FormValue("pageRange")
		if pageRange == "" {
			writeJSONError(w, "Page range is required", http.StatusBadRequest)
			return
		}

		// Save uploaded file
		inputPath := filepath.Join(tmpDir, generateID()+"_input.pdf")
		if err := saveUploadedFile(file, inputPath); err != nil {
			log.Printf("Error saving file: %v", err)
			writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(inputPath)

		// Remove pages from PDF
		outputPath := filepath.Join(tmpDir, generateID()+"_pages_removed.pdf")
		if err := pdf.RemovePDFPages(inputPath, outputPath, pageRange); err != nil {
			log.Printf("Error removing pages: %v", err)
			writeJSONError(w, fmt.Sprintf("Failed to remove pages: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "Pages removed successfully from PDF.", downloadURL, 0, 0)
	}
}

// HandleImageToPDF handles image to PDF conversion requests
func HandleImageToPDF(tmpDir string, maxMemory int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			writeJSONError(w, "Files too large or invalid form data", http.StatusBadRequest)
			return
		}

		// Get uploaded files
		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			writeJSONError(w, "At least 1 image file is required", http.StatusBadRequest)
			return
		}

		if len(files) > 50 {
			writeJSONError(w, "Maximum 50 files allowed", http.StatusBadRequest)
			return
		}

		// Save all uploaded files
		var inputPaths []string
		for i, fileHeader := range files {
			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".tiff" && ext != ".tif" && ext != ".webp" {
				writeJSONError(w, "Only JPEG, PNG, TIFF, and WebP images are allowed", http.StatusBadRequest)
				// Clean up previously saved files
				for _, path := range inputPaths {
					os.Remove(path)
				}
				return
			}

			file, err := fileHeader.Open()
			if err != nil {
				writeJSONError(w, "Failed to read uploaded file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			inputPath := filepath.Join(tmpDir, fmt.Sprintf("%s_input_%d%s", generateID(), i, ext))
			if err := saveUploadedFile(file, inputPath); err != nil {
				log.Printf("Error saving file: %v", err)
				writeJSONError(w, "Failed to save uploaded file", http.StatusInternalServerError)
				return
			}
			inputPaths = append(inputPaths, inputPath)
		}

		// Clean up input files after processing
		defer func() {
			for _, path := range inputPaths {
				os.Remove(path)
			}
		}()

		// Convert images to PDF
		outputPath := filepath.Join(tmpDir, generateID()+"_images_to_pdf.pdf")
		if err := pdf.ConvertImagesToPDF(inputPaths, outputPath); err != nil {
			log.Printf("Error converting images to PDF: %v", err)
			writeJSONError(w, fmt.Sprintf("Failed to convert images to PDF: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate download URL
		downloadURL := fmt.Sprintf("/download/%s", filepath.Base(outputPath))

		writeJSONSuccess(w, "Images converted to PDF successfully.", downloadURL, 0, 0)
	}
}

// HandleDownload handles file download requests
func HandleDownload(tmpDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := filepath.Base(r.URL.Path)
		filePath := filepath.Join(tmpDir, filename)

		// Security check: ensure file is in tmp directory
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}

		absTmpDir, err := filepath.Abs(tmpDir)
		if err != nil {
			http.Error(w, "Invalid tmp directory", http.StatusInternalServerError)
			return
		}

		if !filepath.HasPrefix(absPath, absTmpDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Serve file
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		http.ServeFile(w, r, filePath)
	}
}

// Helper functions

func renderTemplate(w http.ResponseWriter, tmpl string) {
	t, err := template.ParseFiles(filepath.Join("web", "templates", tmpl))
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func writeJSONSuccess(w http.ResponseWriter, message, downloadURL string, originalSize, compressedSize int64) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Success:        true,
		Message:        message,
		DownloadURL:    downloadURL,
		OriginalSize:   originalSize,
		CompressedSize: compressedSize,
	})
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}

func saveUploadedFile(src io.Reader, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
