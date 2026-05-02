// Package remuxer fixes subtitle tracks in MKV files using mkvmerge.
// Non-English subtitles are dropped; English tracks are kept and renamed to just "English".
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
	Properties struct {
		Language  string `json:"language"`
		TrackName string `json:"track_name"`
	} `json:"properties"`
}

type mkvFile struct {
	Tracks []mkvTrack `json:"tracks"`
}

// Remux fixes subtitle tracks in all MKV files found in the entry tree.
// The original file is replaced in-place. When cfg.DryRun is true, the
// operation is logged but not executed.
func Remux(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "Remux", "source", root.Source())
	return remuxTree(root, cfg, lg)
}

func remuxTree(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	if !entry.FileInfo.IsDir && entry.FileInfo.Ext == "MKV" {
		if err := remuxFile(entry.FileInfo.SourcePath, cfg, logger); err != nil {
			return err
		}
	}
	for _, child := range entry.Children {
		if err := remuxTree(child, cfg, logger); err != nil {
			return err
		}
	}
	return nil
}

func identifySubtitleTracks(path string) ([]mkvTrack, error) {
	cmd := exec.Command("mkvmerge", "--identify", "--identification-format", "json", path)
	out, err := cmd.Output()
	if err != nil {
		// exit code 1 means warnings — output is still valid JSON
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

func remuxFile(path string, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "remuxFile", "path", filepath.Base(path))

	subtitles, err := identifySubtitleTracks(path)
	if err != nil {
		return err
	}

	if len(subtitles) == 0 {
		lg.Info("no subtitle tracks, skipping")
		return nil
	}

	var english []mkvTrack
	for _, t := range subtitles {
		if extractor.ParseLanguage([]string{strings.ToUpper(t.Properties.Language)}) == "English" {
			english = append(english, t)
		}
	}

	if cfg.DryRun {
		if len(english) == 0 {
			lg.Info("dry run: would strip all subtitles (no English tracks found)")
		} else {
			ids := make([]string, len(english))
			for i, t := range english {
				ids[i] = fmt.Sprintf("%d", t.ID)
			}
			lg.Info("dry run: would fix subtitles", "english_track_ids", strings.Join(ids, ","), "dropped", len(subtitles)-len(english))
		}
		return nil
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "remux-*.mkv")
	if err != nil {
		return fmt.Errorf("remuxer: failed to create temp file for %q: %w", path, err)
	}
	tmpPath := tmp.Name()
	tmp.Close()

	args := []string{"-o", tmpPath}
	if len(english) == 0 {
		args = append(args, "--no-subtitles")
	} else {
		ids := make([]string, len(english))
		for i, t := range english {
			ids[i] = fmt.Sprintf("%d", t.ID)
		}
		args = append(args, "--subtitle-tracks", strings.Join(ids, ","))
		for _, t := range english {
			args = append(args, "--track-name", fmt.Sprintf("%d:English", t.ID))
			args = append(args, "--language", fmt.Sprintf("%d:eng", t.ID))
		}
	}
	args = append(args, path)

	out, err := exec.Command("mkvmerge", args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			os.Remove(tmpPath)
			return fmt.Errorf("remuxer: mkvmerge failed for %q: %w\n%s", path, err, out)
		}
		lg.Warn("mkvmerge completed with warnings", "output", string(out))
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("remuxer: failed to replace %q with remuxed file: %w", path, err)
	}

	lg.Info("subtitles fixed", "path", filepath.Base(path), "kept", len(english), "dropped", len(subtitles)-len(english))
	return nil
}
