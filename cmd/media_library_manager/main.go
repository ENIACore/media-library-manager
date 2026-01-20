package main

import (
	"os"
	"path/filepath"

	"github.com/ENIACore/media_library_manager/internal/classifier"
	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/parser"
	"github.com/ENIACore/media_library_manager/internal/processor"
	"strings"
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

	for _, entry := range entries {
		path := filepath.Join(cfg.MediaPath, entry.Name())

		separator := strings.Repeat("=", 80)
		logger.Info("")
		logger.Info("")
		logger.Info(separator)
		logger.Info("")
		logger.Info("")
		logger.Info("processing root entry", "entry", path)

		root, err := parser.ParseTree(path, nil, 0, logger)
		if err != nil && cfg.DryRun {
			logger.Error("error occurred, skipping error processing due to dry run", "error", err)
			continue
		}
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}

		err = classifier.ClassifyEntries(root, logger)
		if err != nil && cfg.DryRun {
			logger.Error("error occurred, skipping error processing due to dry run", "error", err)
			continue
		}
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}

		err = processor.ResolveEntries(root, logger)
		if err != nil && cfg.DryRun {
			logger.Error("error occurred, skipping error processing due to dry run", "error", err)
			continue
		}
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}

		finalLibraryPath := cfg.LibraryPath
		switch root.Role {
			case metadata.MovieDir, metadata.MovieFile:
				finalLibraryPath = filepath.Join(finalLibraryPath, "movies")
			case metadata.SeriesDir, metadata.SeasonDir, metadata.EpisodeFile:
				finalLibraryPath = filepath.Join(finalLibraryPath, "shows")
		}


		if cfg.DryRun {
			logger.Info("successfully processed media, no action due to dry run", "source", root.PathInfo.Source, "dest", filepath.Join(finalLibraryPath, root.PathInfo.Dest), "classification", root.Role) 
			logger.Info("")
			logger.Info("")
			logger.Info(separator)
			logger.Info("")
			logger.Info("")
			continue
		}

		logger.Info("successfully processed media", "source", root.PathInfo.Source, "dest", filepath.Join(finalLibraryPath, root.PathInfo.Dest), "classification", root.Role) 
		logger.Info("")
		logger.Info("")
		logger.Info(separator)
		logger.Info("")
		logger.Info("")
		err = processor.Transfer(root, cfg.LibraryPath, logger)
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}
	}
}
