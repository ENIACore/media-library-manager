package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/classifier"
	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/parser"
	"github.com/ENIACore/media_library_manager/internal/processor"
)

func main() {
	cfg := config.Load()
	logger := logger.NewLogger(cfg)

	entries, err := os.ReadDir(cfg.MediaPath)
	if err != nil {
		logger.Error("unable to read from media path", "error", err)
		panic("unable to read from media path")
	}

	errorPath := filepath.Join(cfg.ManagerPath, "error")
	totalEntries := len(entries)

	for idx, entry := range entries {
		path := filepath.Join(cfg.MediaPath, entry.Name())

		// Visual separator for each entry
		separator := strings.Repeat("=", 80)
		logger.Info(separator)
		logger.Info("PROCESSING ENTRY", "number", idx+1, "of", totalEntries, "name", entry.Name())
		logger.Info(separator)

		// Parse phase
		logger.Info("PHASE: PARSING", "path", path)
		root, err := parser.ParseTree(path, nil, 0, logger)
		if err != nil && cfg.DryRun {
			logger.Error("❌ PARSING FAILED (dry-run skip)", "error", err)
			logger.Info(strings.Repeat("-", 80))
			continue
		}
		if err != nil {
			logger.Error("❌ PARSING FAILED", "error", err)
			processor.Error(root, errorPath, logger)
			logger.Info(strings.Repeat("-", 80))
			continue
		}
		logger.Info("✓ Parsing completed")

		// Classification phase
		logger.Info("PHASE: CLASSIFICATION")
		err = classifier.ClassifyEntries(root, logger)
		if err != nil && cfg.DryRun {
			logger.Error("❌ CLASSIFICATION FAILED (dry-run skip)", "error", err)
			logger.Info(strings.Repeat("-", 80))
			continue
		}
		if err != nil {
			logger.Error("❌ CLASSIFICATION FAILED", "error", err)
			processor.Error(root, errorPath, logger)
			logger.Info(strings.Repeat("-", 80))
			continue
		}
		logger.Info("✓ Classification completed", "role", roleToString(root.Role))

		// Resolution phase
		logger.Info("PHASE: RESOLUTION")
		err = processor.ResolveEntries(root, logger)
		if err != nil && cfg.DryRun {
			logger.Error("❌ RESOLUTION FAILED (dry-run skip)", "error", err)
			logger.Info(strings.Repeat("-", 80))
			continue
		}
		if err != nil {
			logger.Error("❌ RESOLUTION FAILED", "error", err)
			processor.Error(root, errorPath, logger)
			logger.Info(strings.Repeat("-", 80))
			continue
		}
		logger.Info("✓ Resolution completed", "dest", root.PathInfo.Dest)

		// Determine final library path
		finalLibraryPath := cfg.LibraryPath
		switch root.Role {
		case metadata.MovieDir, metadata.MovieFile:
			finalLibraryPath = filepath.Join(finalLibraryPath, "movies")
		case metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile:
			finalLibraryPath = filepath.Join(finalLibraryPath, "shows")
		}

		fullDestPath := filepath.Join(finalLibraryPath, root.PathInfo.Dest)

		// Transfer/dry-run phase
		if cfg.DryRun {
			logger.Info("✓ DRY RUN SUCCESS", 
				"source", root.PathInfo.Source, 
				"dest", fullDestPath, 
				"type", roleToString(root.Role))
			logger.Info(strings.Repeat("-", 80))
			continue
		}

		logger.Info("PHASE: TRANSFER")
		err = processor.Transfer(root, cfg.LibraryPath, logger)
		if err != nil {
			logger.Error("❌ TRANSFER FAILED", "error", err)
			processor.Error(root, errorPath, logger)
			logger.Info(strings.Repeat("-", 80))
			continue
		}

		logger.Info("✓ TRANSFER SUCCESS", 
			"source", root.PathInfo.Source, 
			"dest", fullDestPath)
		logger.Info(strings.Repeat("-", 80))
	}

	logger.Info(strings.Repeat("=", 80))
	logger.Info("PROCESSING COMPLETE", "total", totalEntries)
	logger.Info(strings.Repeat("=", 80))
}

func roleToString(role metadata.EntryRole) string {
	switch role {
		case metadata.MovieFile:
			return "MOVIE_FILE"
		case metadata.MovieDir:
			return "MOVIE_DIR"
		case metadata.EpisodeFile:
			return "EPISODE_FILE"
		case metadata.SeasonDir:
			return "SEASON_DIR"
		case metadata.SeriesDir:
			return "SERIES_DIR"
		default:
			return "UNKNOWN"
	}
}
