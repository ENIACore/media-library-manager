// Package remuxer fixes subtitle track metadata in MKV files using mkvmerge.
// All subtitle tracks are kept; each track's language tag and title are normalised,
// and tracks flagged as hearing-impaired are named "<Language> SDH".
// Runs immediately after parser.Parse, before classifier assigns EntryRole.
package remuxer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

type mkvTrack struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Properties struct {
		Language            string `json:"language"`
		TrackName           string `json:"track_name"`
		HearingImpairedFlag bool   `json:"hearing_impaired_flag"`
	} `json:"properties"`
}

type mkvFile struct {
	Tracks []mkvTrack `json:"tracks"`
}

// keyToISO639 maps language pattern keys to ISO 639-2/B codes for mkvmerge --language.
var keyToISO639 = map[string]string{
	"Arabic":                 "ara",
	"Basque":                 "baq",
	"Brazilian-Portuguese":   "por",
	"Canadian-French":        "fre",
	"Catalan":                "cat",
	"Chinese":                "chi",
	"Croatian":               "hrv",
	"Czech":                  "cze",
	"Danish":                 "dan",
	"Dutch":                  "dut",
	"English":                "eng",
	"Filipino":               "fil",
	"Finnish":                "fin",
	"French":                 "fre",
	"Galician":               "glg",
	"German":                 "ger",
	"Greek":                  "gre",
	"Hebrew":                 "heb",
	"Hungarian":              "hun",
	"Indonesian":             "ind",
	"Italian":                "ita",
	"Japanese":               "jpn",
	"Korean":                 "kor",
	"Latin-American-Spanish": "spa",
	"Malay":                  "may",
	"Norwegian":              "nor",
	"Polish":                 "pol",
	"Portuguese":             "por",
	"Romanian":               "rum",
	"Russian":                "rus",
	"Simplified-Chinese":     "chi",
	"Spanish":                "spa",
	"Swedish":                "swe",
	"Tagalog":                "tgl",
	"Thai":                   "tha",
	"Traditional-Chinese":    "chi",
	"Turkish":                "tur",
	"Ukrainian":              "ukr",
	"Vietnamese":             "vie",
}

var sdhNameRe = regexp.MustCompile(`(?i)\bSDH\b|\bHI\b|hearing.impaired`)

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

// isSDH returns true if the track is flagged hearing-impaired or its name
// contains a common SDH/HI marker.
func isSDH(t mkvTrack) bool {
	return t.Properties.HearingImpairedFlag || sdhNameRe.MatchString(t.Properties.TrackName)
}

// trackName builds the normalised display name for a subtitle track.
// Returns "" when the language is unrecognised so callers can skip renaming.
func trackName(t mkvTrack) string {
	lang := extractor.ParseLanguage([]string{strings.ToUpper(t.Properties.Language)})
	if lang == "" {
		return ""
	}
	if isSDH(t) {
		return lang + " SDH"
	}
	return lang
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

	// Build per-track metadata corrections. Tracks whose language is
	// unrecognised are still kept but not renamed.
	type trackFix struct {
		track   mkvTrack
		name    string // "" = no rename
		isoCode string // "" = no language correction
	}
	fixes := make([]trackFix, len(subtitles))
	for i, t := range subtitles {
		name := trackName(t)
		lang := extractor.ParseLanguage([]string{strings.ToUpper(t.Properties.Language)})
		fixes[i] = trackFix{track: t, name: name, isoCode: keyToISO639[lang]}
	}

	if cfg.DryRun {
		for _, f := range fixes {
			if f.name != "" {
				lg.Info("dry run: would rename subtitle track", "id", f.track.ID, "name", f.name, "iso", f.isoCode)
			} else {
				lg.Info("dry run: would keep subtitle track unchanged (unrecognised language)", "id", f.track.ID, "lang", f.track.Properties.Language)
			}
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
	for _, f := range fixes {
		id := fmt.Sprintf("%d", f.track.ID)
		if f.name != "" {
			args = append(args, "--track-name", id+":"+f.name)
		}
		if f.isoCode != "" {
			args = append(args, "--language", id+":"+f.isoCode)
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

	lg.Info("subtitle tracks fixed", "path", filepath.Base(path), "tracks", len(subtitles))
	return nil
}
