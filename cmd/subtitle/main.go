package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/detector"
	"github.com/ENIACore/media_library_manager/internal/enhancer"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func main() {
	cfg := config.Load()
	lg := logger.NewLogger(cfg, "subtitle")
	subtitle(cfg, lg)
}

func subtitle(cfg *config.Config, logger *slog.Logger) {
	cfg.Interactive = false

	var session *enhancer.Session

	if cached := enhancer.LoadCachedSession(cfg); cached != nil {
		valid, err := enhancer.VerifySession(cached, logger)
		if err != nil {
			logger.Warn("session verification failed, falling back to login", "error", err)
		} else if valid {
			logger.Info("using cached OpenSubtitles session")
			session = cached
		}
	}

	if session == nil {
		var err error
		session, err = enhancer.Login(cfg, logger)
		if err != nil {
			logger.Error("OpenSubtitles login failed", "error", err)
			return
		}
		if err := enhancer.SaveSession(session, cfg); err != nil {
			logger.Warn("failed to cache OpenSubtitles session", "error", err)
		}
	}

	if cfg.SubtitlePath != "" {
		if _, err := processLibrary(cfg.SubtitlePath, 0, session, cfg, logger); err != nil {
			logger.Error("failed to process path", "error", err)
		}
		return
	}

	count, err := processLibrary(cfg.MoviePath, 0, session, cfg, logger)
	if err != nil {
		logger.Error("failed to process movie library", "error", err)
		return
	}
	if _, err := processLibrary(cfg.ShowPath, count, session, cfg, logger); err != nil {
		logger.Error("failed to process show library", "error", err)
	}
}

func processLibrary(libraryPath string, count int, session *enhancer.Session, cfg *config.Config, logger *slog.Logger) (int, error) {
	entries, err := os.ReadDir(libraryPath)
	if err != nil {
		return count, fmt.Errorf("unable to read library path %q: %w", libraryPath, err)
	}

	for _, entry := range entries {
		if cfg.OverLimit(count) {
			return count, nil
		}

		entryPath := filepath.Join(libraryPath, entry.Name())

		paths, err := detector.DetectSubtitle(entryPath, logger)
		if err != nil {
			logger.Error("DetectSubtitle returned error", "error", err)
			continue
		}
		if len(paths) == 0 {
			continue
		}

		for _, videoPath := range paths {
			if cfg.OverLimit(count) {
				return count, nil
			}

			mediaEntry := buildEntry(videoPath, logger)
			if mediaEntry.MediaInfo.TMDBid == 0 {
				logger.Warn("no TMDBid found in directory path, skipping", "path", videoPath)
				continue
			}

			time.Sleep(2 * time.Second)
			if err := enhancer.FetchSubtitle(mediaEntry, session, cfg, logger); err != nil {
				logger.Error("FetchSubtitle returned error", "error", err)
				continue
			}
			count++
		}
	}

	return count, nil
}

func buildEntry(videoPath string, logger *slog.Logger) *metadata.Entry {
	mediaInfo := extractor.ExtractMedia(videoPath, logger)
	mediaInfo.TMDBid = extractor.ExtractTMDBid(videoPath)

	ext := filepath.Ext(videoPath)
	subtitlePath := strings.TrimSuffix(videoPath, ext) + ".English.srt"

	role := metadata.MovieFile
	if mediaInfo.Season != nil || mediaInfo.Episode != nil {
		role = metadata.EpisodeFile
	}

	return &metadata.Entry{
		Role:      role,
		MediaInfo: mediaInfo,
		FileInfo: metadata.FileInfo{
			SourcePath:  videoPath,
			DestPath:    subtitlePath,
			Ext:         strings.ToUpper(strings.TrimPrefix(ext, ".")),
			ContentType: metadata.Video,
		},
	}
}
