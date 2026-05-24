// Package remuxer extracts subtitle tracks from MKV files using mkvmerge.
// Full subtitle tracks are written as <DestPath>.<language>.<ext>.
// Variant tracks (forced, SDH, signs/songs, commentary) are written as
// <DestPath>.<variant>.<language>.<ext> — e.g. "Show.Signs.Songs.English.ass"
// or "Show.SDH.English.srt" — so they coexist with full tracks of the same
// language without collision. The extension is derived from the source
// codec: srt, ass, or ssa. Runs immediately after parser.Parse, before
// classifier assigns EntryRole.
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
		CodecID             string `json:"codec_id"`
		Language            string `json:"language"`
		LanguageIETF        string `json:"language_ietf"`
		TrackName           string `json:"track_name"`
		ForcedTrack         bool   `json:"forced_track"`
		FlagHearingImpaired bool   `json:"flag_hearing_impaired"`
		FlagVisualImpaired  bool   `json:"flag_visual_impaired"`
		FlagOriginal        bool   `json:"flag_original"`
		FlagCommentary      bool   `json:"flag_commentary"`
		DefaultTrack        bool   `json:"default_track"`
	} `json:"properties"`
}

type mkvFile struct {
	Tracks []mkvTrack `json:"tracks"`
}

// Subtitle variant keys, used as filename prefixes before the language token.
// Empty string means "full" (no prefix). Order matters in classifySubtitleVariant —
// more specific variants are checked before more general ones (Signs.Songs before
// Forced, since signs/songs tracks are typically also flagged forced).
const (
	variantFull        = ""
	variantSignsSongs  = "Signs.Songs"
	variantForced      = "Forced"
	variantSDH         = "SDH"
	variantCommentary  = "Commentary"
	variantDescriptive = "Descriptive"
)

// Subtitle file extensions we emit.
const (
	extSRT = "srt"
	extASS = "ass"
	extSSA = "ssa"
)

// Title substring markers, checked case-insensitively.
var (
	signsSongsMarkers  = []string{"sign", "song", "sing"}
	sdhMarkers         = []string{"sdh", "cc", "closed caption", "hearing impaired", "hearing-impaired"}
	forcedMarkers      = []string{"forced"}
	commentaryMarkers  = []string{"commentary"}
	descriptiveMarkers = []string{"description", "descriptive", "audio description"}
)

// Remux examines the entry: if it is not an MKV file it returns immediately.
// Extracts text subtitles when possible (full + variants), then ALWAYS strips
// all subtitle tracks.
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
		track   mkvTrack
		lang    string
		variant string
		ext     string
		dest    string
	}

	base := strings.TrimSuffix(entry.FileInfo.DestPath, filepath.Ext(entry.FileInfo.DestPath))
	seen := make(map[string]bool) // key = variant + "|" + lang
	var candidates []candidateTrack

	for _, t := range subtitles {
		ext, ok := subtitleExtension(t)
		if !ok {
			lg.Info("skipping subtitle track with unsupported codec",
				"id", t.ID,
				"codec", t.Codec,
				"codec_id", t.Properties.CodecID,
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

		variant := classifySubtitleVariant(t)
		dedupKey := variant + "|" + lang

		if seen[dedupKey] {
			lg.Info("skipping duplicate subtitle track",
				"id", t.ID,
				"language", lang,
				"variant", variantLabel(variant),
				"title", t.Properties.TrackName,
			)
			continue
		}
		seen[dedupKey] = true

		candidates = append(candidates, candidateTrack{
			track:   t,
			lang:    lang,
			variant: variant,
			ext:     ext,
			dest:    buildDestPath(base, lang, variant, ext),
		})
	}

	if len(candidates) > 0 {
		for _, c := range candidates {
			if err := extractTrack(entry.FileInfo.SourcePath, c.track.ID, c.dest, lg); err != nil {
				return err
			}
			lg.Info("extracted subtitle track",
				"id", c.track.ID,
				"language", c.lang,
				"variant", variantLabel(c.variant),
				"format", c.ext,
				"title", c.track.Properties.TrackName,
				"dest", filepath.Base(c.dest),
			)
		}
	} else {
		lg.Info("no extractable subtitle tracks after filtering, proceeding to strip all subtitles")
	}

	if err := stripSubtitles(entry.FileInfo.SourcePath, lg); err != nil {
		return err
	}

	lg.Info("stripped subtitle tracks from MKV",
		"path", filepath.Base(entry.FileInfo.SourcePath),
	)

	return nil
}

// classifySubtitleVariant returns the variant key for a subtitle track, or
// variantFull ("") for a full dialogue track. Disposition flags are checked
// alongside title-based heuristics; titles take precedence for the more
// specific variants because the forced flag alone can't distinguish
// Signs/Songs from generic forced subs. Order is significant: Signs/Songs
// is checked before Forced because signs/songs tracks usually also have
// the forced flag set.
func classifySubtitleVariant(t mkvTrack) string {
	p := t.Properties
	title := strings.ToLower(p.TrackName)

	if titleContainsAny(title, signsSongsMarkers) {
		return variantSignsSongs
	}
	if titleContainsAny(title, sdhMarkers) || p.FlagHearingImpaired {
		return variantSDH
	}
	if titleContainsAny(title, commentaryMarkers) || p.FlagCommentary {
		return variantCommentary
	}
	if titleContainsAny(title, descriptiveMarkers) || p.FlagVisualImpaired {
		return variantDescriptive
	}
	if titleContainsAny(title, forcedMarkers) || p.ForcedTrack {
		return variantForced
	}

	return variantFull
}

func titleContainsAny(title string, markers []string) bool {
	if title == "" {
		return false
	}
	for _, m := range markers {
		if strings.Contains(title, m) {
			return true
		}
	}
	return false
}

// variantLabel returns a human-friendly label for logging.
func variantLabel(v string) string {
	if v == variantFull {
		return "Full"
	}
	return v
}

// buildDestPath constructs the output subtitle path, all dot-separated.
// Full tracks: "<base>.<lang>.<ext>"
// Variants:   "<base>.<variant>.<lang>.<ext>"
func buildDestPath(base, lang, variant, ext string) string {
	if variant == variantFull {
		return base + "." + lang + "." + ext
	}
	return base + "." + variant + "." + lang + "." + ext
}

// subtitleExtension returns the file extension to use for a given subtitle
// track, derived from its codec. Returns false for codecs we can't extract
// as text (PGS, VobSub, etc.).
//
// mkvmerge's `codec` field is a human-readable label that varies by version,
// while `codec_id` is the stable Matroska identifier. We check codec_id first
// for reliability and fall back to codec for older mkvmerge output.
func subtitleExtension(t mkvTrack) (string, bool) {
	switch strings.ToUpper(t.Properties.CodecID) {
	case "S_TEXT/UTF8", "S_TEXT/ASCII":
		return extSRT, true
	case "S_TEXT/ASS":
		return extASS, true
	case "S_TEXT/SSA":
		return extSSA, true
	}

	switch t.Codec {
	case "SubRip/SRT", "S_TEXT/UTF8":
		return extSRT, true
	case "SubStationAlpha", "S_TEXT/ASS", "ASS":
		return extASS, true
	case "S_TEXT/SSA", "SSA":
		return extSSA, true
	}

	return "", false
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

func extractTrack(srcPath string, trackID int, destPath string, lg *slog.Logger) error {
	args := []string{
		"tracks",
		srcPath,
		fmt.Sprintf("%d:%s", trackID, destPath),
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
