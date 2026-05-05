package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/detector"
	"github.com/ENIACore/media_library_manager/internal/enhancer"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func subtitle(cfg *config.Config, logger *slog.Logger) {
	cfg.Interactive = false // If processing subtitles, it is assumed the existing title/year/tmdbid is accurate

	var session *enhancer.Session

	if cached := enhancer.LoadCachedSession(cfg); cached != nil {
		valid, err := enhancer.VerifySession(cached, cfg, logger)
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

	count := processLibrary(cfg.MoviePath, 0, session, cfg, logger)
	processLibrary(cfg.ShowPath, count, session, cfg, logger)
}

func processLibrary(libraryPath string, count int, session *enhancer.Session, cfg *config.Config, logger *slog.Logger) int {

	entries, err := os.ReadDir(libraryPath)
	if err != nil {
		panic("unable to read from library path: " + err.Error())
	}

	for _, entry := range entries {
		if overLimit(count, cfg) {
			return count
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
			if overLimit(count, cfg) {
				return count
			}

			mediaEntry := buildEntry(videoPath, logger)
			if mediaEntry.MediaInfo.TMDBid == 0 {
				logger.Warn("no TMDBid found in directory path, skipping", "path", videoPath)
				continue
			}

			time.Sleep(2 * time.Second)
			count++
			if err := enhancer.FetchSubtitle(mediaEntry, session, cfg, logger); err != nil {
				logger.Error("FetchSubtitle returned error", "error", err)
			}
		}
	}

	return count
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

