package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

	count := processLibrary(cfg.MoviePath, cfg, logger)
	if count > 80 {
		return
	}
	processLibrary(cfg.ShowPath, cfg, logger)
}

func processLibrary(libraryPath string, cfg *config.Config, logger *slog.Logger) int {
	count := 0
	entries, err := os.ReadDir(libraryPath)
	if err != nil {
		panic("unable to read from library path: " + err.Error())
	}

	for _, entry := range entries {
		if count > 80 {
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

		session, err := enhancer.Login(cfg, logger)
		if err != nil {
			logger.Error("OpenSubtitles login failed", "error", err)
			continue
		}

		for _, videoPath := range paths {
			if count > 80 {
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

var tmdbIDRe = regexp.MustCompile(`\[tmdb-(\d+)\]`)

func buildEntry(videoPath string, logger *slog.Logger) *metadata.Entry {
	mediaInfo := extractor.ExtractMedia(videoPath, logger)
	mediaInfo.TMDBid = extractTMDBid(videoPath)

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

func extractTMDBid(path string) int {
	dir := filepath.Dir(path)
	for {
		if m := tmdbIDRe.FindStringSubmatch(filepath.Base(dir)); m != nil {
			if id, err := strconv.Atoi(m[1]); err == nil {
				return id
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return 0
}
