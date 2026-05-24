package remuxer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/ENIACore/media_library_manager/internal/config"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

// keyToISO639 maps language pattern Keys to ISO 639-2 codes for mkvpropedit.
var keyToISO639 = map[string]string{
	"English":               "eng",
	"Spanish":               "spa",
	"Latin-American-Spanish": "spa",
	"French":                "fra",
	"Canadian-French":       "fra",
	"German":                "deu",
	"Italian":               "ita",
	"Portuguese":            "por",
	"Brazilian-Portuguese":  "por",
	"Russian":               "rus",
	"Japanese":              "jpn",
	"Korean":                "kor",
	"Arabic":                "ara",
	"Hebrew":                "heb",
	"Thai":                  "tha",
	"Turkish":               "tur",
	"Greek":                 "ell",
	"Polish":                "pol",
	"Hungarian":             "hun",
	"Czech":                 "ces",
	"Chinese":               "zho",
	"Simplified-Chinese":    "zho",
	"Traditional-Chinese":   "zho",
	"Dutch":                 "nld",
	"Danish":                "dan",
	"Finnish":               "fin",
	"Swedish":               "swe",
	"Norwegian":             "nor",
	"Croatian":              "hrv",
	"Romanian":              "ron",
	"Ukrainian":             "ukr",
	"Vietnamese":            "vie",
	"Indonesian":            "ind",
	"Malay":                 "msa",
	"Filipino":              "fil",
}

// NormalizeAudioTracks sets each audio track's title to its language name and
// corrects the language tag, using mkvpropedit for an in-place edit.
// Language is resolved from the track's language tag first, then its title.
func NormalizeAudioTracks(entry *metadata.Entry, cfg *config.Config, logger *slog.Logger) error {
	lg := logger.With("func", "NormalizeAudioTracks", "source", entry.Source())

	if entry.FileInfo.IsDir || entry.FileInfo.Ext != "MKV" {
		return nil
	}

	audioTracks, err := identifyAudioTracks(entry.FileInfo.SourcePath)
	if err != nil {
		return err
	}

	if len(audioTracks) == 0 {
		return nil
	}

	// Build mkvpropedit args: one --edit + --set pair per track.
	// track:@N addresses the Nth track (1-based, counting all track types).
	var editArgs []string
	for _, t := range audioTracks {
		key := resolveAudioLanguage(t)
		if key == "" {
			lg.Warn("could not determine language for audio track, skipping",
				"id", t.ID,
				"language", t.Properties.Language,
				"title", t.Properties.TrackName,
			)
			continue
		}

		isoCode := keyToISO639[key]

		lg.Info("normalizing audio track",
			"id", t.ID,
			"old_title", t.Properties.TrackName,
			"new_title", key,
			"old_language", t.Properties.Language,
			"new_language", isoCode,
		)

		editArgs = append(editArgs,
			"--edit", fmt.Sprintf("track:@%d", t.ID+1),
			"--set", "name="+key,
			"--set", "language="+isoCode,
		)
	}

	if len(editArgs) == 0 {
		return nil
	}

	args := append([]string{entry.FileInfo.SourcePath}, editArgs...)
	out, err := exec.Command("mkvpropedit", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("remuxer: mkvpropedit failed for %q: %w\n%s", entry.FileInfo.SourcePath, err, out)
	}

	return nil
}

// resolveAudioLanguage returns the language Key for an audio track.
// It tries the ISO language tag first, then falls back to tokenising the title.
func resolveAudioLanguage(t mkvTrack) string {
	if t.Properties.Language != "" && t.Properties.Language != "und" {
		if key := extractor.ParseLanguage([]string{strings.ToUpper(t.Properties.Language)}); key != "" {
			return key
		}
	}

	// Split the title on non-alphanumeric characters and try each token.
	if t.Properties.TrackName != "" {
		tokens := strings.FieldsFunc(strings.ToUpper(t.Properties.TrackName), func(r rune) bool {
			return !('A' <= r && r <= 'Z') && !('0' <= r && r <= '9')
		})
		for _, tok := range tokens {
			if key := extractor.ParseLanguage([]string{tok}); key != "" {
				return key
			}
		}
	}

	return ""
}

func identifyAudioTracks(path string) ([]mkvTrack, error) {
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

	var audio []mkvTrack
	for _, t := range info.Tracks {
		if t.Type == "audio" {
			audio = append(audio, t)
		}
	}
	return audio, nil
}
