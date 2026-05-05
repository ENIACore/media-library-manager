package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
	"log/slog"
)

// Classify calls recursive helper functions to determine classification for all [metadata.Entry] objects in tree.
// Classify (and its helper functions) use order of most specificity to ensure proper classification.
//	Dir order: subtitle > extras (ds, bts, bonus) > season > series > movie
//	File order: subtitle > extras (ds, bts, bonus) > episode > movie
func Classify(root *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "Classify")

	if root.FileInfo.IsDir {
		isRoot := root.Depth == 0
		if classifySubtitleDir(root, logger) || (!isRoot && classifyExtrasDir(root, metadata.UnknownRole, logger)) || classifySeasonDir(root, logger) || classifySeriesDir(root, logger) || classifyMovieDir(root, logger) {
			return nil
		}
	} else {
		if classifySubtitleFile(root) || classifyDSFile(root) || classifyBTSFile(root) || classifyBonusFile(root) || classifyEpisodeFile(root) || classifyMovieFile(root) {
			return nil
		}
	}

	lg.Error("Unable to classify entry", "entry", root.Source(), "height", root.Height())
	return fmt.Errorf("Failed to classify root entry %v", root.Source())
}

// classifySubtitleDir recursively determines if entry is [metadata.SubtitleDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySubtitleDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySubtitleDir")
	if !entry.FileInfo.IsDir || entry.Height() > 2 {
		return false
	}

	hasSubtitle := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) || classifySubtitleDir(child, logger) {
			hasSubtitle = true
		} else {
			return false
		}
	}

	entry.Role = metadata.SubtitleDir
	lg.Info("Classification determined", "source", entry.Source(), "role", entry.Role.String())
	return hasSubtitle
}

// Handles DS/BTS/Bonus directories and files
func classifyExtrasDir(entry *metadata.Entry, inheritedRole metadata.EntryRole, logger *slog.Logger) bool {
	lg := logger.With("func", "classifyExtrasDir")
	if !entry.FileInfo.IsDir || entry.Height() > 7 {
		return false
	}

	// DS/BTS/Bonus directories HAVE to have the indicator in directory name OR
	// DS/BTS/Bonus directories can be directories that inherit the bonus type of a parent/grandparent/etc..
	if entry.MediaInfo.DS != "" {
		entry.Role = metadata.DSDir
	} else if entry.MediaInfo.BTS != "" {
		entry.Role = metadata.BTSDir
	} else if entry.MediaInfo.Bonus != "" {
		entry.Role = metadata.BonusDir
	} else if inheritedRole != metadata.UnknownRole {
		entry.Role = inheritedRole
	} else {
		return false
	}

	var childRole metadata.EntryRole
	if entry.Role == metadata.DSDir {
		childRole = metadata.DSFile
	}
	if entry.Role == metadata.BTSDir {
		childRole = metadata.BTSFile
	}
	if entry.Role == metadata.BonusDir {
		childRole = metadata.BonusFile
	}

	for _, child := range entry.Children {
		if classifySubtitleFile(child) || classifySubtitleDir(child, logger) {
			continue
		} 
		if !child.FileInfo.IsDir {
			child.Role = childRole
		} else if classifyExtrasDir(child, entry.Role, logger) {
			continue
		} else {
			return false
		}
	}

	lg.Info("Classification determined", "source", entry.Source(), "role", entry.Role.String())
	return true
}

// classifySeasonDir recursively determines if entry is [metadata.SeasonDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySeasonDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeasonDir")
	if !entry.FileInfo.IsDir || entry.Height() > 7 {
		return false
	}

	hasEpisode := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) || classifySubtitleDir(child, logger) || classifyExtrasDir(child, metadata.UnknownRole, logger) {
			continue
		}
		if classifyEpisodeFile(child) {
			hasEpisode = true
		} else {
			return false
		}
	}

	entry.Role = metadata.SeasonDir
	lg.Info("Classification determined", "source", entry.Source(), "role", entry.Role.String())
	return hasEpisode
}

// classifySeriesDir recursively determines if entry is [metadata.SeriesDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySeriesDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeriesDir")
	if !entry.FileInfo.IsDir || entry.Height() > 7 {
		return false
	}

	hasSeason := false
	for _, child := range entry.Children {
		if classifyExtrasDir(child, metadata.UnknownRole, logger) || classifySubtitleDir(child, logger) {
			continue
		}
		if classifySeasonDir(child, logger) {
			hasSeason = true
		} else {
			return false
		}
	}

	entry.Role = metadata.SeriesDir
	lg.Info("Classification determined", "source", entry.Source(), "role", entry.Role.String())
	return hasSeason
}

