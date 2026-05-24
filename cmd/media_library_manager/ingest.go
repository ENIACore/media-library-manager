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
	"github.com/ENIACore/media_library_manager/internal/parser"
	"github.com/ENIACore/media_library_manager/internal/remuxer"
	"github.com/ENIACore/media_library_manager/internal/verifier"
	"github.com/ENIACore/media_library_manager/internal/resolver"
	"github.com/ENIACore/media_library_manager/internal/transfer"
)

func ingest(tempDir string, cfg *config.Config, logger *slog.Logger) {

	entries, err := os.ReadDir(cfg.TorrentPath)
	if err != nil {
		panic("unable to read from torrent path")
	}

	numFailure := 0
	numSuccess := 0
	var processedNames []string

	for i, entry := range entries {

		if overLimit(i, cfg) {
			break
		}

		root, err := process(entry, cfg, logger)

		if err != nil {
			numFailure += 1
			processedNames = append(processedNames, entry.Name())
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

		transfer.Transfer(root, cfg, logger)

		remuxFiles(root, cfg, logger)

		if !root.FileInfo.IsDir {
      		root.FileInfo.DestPath = filepath.Dir(root.FileInfo.DestPath)
  		}
		output = tree(root.FileInfo.DestPath)
		fmt.Println("")
		fmt.Println("------------- New Structure")
		fmt.Println(output)

		processedNames = append(processedNames, entry.Name())
		numSuccess += 1
	}
	logger.Info("==================================================")
	logger.Info("Total Num Success and Failure", "Success", numSuccess, "Failure", numFailure)
	logger.Info("==================================================")

	transfer.Cleanup(cfg, logger, processedNames)
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

	err = enricher.Enrich(root, logger)
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

func remuxFiles(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) {
	entry.FileInfo.SourcePath = entry.FileInfo.DestPath // File now exists at destination
	err := remuxer.Remux(entry, cfg, logger)
	if err != nil {
		logger.Error("Remux returned error", "error", err)
		return
	}

	err = remuxer.NormalizeAudioTracks(entry, cfg, logger)
	if err != nil {
		logger.Error("Normalize audio tracks returned error", "error", err)
		return
	}

	for _, child := range entry.Children {
		remuxFiles(child, cfg, logger)	
	}
}
