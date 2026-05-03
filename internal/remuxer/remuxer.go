// Package remuxer extracts subtitle tracks from MKV files using mkvmerge.
// For every FULL subtitle track whose language can be identified, an SRT file
// is written alongside the destination path: <DestPath>.<language>.srt
// Forced, SDH/CC, and signs/songs tracks are skipped to avoid duplicates of
// the same language. Runs immediately after parser.Parse, before classifier
// assigns EntryRole.
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
		Language         string `json:"language"`
		LanguageIETF     string `json:"language_ietf"`
		TrackName        string `json:"track_name"`
		ForcedTrack      bool   `json:"forced_track"`
		FlagHearingImpaired bool `json:"flag_hearing_impaired"`
		FlagVisualImpaired  bool `json:"flag_visual_impaired"`
		FlagOriginal     bool   `json:"flag_original"`
		FlagCommentary   bool   `json:"flag_commentary"`
		DefaultTrack     bool   `json:"default_track"`
	} `json:"properties"`
}

type mkvFile struct {
	Tracks []mkvTrack `json:"tracks"`
}

// Substrings in track titles that indicate a partial/specialty subtitle track
// rather than a full dialogue track. Matched case-insensitively.
var partialTitleMarkers = []string{
	"forced",
	"sign",  // matches "signs", "signs/songs", "signs & songs"
	"song",  // matches "songs", "sings/songs"
	"sdh",
	"cc",
	"closed caption",
	"hearing impaired",
	"hearing-impaired",
	"commentary",
	"karaoke",
	"description",
}

// Remux examines the entry: if it is not an MKV file it returns immediately.
// Extracts full text subtitles when possible, then ALWAYS strips all subtitle tracks.
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
		lg.Info("no subtitle tracks found, nothing to strip")
		return nil
	}

	type candidateTrack struct {
		track mkvTrack
		lang  string
		dest  string
	}

	base := strings.TrimSuffix(entry.FileInfo.DestPath, filepath.Ext(entry.FileInfo.DestPath))
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

		if reason := isPartialSubtitle(t); reason != "" {
			lg.Info("skipping non-full subtitle track",
				"id", t.ID,
				"title", t.Properties.TrackName,
				"reason", reason,
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
				"title", t.Properties.TrackName,
			)
			continue
		}

		seen[lang] = true
		candidates = append(candidates, candidateTrack{
			track: t,
			lang:  lang,
			dest:  base + "." + lang + ".srt",
		})
	}

	if cfg.DryRun {
		if len(candidates) > 0 {
			for _, c := range candidates {
				lg.Info("dry run: would extract subtitle track",
					"id", c.track.ID,
					"language", c.lang,
					"title", c.track.Properties.TrackName,
					"dest", filepath.Base(c.dest),
				)
			}
		} else {
			lg.Info("dry run: no extractable full subtitle tracks found")
		}

		lg.Info("dry run: would strip all subtitle tracks from MKV",
			"path", filepath.Base(entry.FileInfo.SourcePath),
		)
		return nil
	}

	if len(candidates) > 0 {
		for _, c := range candidates {
			if err := extractTrack(entry.FileInfo.SourcePath, c.track.ID, c.dest, lg); err != nil {
				return err
			}
			lg.Info("extracted subtitle track",
				"id", c.track.ID,
				"language", c.lang,
				"title", c.track.Properties.TrackName,
				"dest", filepath.Base(c.dest),
			)
		}
	} else {
		lg.Info("no extractable full subtitle tracks after filtering, proceeding to strip all subtitles")
	}

	// ALWAYS strip subtitles
	if err := stripSubtitles(entry.FileInfo.SourcePath, lg); err != nil {
		return err
	}

	lg.Info("stripped subtitle tracks from MKV",
		"path", filepath.Base(entry.FileInfo.SourcePath),
	)

	return nil
}

// isPartialSubtitle returns a non-empty reason string if the track is a
// partial/specialty track (forced, SDH, signs/songs, commentary, etc.) rather
// than a full dialogue track. An empty string means it's a full track.
func isPartialSubtitle(t mkvTrack) string {
	p := t.Properties

	if p.ForcedTrack {
		return "forced flag set"
	}
	if p.FlagHearingImpaired {
		return "hearing-impaired flag set"
	}
	if p.FlagVisualImpaired {
		return "visual-impaired flag set"
	}
	if p.FlagCommentary {
		return "commentary flag set"
	}

	title := strings.ToLower(p.TrackName)
	if title == "" {
		return ""
	}
	for _, marker := range partialTitleMarkers {
		if strings.Contains(title, marker) {
			return "title contains '" + marker + "'"
		}
	}

	return ""
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

// stripSubtitles removes ALL subtitle tracks from the MKV
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