// classifyMovieDir recursively determines if entry is [metadata.MovieDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifyMovieDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifyMovieDir")
	if !entry.FileInfo.IsDir || entry.Height() > 4 {
		return false
	}
	if entry.MediaInfo.Season != nil || entry.MediaInfo.Episode != nil {
		return false
	}

	hasMovie := false
	for _, child := range entry.Children {
		if 	classifySubtitleFile(child) || 
			classifySubtitleDir(child, logger) || 
			classifyDSFile(child) || 
			classifyBTSFile(child) || 
			classifyBonusFile(child) || 
			classifyExtrasDir(child, metadata.UnknownRole, logger) {
			continue
		}
		if classifyMovieFile(child) && !hasMovie {
			hasMovie = true
		} else {
			return false
		}
	}

	entry.Role = metadata.MovieDir
	lg.Info("Classification determined", "source", entry.Source(), "role", entry.Role.String())
	return hasMovie
}

// Reclassify overrides the role assigned by [Classify] using the authoritative media_type
// returned by TMDb. Call this after [Verify] has confirmed the title.
func Reclassify(root *metadata.Entry, tmdbMediaType string, logger *slog.Logger) error {
	lg := logger.With("func", "Reclassify", "source", root.Source())

	switch tmdbMediaType {
	case "movie":
		return reclassifyAsMovie(root, lg)
	case "tv":
		// TODO: reclassify movie-classified entries as TV (SeasonDir, SeriesDir, EpisodeFile, etc.)
		lg.Info("Classification confirmed as tv", "source", root.Source(), "role", root.Role.String())
		return nil
	default:
		return fmt.Errorf("classifier: unknown TMDb media type %q", tmdbMediaType)
	}
}

func reclassifyAsMovie(root *metadata.Entry, logger *slog.Logger) error {
	if !root.FileInfo.IsDir {
		root.Role = metadata.MovieFile
		logger.Info("Reclassified as movie file", "source", root.Source())
		return nil
	}

	root.Role = metadata.MovieDir
	logger.Info("Reclassified as movie dir", "source", root.Source())

	for _, child := range root.Children {
		if classifySubtitleFile(child) || classifySubtitleDir(child, logger) {
			continue
		}
		if classifyDSFile(child) || classifyBTSFile(child) || classifyBonusFile(child) {
			continue
		}
		if classifyExtrasDir(child, metadata.UnknownRole, logger) {
			continue
		}
		if child.FileInfo.ContentType == metadata.Video {
			child.Role = metadata.MovieFile
			logger.Info("Reclassified as movie file", "source", child.Source())
			continue
		}
		logger.Warn("Reclassify: unclassified child", "source", child.Source())
	}

	return nil
}

func classifySubtitleFile(entry *metadata.Entry) bool {
	if entry.FileInfo.ContentType == metadata.Subtitle {
		entry.Role = metadata.SubtitleFile
		return true
	}
	return false
}

func classifyDSFile(entry *metadata.Entry) bool {
	if entry.FileInfo.ContentType == metadata.Video {
		parentDS := entry.Parent != nil && entry.Parent.Depth > 0 && entry.Parent.MediaInfo.DS != ""
		if entry.MediaInfo.DS != "" || parentDS {
			entry.Role = metadata.DSFile
			return true
		}
	}
	return false
}

func classifyBTSFile(entry *metadata.Entry) bool {
	if entry.FileInfo.ContentType == metadata.Video {
		parentBTS := entry.Parent != nil && entry.Parent.Depth > 0 && entry.Parent.MediaInfo.BTS != ""
		if entry.MediaInfo.BTS != "" || parentBTS {
			entry.Role = metadata.BTSFile
			return true
		}
	}
	return false
}

func classifyBonusFile(entry *metadata.Entry) bool {
	if entry.FileInfo.ContentType == metadata.Video {
		parentBonus := entry.Parent != nil && entry.Parent.Depth > 0 && entry.Parent.MediaInfo.Bonus != ""
		if entry.MediaInfo.Bonus != "" || parentBonus {
			entry.Role = metadata.BonusFile
			return true
		}
	}
	return false
}

func classifyEpisodeFile(entry *metadata.Entry) bool {
	if entry.FileInfo.ContentType == metadata.Video {
		if (entry.MediaInfo.Episode != nil || entry.MediaInfo.Season != nil) || (entry.Parent != nil && entry.Parent.MediaInfo.Season != nil) {
			entry.Role = metadata.EpisodeFile
			return true
		}
	}
	return false
}

func classifyMovieFile(entry *metadata.Entry) bool {
	if entry.FileInfo.ContentType == metadata.Video && entry.MediaInfo.Episode == nil && entry.MediaInfo.Season == nil && entry.MediaInfo.Bonus == "" {
		entry.Role = metadata.MovieFile
		return true
	}
	return false
}
