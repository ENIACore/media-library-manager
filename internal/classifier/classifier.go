package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
	"log/slog"
)

// Classify calls recursive helper functions to determine classification for all [metadata.Entry] objects in tree.
// Classify (and its helper functions) use order of most specificity to ensure proper classification.
//	Dir order: subtitle > bonus > season > series > movie
//	File order: subtitle > bonus > episode > movie
func Classify(root *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "Classify")

	lg.Debug("Attempting to classify root entry", "entry", root.PathInfo.Source, "isDir", root.PathInfo.IsDir, "height", root.Height())

	if root.PathInfo.IsDir {
		if classifySubtitleDir(root, logger) || classifyExtrasDir(root, logger) || classifySeasonDir(root, logger) || classifySeriesDir(root, logger) || classifyMovieDir(root, logger) {
			lg.Info("Root entry classification determined", "entry", root.PathInfo.Source, "role", root.Role)
			return nil
		}
	} else {
		if classifySubtitleFile(root) || classifyDSFile(root) || classifyBTSFile(root) || classifyBonusFile(root) || classifyEpisodeFile(root) || classifyMovieFile(root) {
			lg.Info("Root entry classification determined", "entry", root.PathInfo.Source, "role", root.Role)
			return nil
		}
	}

	lg.Error("Unable to classify entry", "entry", root.PathInfo.Source, "height", root.Height(), "season", root.MediaInfo.Season, "episode", root.MediaInfo.Episode, "bonus", root.MediaInfo.Bonus, "ds", root.MediaInfo.DS, "bts", root.MediaInfo.BTS)
	return fmt.Errorf("Failed to classify root entry %v", root.PathInfo.Source)
}

// classifySubtitleDir recursively determines if entry is [metadata.SubtitleDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySubtitleDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySubtitleDir")
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		lg.Debug("Skipping: invalid height or not a dir", "source", entry.PathInfo.Source, "isDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	lg.Debug("Checking entry", "source", entry.PathInfo.Source, "childCount", len(entry.Children))

	hasSubtitle := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			lg.Debug("Child classified as subtitle file", "child", child.PathInfo.Source)
			hasSubtitle = true
		} else if classifySubtitleDir(child, logger) {
			lg.Debug("Child classified as subtitle dir", "child", child.PathInfo.Source)
			hasSubtitle = true
		} else {
			lg.Debug("Child failed subtitle classification, rejecting dir", "child", child.PathInfo.Source, "type", child.PathInfo.Type)
			return false
		}
	}

	entry.Role = metadata.SubtitleDir
	lg.Info("Classification determined", "source", entry.PathInfo.Source, "role", entry.Role)
	return hasSubtitle
}

// Handles DS/BTS/Bonus directories and files
func classifyExtrasDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifyExtrasDir")
	if !entry.PathInfo.IsDir || entry.Height() > 7 {
		lg.Debug("Skipping: invalid height or not a dir", "source", entry.PathInfo.Source, "isDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	lg.Debug("Checking entry", "source", entry.PathInfo.Source, "ds", entry.MediaInfo.DS, "bts", entry.MediaInfo.BTS, "bonus", entry.MediaInfo.Bonus, "season", entry.MediaInfo.Season)

	var childRole metadata.EntryRole

	// DS/BTS/Bonus directories HAVE to have the indicator in directory name OR
	// DS/BTS/Bonus directories can be a season directory WITH a parent DS/BTS/Bonus directory
	if entry.MediaInfo.DS != "" || (entry.MediaInfo.Season != nil && entry.Parent != nil && entry.Parent.MediaInfo.DS != "") {
		lg.Debug("Matched as DS dir", "source", entry.PathInfo.Source)
		entry.Role = metadata.DSDir
		childRole = metadata.DSFile
	} else if entry.MediaInfo.BTS != "" || (entry.MediaInfo.Season != nil && entry.Parent != nil && entry.Parent.MediaInfo.BTS != "") {
		lg.Debug("Matched as BTS dir", "source", entry.PathInfo.Source)
		entry.Role = metadata.BTSDir
		childRole = metadata.BTSFile
	} else if entry.MediaInfo.Bonus != "" || (entry.MediaInfo.Season != nil && entry.Parent != nil && entry.Parent.MediaInfo.Bonus != "") {
		lg.Debug("Matched as Bonus dir", "source", entry.PathInfo.Source)
		entry.Role = metadata.BonusDir
		childRole = metadata.BonusFile
	} else {
		lg.Debug("No extras pattern matched, rejecting", "source", entry.PathInfo.Source)
		return false
	}

	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			lg.Debug("Child classified as subtitle file", "child", child.PathInfo.Source)
			continue
		} else if classifySubtitleDir(child, logger) {
			lg.Debug("Child classified as subtitle dir", "child", child.PathInfo.Source)
			continue
		}

		if child.PathInfo.IsDir {
			if classifyExtrasDir(child, logger) {
				lg.Debug("Child classified as nested extras dir", "child", child.PathInfo.Source, "role", child.Role)
				continue
			} else {
				lg.Debug("Child dir failed extras classification, rejecting parent", "child", child.PathInfo.Source)
				return false
			}
		} else {
			lg.Debug("Assigning child role", "child", child.PathInfo.Source, "role", childRole)
			child.Role = childRole
		}
	}

	lg.Info("Classification determined", "source", entry.PathInfo.Source, "role", entry.Role)
	return true
}

