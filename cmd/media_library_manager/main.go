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

		fmt.Println("source is: ", root.Source())
		fmt.Println("title is: ", root.MediaInfo.Title)
		fmt.Println("year is: ", root.MediaInfo.YearString())
		fmt.Println("episode is: ", root.MediaInfo.EpisodeString())
		fmt.Println("season is: ", root.MediaInfo.SeasonString())
		fmt.Println("deleted scenes is: ", root.MediaInfo.DS)
		fmt.Println("behind the scenes is: ", root.MediaInfo.BTS)
		fmt.Println("bonus is: ", root.MediaInfo.Bonus)

		fmt.Println("source path is: ", root.FileInfo.SourcePath)
		fmt.Println("ext is: ", root.FileInfo.Ext)
		fmt.Println("content type is: ", root.FileInfo.ContentType)
		fmt.Println("is dir is: ", root.FileInfo.IsDir)
		fmt.Println("resolution is: ", root.FileInfo.Resolution)
		fmt.Println("codec is: ", root.FileInfo.Codec)
		fmt.Println("audio is: ", root.FileInfo.Audio)
		fmt.Println("language is: ", root.FileInfo.Language)
		fmt.Println("bitrate is: ", root.FileInfo.Bitrate)
	}
}
