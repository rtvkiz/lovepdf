package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"lovepdf/internal/handlers"
	"lovepdf/internal/middleware"
)

type Server struct {
	addr      string
	tmpDir    string
	maxMemory int64
}

func New(addr, tmpDir string, maxMemory int64) *Server {
	return &Server{
		addr:      addr,
		tmpDir:    tmpDir,
		maxMemory: maxMemory,
	}
}

func (s *Server) Start() error {
	// Ensure tmp directory exists
	if err := os.MkdirAll(s.tmpDir, 0755); err != nil {
		return err
	}

	// Start cleanup goroutine
	go s.cleanupOldFiles()

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Page routes
	mux.HandleFunc("/", handlers.Home)
	mux.HandleFunc("/split", handlers.SplitPage)
	mux.HandleFunc("/merge", handlers.MergePage)
	mux.HandleFunc("/compress", handlers.CompressPage)
	mux.HandleFunc("/compress-image", handlers.CompressImagePage)
	mux.HandleFunc("/compress-gif", handlers.CompressGIFPage)
	mux.HandleFunc("/remove-password", handlers.RemovePasswordPage)
	mux.HandleFunc("/add-password", handlers.AddPasswordPage)
	mux.HandleFunc("/remove-page", handlers.RemovePagePage)
	mux.HandleFunc("/image-to-pdf", handlers.ImageToPDFPage)

	// API routes
	mux.HandleFunc("/api/split", handlers.HandleSplit(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/merge", handlers.HandleMerge(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/compress", handlers.HandleCompress(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/compress-image", handlers.HandleCompressImage(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/compress-gif", handlers.HandleCompressGIF(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/remove-password", handlers.HandleRemovePassword(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/add-password", handlers.HandleAddPassword(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/remove-page", handlers.HandleRemovePage(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/api/image-to-pdf", handlers.HandleImageToPDF(s.tmpDir, s.maxMemory))
	mux.HandleFunc("/download/", handlers.HandleDownload(s.tmpDir))

	// Wrap with middleware
	var handler http.Handler = mux

	// Add rate limiting (30 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(30)
	handler = rateLimiter.Limit(handler)

	// Add basic auth if configured via environment variables
	handler = middleware.BasicAuth(handler)

	log.Printf("Server starting on %s", s.addr)
	log.Printf("Open http://localhost%s in your browser", s.addr)

	// Check if auth is enabled
	if os.Getenv("AUTH_USERNAME") != "" {
		log.Printf("Basic authentication enabled")
	}

	return http.ListenAndServe(s.addr, handler)
}

// cleanupOldFiles removes temporary files older than 1 hour
func (s *Server) cleanupOldFiles() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		files, err := os.ReadDir(s.tmpDir)
		if err != nil {
			log.Printf("Error reading tmp directory: %v", err)
			continue
		}

		now := time.Now()
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			if now.Sub(info.ModTime()) > time.Hour {
				path := filepath.Join(s.tmpDir, file.Name())
				if err := os.Remove(path); err != nil {
					log.Printf("Error removing old file %s: %v", path, err)
				} else {
					log.Printf("Cleaned up old file: %s", file.Name())
				}
			}
		}
	}
}