// classifySeasonDir recursively determines if entry is [metadata.SeasonDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySeasonDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeasonDir")
	if !entry.PathInfo.IsDir || entry.Height() > 7 {
		lg.Debug("Skipping: invalid height or not a dir", "source", entry.PathInfo.Source, "isDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	lg.Debug("Checking entry", "source", entry.PathInfo.Source, "season", entry.MediaInfo.Season, "childCount", len(entry.Children))

	hasEpisode := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			lg.Debug("Child classified as subtitle file", "child", child.PathInfo.Source)
			continue
		}
		if classifySubtitleDir(child, logger) {
			lg.Debug("Child classified as subtitle dir", "child", child.PathInfo.Source)
			continue
		}
		if classifyExtrasDir(child, logger) {
			lg.Debug("Child classified as extras dir", "child", child.PathInfo.Source, "role", child.Role)
			continue
		}

		if classifyEpisodeFile(child) {
			lg.Debug("Child classified as episode file", "child", child.PathInfo.Source, "season", child.MediaInfo.Season, "episode", child.MediaInfo.Episode)
			hasEpisode = true
		} else {
			lg.Debug("Child failed all classifications, rejecting season dir", "child", child.PathInfo.Source, "isDir", child.PathInfo.IsDir, "type", child.PathInfo.Type, "season", child.MediaInfo.Season, "episode", child.MediaInfo.Episode, "bonus", child.MediaInfo.Bonus)
			return false
		}
	}

	entry.Role = metadata.SeasonDir
	lg.Info("Classification determined", "source", entry.PathInfo.Source, "role", entry.Role)
	return hasEpisode
}

