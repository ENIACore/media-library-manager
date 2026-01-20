package main

import (
	"github.com/ENIACore/media_library_manager/internal/processor"
	"github.com/ENIACore/media_library_manager/internal/classifier"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/parser"
	"os"
	"path/filepath"
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

		root, err := parser.ParseTree(path, nil, 0, logger)
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}

		err = classifier.ClassifyEntries(root, logger)
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}

		err = processor.ResolveEntries(root, logger)
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}

		if cfg.DryRun {
			logger.Info("successfully processed media, no action due to dry run", "source", root.PathInfo.Source, "dest", root.PathInfo.Dest, "classification", root.Role) 
			continue
		}

		err = processor.Transfer(root, cfg.LibraryPath, logger)
		if err != nil {
			processor.Error(root, errorPath, logger)	
			continue
		}
	}
}
