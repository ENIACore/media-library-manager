// Package remuxer extracts subtitle tracks from MKV files using mkvmerge.
// For every subtitle track whose language can be identified, an SRT file is
// written alongside the destination path: <DestPath>.<language>.srt
// Runs immediately after parser.Parse, before classifier assigns EntryRole.
package remuxer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

type mkvTrack struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Codec      string `json:"codec"`
	Properties struct {
		Language string `json:"language"`
		TrackName string `json:"track_name"`
	} `json:"properties"`
}

type mkvFile struct {
	Tracks []mkvTrack `json:"tracks"`
}

// Remux examines the entry: if it is not an MKV file it returns immediately.
// For each subtitle track that is S_TEXT/UTF8 and whose language can be
// identified, it extracts the track to <DestPath>.<language>.srt — taking
// only the first track when multiple share the same language. Subtitle tracks
// are then stripped from the source MKV in-place.
// When cfg.DryRun is true all operations are logged but not executed.
func Remux(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Remux", "source", entry.Source())

	if entry.FileInfo.IsDir || entry.FileInfo.Ext != "MKV" {
		return nil
	}

	subtitles, err := identifySubtitleTracks(entry.FileInfo.SourcePath)
	if err != nil {
		return err
	}
	if len(subtitles) == 0 {
		lg.Info("no subtitle tracks, skipping")
		return nil
	}

	// Filter to S_TEXT/UTF8 tracks with a recognisable language, keeping only
	// the first track encountered per language.
	type candidateTrack struct {
		track mkvTrack
		lang  string
		dest  string
	}
	seen := make(map[string]bool)
	var candidates []candidateTrack

	for _, t := range subtitles {
		if !isTextSubtitle(t.Codec) {
			lg.Info("skipping subtitle track with unsupported codec",
				"id", t.ID,
				"codec", t.Codec,
			)
			continue
		}
		lang := extractor.ParseLanguage([]string{strings.ToUpper(t.Properties.Language)})
		if lang == "" {
			lg.Info("skipping subtitle track with unrecognised language",
				"id", t.ID,
				"raw_lang", t.Properties.Language,
			)
			continue
		}
		if seen[lang] {
			lg.Info("skipping duplicate subtitle track for language",
				"id", t.ID,
				"language", lang,
			)
			continue
		}
		seen[lang] = true
		candidates = append(candidates, candidateTrack{
			track: t,
			lang:  lang,
			dest:  entry.FileInfo.DestPath + "." + lang + ".srt",
		})
	}

	if len(candidates) == 0 {
		lg.Info("no extractable subtitle tracks after filtering, skipping")
		return nil
	}

	if cfg.DryRun {
		for _, c := range candidates {
			lg.Info("dry run: would extract subtitle track",
				"id", c.track.ID,
				"language", c.lang,
				"dest", filepath.Base(c.dest),
			)
		}
		lg.Info("dry run: would strip all subtitle tracks from MKV",
			"path", filepath.Base(entry.FileInfo.SourcePath),
		)
		return nil
	}

	// Extract each candidate track to its SRT file.
	for _, c := range candidates {
		if err := extractTrack(entry.FileInfo.SourcePath, c.track.ID, c.dest, lg); err != nil {
			return err
		}
		lg.Info("extracted subtitle track",
			"id", c.track.ID,
			"language", c.lang,
			"dest", filepath.Base(c.dest),
		)
	}

	// Strip all subtitle tracks from the MKV in-place.
	if err := stripSubtitles(entry.FileInfo.SourcePath, lg); err != nil {
		return err
	}
	lg.Info("stripped subtitle tracks from MKV",
		"path", filepath.Base(entry.FileInfo.SourcePath),
	)

	return nil
}

func identifySubtitleTracks(path string) ([]mkvTrack, error) {
	cmd := exec.Command("mkvmerge", "--identify", "--identification-format", "json", path)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			return nil, fmt.Errorf("remuxer: mkvmerge identify failed for %q: %w", path, err)
		}
	}
	var info mkvFile
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("remuxer: failed to parse mkvmerge identify output for %q: %w", path, err)
	}
	var subtitles []mkvTrack
	for _, t := range info.Tracks {
		if t.Type == "subtitles" {
			subtitles = append(subtitles, t)
		}
	}
	return subtitles, nil
}

func extractTrack(srcPath string, trackID int, destSRT string, lg *slog.Logger) error {
	args := []string{
		"tracks",
		srcPath,
		fmt.Sprintf("%d:%s", trackID, destSRT),
	}
	out, err := exec.Command("mkvextract", args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			return fmt.Errorf("remuxer: mkvextract failed for track %d of %q: %w\n%s",
				trackID, srcPath, err, out)
		}
		lg.Warn("mkvextract completed with warnings", "output", string(out))
	}
	return nil
}

// stripSubtitles rewrites the MKV at path with all subtitle tracks removed,
// replacing the original file in-place via a temp file.
func stripSubtitles(path string, lg *slog.Logger) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "remux-*.mkv")
	if err != nil {
		return fmt.Errorf("remuxer: failed to create temp file for %q: %w", path, err)
	}
	tmpPath := tmp.Name()
	tmp.Close()

	args := []string{"-o", tmpPath, "--no-subtitles", path}
	out, err := exec.Command("mkvmerge", args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			os.Remove(tmpPath)
			return fmt.Errorf("remuxer: mkvmerge strip failed for %q: %w\n%s", path, err, out)
		}
		lg.Warn("mkvmerge completed with warnings", "output", string(out))
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("remuxer: failed to replace %q with stripped file: %w", path, err)
	}
	return nil
}

func isTextSubtitle(codec string) bool {
    switch codec {
    case "S_TEXT/UTF8", "SubRip/SRT", "S_TEXT/ASS", "SubStationAlpha", "S_TEXT/SSA":
        return true
    }
    return false
}