// classifySeriesDir recursively determines if entry is [metadata.SeriesDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySeriesDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeriesDir")
	if !entry.PathInfo.IsDir || entry.Height() > 7 {
		lg.Debug("Skipping: invalid height or not a dir", "source", entry.PathInfo.Source, "isDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	lg.Debug("Checking entry", "source", entry.PathInfo.Source, "season", entry.MediaInfo.Season, "childCount", len(entry.Children))

	hasSeason := false
	for _, child := range entry.Children {
		if classifyExtrasDir(child, logger) {
			lg.Debug("Child classified as extras dir", "child", child.PathInfo.Source, "role", child.Role)
			continue
		}
		if classifySubtitleDir(child, logger) {
			lg.Debug("Child classified as subtitle dir", "child", child.PathInfo.Source)
			continue
		}

		if classifySeasonDir(child, logger) {
			lg.Debug("Child classified as season dir", "child", child.PathInfo.Source)
			hasSeason = true
		} else {
			lg.Debug("Child failed all classifications, rejecting series dir", "child", child.PathInfo.Source, "isDir", child.PathInfo.IsDir, "type", child.PathInfo.Type, "season", child.MediaInfo.Season, "episode", child.MediaInfo.Episode, "bonus", child.MediaInfo.Bonus, "height", child.Height())
			return false
		}
	}

	entry.Role = metadata.SeriesDir
	lg.Info("Classification determined", "source", entry.PathInfo.Source, "role", entry.Role)
	return hasSeason
}

// classifyMovieDir recursively determines if entry is [metadata.MovieDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifyMovieDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifyMovieDir")
	if !entry.PathInfo.IsDir || entry.Height() > 4 {
		lg.Debug("Skipping: invalid height or not a dir", "source", entry.PathInfo.Source, "isDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}
	if entry.MediaInfo.Season != nil || entry.MediaInfo.Episode != nil {
		lg.Debug("Skipping: unexpected season or episode metadata present", "source", entry.PathInfo.Source, "season", entry.MediaInfo.Season, "episode", entry.MediaInfo.Episode)
		return false
	}

	lg.Debug("Checking entry", "source", entry.PathInfo.Source, "childCount", len(entry.Children))

	hasMovie := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			lg.Debug("Child classified as subtitle file", "child", child.PathInfo.Source)
			continue
		}
		if classifySubtitleDir(child, logger) {
			lg.Debug("Child classified as subtitle dir", "child", child.PathInfo.Source)
			continue
		}
		if classifyDSFile(child) {
			lg.Debug("Child classified as DS file", "child", child.PathInfo.Source)
			continue
		}
		if classifyBTSFile(child) {
			lg.Debug("Child classified as BTS file", "child", child.PathInfo.Source)
			continue
		}
		if classifyBonusFile(child) {
			lg.Debug("Child classified as bonus file", "child", child.PathInfo.Source)
			continue
		}
		if classifyExtrasDir(child, logger) {
			lg.Debug("Child classified as extras dir", "child", child.PathInfo.Source, "role", child.Role)
			continue
		}

		if classifyMovieFile(child) && !hasMovie {
			lg.Debug("Child classified as movie file", "child", child.PathInfo.Source)
			hasMovie = true
		} else {
			lg.Debug("Child failed all classifications, rejecting movie dir", "child", child.PathInfo.Source, "isDir", child.PathInfo.IsDir, "type", child.PathInfo.Type, "season", child.MediaInfo.Season, "episode", child.MediaInfo.Episode, "bonus", child.MediaInfo.Bonus, "alreadyHasMovie", hasMovie)
			return false
		}
	}

	entry.Role = metadata.MovieDir
	lg.Info("Classification determined", "source", entry.PathInfo.Source, "role", entry.Role)
	return hasMovie
}

// classifySubtitleFile determines if entry is [metadata.SubtitleFile].
// Assigns role if successful.
func classifySubtitleFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Subtitle {
		entry.Role = metadata.SubtitleFile
		return true
	}
	return false
}

func classifyDSFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video {
		if entry.MediaInfo.DS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
		if entry.Parent != nil && entry.Parent.MediaInfo.DS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
	}
	return false
}

func classifyBTSFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video {
		if entry.MediaInfo.BTS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
		if entry.Parent != nil && entry.Parent.MediaInfo.BTS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
	}
	return false
}

func classifyBonusFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video {
		if entry.MediaInfo.Bonus != "" {
			entry.Role = metadata.BonusFile
			return true
		}
		if entry.Parent != nil && entry.Parent.MediaInfo.Bonus != "" {
			entry.Role = metadata.BonusFile
			return true
		}
	}
	return false
}

// classifyEpisodeFile determines if entry is [metadata.EpisodeFile].
// Assigns role if successful.
func classifyEpisodeFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Bonus == "" {
		if entry.MediaInfo.Episode != nil || entry.MediaInfo.Season != nil {
			entry.Role = metadata.EpisodeFile
			return true
		}
		if entry.Parent != nil && entry.Parent.MediaInfo.Season != nil {
			entry.Role = metadata.EpisodeFile
			return true
		}
	}
	return false
}

// classifyMovieFile determines if entry is [metadata.MovieFile].
// Assigns role if successful.
func classifyMovieFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Episode == nil && entry.MediaInfo.Season == nil && entry.MediaInfo.Bonus == "" {
		entry.Role = metadata.MovieFile
		return true
	}
	return false
}
