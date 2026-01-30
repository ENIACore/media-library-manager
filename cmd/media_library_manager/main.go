// Package main is entry point for media library manager application.
// Orchestrates processing pipeline: [parser], [classifier], [enricher], [resolver], and [transfer].
package main

import (
	"os"
	"path/filepath"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/parser"
	"github.com/ENIACore/media_library_manager/internal/classifier"
	"github.com/ENIACore/media_library_manager/internal/enricher"
	"github.com/ENIACore/media_library_manager/internal/resolver"
	"github.com/ENIACore/media_library_manager/internal/transfer"
)


// main loads configuration and processes each torrent directory entry through processing pipeline.
// Failed entries are moved to error directory. Logs summary of successful and failed processing counts.
func main() {
	cfg := config.Load()
	logger := logger.NewLogger(cfg)

	entries, err := os.ReadDir(cfg.TorrentPath)
	if err != nil {
		panic("unable to read from torrent path")
	}

	numSuccess := 0
	numFailure := 0
	for _, entry := range entries {
		logger.Info("")
		logger.Info("==================================================")
		logger.Info("")

		entryPath := filepath.Join(cfg.TorrentPath, entry.Name())
		if entry.IsDir() && entry.Name() == filepath.Base(cfg.IncompletePath) {
        	logger.Debug("Skipping temp directory", "name", entry.Name())
        	continue
    	}

		root, err := parser.Parse(entryPath, logger)
		if err != nil {
			logger.Error("Parse returned error", "error", err)
			numFailure += 1
			transfer.Error(root, cfg, logger)
			continue
		}

		err = classifier.Classify(root, logger)
		if err != nil {
			logger.Error("Classify returned error", "error", err)
			numFailure += 1
			transfer.Error(root, cfg, logger)
			continue
		}

		err = enricher.Enrich(root, cfg, logger)
		if err != nil {
			logger.Error("Enrich returned error", "error", err)
			numFailure += 1
			transfer.Error(root, cfg, logger)
			continue
		}

		err = resolver.Resolve(root, cfg, logger)
		if err != nil {
			logger.Error("Resolve returned error", "error", err)
			numFailure += 1
			transfer.Error(root, cfg, logger)
			continue
		}

		err = transfer.Transfer(root, cfg, logger)
		if err != nil {
			logger.Error("Transfer returned error", "error", err)
			numFailure += 1
			transfer.Error(root, cfg, logger)
			continue
		}
		numSuccess += 1

		logger.Info("")
		logger.Info("==================================================")
		logger.Info("")
	}
	
	
	logger.Info("")
	logger.Info("==================================================")
	logger.Info("Total Num Success and Failure", "Success", numSuccess, "Failure", numFailure)
	logger.Info("==================================================")
	logger.Info("")

	transfer.Cleanup(cfg, logger)
}
