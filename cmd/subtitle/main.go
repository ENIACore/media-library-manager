package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ENIACore/media_library_manager/internal/classifier"
	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/detector"
	"github.com/ENIACore/media_library_manager/internal/enhancer"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"github.com/ENIACore/media_library_manager/internal/logger"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/parser"
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

	count, err := processLibraryStrict(cfg.MoviePath, 0, session, cfg, logger)
	if err != nil {
		logger.Error("failed to process movie library", "error", err)
		return
	}
	if _, err := processLibraryStrict(cfg.ShowPath, count, session, cfg, logger); err != nil {
		logger.Error("failed to process show library", "error", err)
	}
}

// processLibraryStrict processes libraryPath (cfg.MoviePath or cfg.ShowPath) one top-level
// entry at a time, where each entry is a single already-ingested movie or show directory.
// Unlike processLibrary, it runs the same parser+classifier pipeline ingest uses so every
// file is classified with full tree context (season/extras directories, siblings, etc.)
// instead of the single-file heuristic in buildEntry. Extras/DS/BTS files are still attempted
// here (and will fail with "unsupported entry role") since filtering them out is a follow-up.
func processLibraryStrict(libraryPath string, count int, session *enhancer.Session, cfg *config.Config, logger *slog.Logger) (int, error) {
	entries, err := os.ReadDir(libraryPath)
	if err != nil {
		return count, fmt.Errorf("unable to read library path %q: %w", libraryPath, err)
	}

	for _, entry := range entries {
		if cfg.OverLimit(count) {
			return count, nil
		}

		entryPath := filepath.Join(libraryPath, entry.Name())

		root, err := parser.Parse(entryPath, logger)
		if err != nil {
			logger.Error("Parse returned error", "error", err, "path", entryPath)
			continue
		}

		if err := classifier.Classify(root, logger); err != nil {
			logger.Error("Classify returned error", "error", err, "path", entryPath)
			continue
		}

		for _, mediaEntry := range collectVideoFiles(root) {
			if cfg.OverLimit(count) {
				return count, nil
			}

			mediaEntry.MediaInfo.TMDBid = extractor.ExtractTMDBid(mediaEntry.FileInfo.SourcePath)
			if mediaEntry.MediaInfo.TMDBid == 0 {
				logger.Warn("no TMDBid found in directory path, skipping", "path", mediaEntry.FileInfo.SourcePath)
				continue
			}

			ext := filepath.Ext(mediaEntry.FileInfo.SourcePath)
			mediaEntry.FileInfo.DestPath = strings.TrimSuffix(mediaEntry.FileInfo.SourcePath, ext) + ".English.srt"
			if _, err := os.Stat(mediaEntry.FileInfo.DestPath); err == nil {
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

// collectVideoFiles returns every video leaf in a classified tree, regardless of role.
func collectVideoFiles(entry *metadata.Entry) []*metadata.Entry {
	if !entry.FileInfo.IsDir {
		if entry.FileInfo.ContentType == metadata.Video {
			return []*metadata.Entry{entry}
		}
		return nil
	}

	var out []*metadata.Entry
	for _, child := range entry.Children {
		out = append(out, collectVideoFiles(child)...)
	}
	return out
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
