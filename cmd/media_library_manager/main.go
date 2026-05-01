// Package main is entry point for media library manager application.
// Orchestrates processing pipeline: [parser], [classifier], [enricher], [resolver], and [transfer].
package main

import (
	"fmt"
	//"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	//"strings"

	//"github.com/ENIACore/media_library_manager/internal/classifier"
	//"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
	//"github.com/ENIACore/media_library_manager/internal/enricher"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/parser"
	//"github.com/ENIACore/media_library_manager/internal/resolver"
	//"github.com/ENIACore/media_library_manager/internal/transfer"
)

func main() {

	if _, err := exec.LookPath("ffprobe"); err != nil {
    	panic("ffprobe is not installed or not in PATH")
	}

	cfg := config.Load()
	logger := logger.NewLogger(cfg)
	//tempDir := createTempDir()

	entries, err := os.ReadDir(cfg.TorrentPath)
	if err != nil {
		panic("unable to read from torrent path")
	}


	for _, entry := range entries {

		entryPath := filepath.Join(cfg.TorrentPath, entry.Name())
		if entry.IsDir() && entry.Name() == filepath.Base(cfg.IncompletePath) {
			logger.Debug("Skipping temp directory", "name", entry.Name())
		}

		root, err := parser.Parse(entryPath, logger)
		if err != nil {
			logger.Error("Parse returned error", "error", err)
		}
		fmt.Println("hi %v", root)
	}
}
