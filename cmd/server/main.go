package main

import (
	"flag"
	"log"

	"lovepdf/internal/server"
)

const (
	defaultPort      = ":8080"
	defaultTmpDir    = "./tmp"
	defaultMaxMemory = 100 * 1024 * 1024 // 100MB
)

func main() {
	port := flag.String("port", defaultPort, "Server port")
	tmpDir := flag.String("tmp", defaultTmpDir, "Temporary files directory")
	maxMemory := flag.Int64("max-memory", defaultMaxMemory, "Maximum memory for file uploads in bytes")
	flag.Parse()

	srv := server.New(*port, *tmpDir, *maxMemory)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
