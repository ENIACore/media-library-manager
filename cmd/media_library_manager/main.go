// Package main is entry point for media library manager application.
// Orchestrates processing pipeline: [parser], [classifier], [enricher], [resolver], and [transfer].
package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/classifier"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/enricher"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/parser"
	"github.com/ENIACore/media_library_manager/internal/remuxer"
	"github.com/ENIACore/media_library_manager/internal/verifier"
	"github.com/ENIACore/media_library_manager/internal/resolver"
	"github.com/ENIACore/media_library_manager/internal/transfer"
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

	entries, err := os.ReadDir(cfg.TorrentPath)
	if err != nil {
		panic("unable to read from torrent path")
	}

	numFailure := 0
	numSuccess := 0

	for _, entry := range entries {

		root, err := process(entry, cfg, logger)

		if err != nil {
			numFailure += 1
			transfer.Error(root, cfg, logger)
			continue
		} 
		if root == nil {
			continue
		}

		output := tree(root.FileInfo.SourcePath)
		fmt.Println("")
		fmt.Println("------------- Old Structure") 
		fmt.Println(output)

		if cfg.DryRun {
			transfer.TestTransfer(root, tempDir, logger)
		} else {
			transfer.Transfer(root, cfg, logger)
		}

		if !root.FileInfo.IsDir {
      		root.FileInfo.DestPath = filepath.Dir(root.FileInfo.DestPath)
  		}
		output = tree(root.FileInfo.DestPath)
		fmt.Println("")
		fmt.Println("------------- New Structure")
		fmt.Println(output)

		numSuccess += 1
	}
	logger.Info("==================================================")
	logger.Info("Total Num Success and Failure", "Success", numSuccess, "Failure", numFailure)
	logger.Info("==================================================")

	transfer.Cleanup(cfg, logger)
}


func process(entry os.DirEntry, cfg *config.Config, logger *slog.Logger) (*metadata.Entry, error) {

	entryPath := filepath.Join(cfg.TorrentPath, entry.Name())
	if entry.IsDir() && entry.Name() == filepath.Base(cfg.IncompletePath) {
		logger.Debug("Skipping temp directory", "name", entry.Name())
		return nil, nil
	}

	root, err := parser.Parse(entryPath, logger)
	if err != nil {
		logger.Error("Parse returned error", "error", err)
		return nil, err 
	}

	err = remuxer.Remux(root, cfg, logger)
	if err != nil {
		logger.Error("Remux returned error", "error", err)
		return nil, err 
	}

	err = classifier.Classify(root, logger)
	if err != nil {
		logger.Error("Classify returned error", "error", err)
		return root, err
	}

	err = verifier.Verify(root, cfg, logger)
	if err != nil {
		logger.Error("Verify returned error", "error", err)
		return root, err
	}

	err = enricher.Enrich(root, cfg, logger)
	if err != nil {
		logger.Error("Enrich returned error", "error", err)
		return root, err
	}

	err = resolver.Resolve(root, cfg)
	if err != nil {
		logger.Error("Resolve returned error", "error", err)
		return root, err
	}

	return root, nil
}

func createTempDir() string {
	tempDir, err := os.MkdirTemp("", "media-library-manager-*")
	if err != nil {
		panic("Error creating temporary directory for dummy output")
	}
	defer os.RemoveAll(tempDir)
	return tempDir
}

func tree(path string) string {
	cmd := exec.Command("tree", "--noreport", "-C", path)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error executing tree command: %v", err)
	}
	lines := strings.SplitN(string(output), "\n", 2)
	if len(lines) == 2 {
		return filepath.Base(path) + "\n" + strings.TrimRight(lines[1], "\n")
	}
	return string(output)
}
