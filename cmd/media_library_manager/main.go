// Package main is entry point for media library manager application.
// Orchestrates processing pipeline: [parser], [classifier], [enricher], [resolver], and [transfer].
package main

import (
	"os/exec"
	"os"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/logger"
)

func main() {

	if _, err := exec.LookPath("ffprobe"); err != nil {
		panic("ffprobe is not installed or not in PATH")
	}
	if _, err := exec.LookPath("mkvmerge"); err != nil {
		panic("mkvmerge is not installed or not in PATH")
	}

	cfg := config.Load()
	logger := logger.NewLogger(cfg)
	tempDir := createTempDir()

	switch cfg.Mode {
	case "ingest":
		ingest(tempDir, cfg, logger)
	case "subtitle":
		subtitle(cfg, logger)
	default:
		panic("Invalid mode detected, only ingest or subtitle authorized")
	}
}

func createTempDir() string {
	tempDir, err := os.MkdirTemp("", "media-library-manager-*")
	if err != nil {
		panic("Error creating temporary directory for dummy output")
	}
	defer os.RemoveAll(tempDir)
	return tempDir
}

func overLimit(count int, cfg *config.Config) bool {
	if cfg.Limit != 0 && count > cfg.Limit {
		return true
	}
	return false
}
