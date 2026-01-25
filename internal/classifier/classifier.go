package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
	"log/slog"
)

// Classify determines the role of a media entry in the library hierarchy.
// Attempts classification in order of specificity, first checking if entry
// is a directory or file, then trying type-specific classifiers.
// Returns an error if classification fails for the root entry.
func Classify(root *metadata.Entry, logger *slog.Logger) error {
	lg := logger.With("func", "Classify")

	if root.PathInfo.IsDir {
		if classifySubtitleDir(root, logger) || classifyBonusDir(root, logger) || classifySeasonDir(root, logger) || classifySeriesDir(root, logger) || classifyMovieDir(root, logger) {
			lg.Info("Root entry classification determined", "entry", root.PathInfo.Source, "role", root.Role)
			return nil
		}
	} else {
		if classifySubtitleFile(root) || classifyBonusFile(root) || classifyEpisodeFile(root) || classifyMovieFile(root) {
			lg.Info("Root entry classification determined", "entry", root.PathInfo.Source, "role", root.Role)
			return nil
		} 
	}
	
	lg.Error("Unable to classify entry", "entry", root.PathInfo.Source)
	return fmt.Errorf("Failed to classify root entry %v", root.PathInfo.Source)
}

// classifySubtitleDir classifies a directory as SubtitleDir if it contains only subtitle files.
// Only directories with height 1 or less (leaf or near-leaf) are considered.
// Returns true and sets entry.Role to SubtitleDir if classification succeeds.
func classifySubtitleDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySubtitleDir")
	if !entry.PathInfo.IsDir || entry.Height() > 1 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	hasSubtitle := false
	for _, child := range entry.Children {
		if !classifySubtitleFile(child) {
			return false

		}
		hasSubtitle = true
	}

	entry.Role = metadata.SubtitleDir
	lg.Debug("Classification determined", "value", hasSubtitle) 
	return hasSubtitle
}

// classifyBonusDir classifies a directory as BonusDir if it contains bonus content files.
// Only directories with height 1 or less are considered. Subtitle files are allowed.
// Returns true and sets entry.Role to BonusDir if classification succeeds.
func classifyBonusDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifyBonusDir")
	if !entry.PathInfo.IsDir || entry.Height() > 1 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	hasBonus := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			continue
		}

		if classifyBonusFile(child) {
			hasBonus = true
		}  else {
			return false
		}
	}

	entry.Role = metadata.BonusDir	
	lg.Debug("Classification determined", "value", hasBonus) 
	return hasBonus
}

// classifySeasonDir classifies a directory as SeasonDir if it contains episode files.
// Only directories with height 2 or less are considered. Subtitle files and directories are allowed.
// Returns true and sets entry.Role to SeasonDir if classification succeeds.
func classifySeasonDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeasonDir")
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	hasEpisode := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			continue
		}
		if classifySubtitleDir(child, logger) {
			continue
		} 

		if classifyEpisodeFile(child) {
			hasEpisode = true
		} else {
			return false
		}
	}


	entry.Role = metadata.SeasonDir	
	lg.Debug("Classification determined", "value", hasEpisode) 
	return hasEpisode
}

// classifySeriesDir classifies a directory as SeriesDir if it contains season directories.
// Only directories with height 3 or less are considered. Bonus and subtitle directories are allowed.
// Returns true and sets entry.Role to SeriesDir if classification succeeds.
func classifySeriesDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeriesDir")
	if !entry.PathInfo.IsDir || entry.Height() > 3 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	hasSeason := false
	for _, child := range entry.Children {
		if classifyBonusDir(child, logger) {
			continue
		}
		if classifySubtitleDir(child, logger) {
			continue
		}


		if classifySeasonDir(child, logger) {
			hasSeason = true
		}  else {
			return false
		}
	}

	entry.Role = metadata.SeriesDir
	lg.Debug("Classification determined", "value", hasSeason) 
	return hasSeason
}

// classifyMovieDir classifies a directory as MovieDir if it contains exactly one movie file.
// Only directories with height 2 or less are considered. Entry must not have season or episode metadata.
// Subtitle and bonus content (files and directories) are allowed.
// Returns true and sets entry.Role to MovieDir if classification succeeds.
func classifyMovieDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifyMovieDir")
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}
	if entry.MediaInfo.Season != nil || entry.MediaInfo.Episode != nil {
		lg.Debug("Unexpected season or ep present", "season", entry.MediaInfo.Season, "episode", entry.MediaInfo.Episode)
		return false
	}

	hasMovie := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			continue
		} 
		if classifySubtitleDir(child, logger) {
			continue
		} 
		if classifyBonusFile(child) {
			continue
		} 
		if classifyBonusDir(child, logger) {
			continue
		}



		if classifyMovieFile(child) && !hasMovie{
			hasMovie = true
		} else {
			lg.Debug("Unable to to classify child", "entry", child.PathInfo.Source)
			return false
		}
	}

	entry.Role = metadata.MovieDir
	lg.Debug("Classification determined", "value", hasMovie) 
	return hasMovie
}


// classifySubtitleFile classifies a file as SubtitleFile based on its content type.
// Returns true and sets entry.Role to SubtitleFile if the file type is Subtitle.
func classifySubtitleFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Subtitle {
		entry.Role = metadata.SubtitleFile
		return true
	}
	return false
}

// classifyBonusFile classifies a file as BonusFile if it contains bonus content metadata.
// Returns true and sets entry.Role to BonusFile if the file is a video with bonus metadata.
func classifyBonusFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Bonus != "" {
		entry.Role = metadata.BonusFile
		return true
	}
	return false
}

// classifyEpisodeFile classifies a file as EpisodeFile based on episode or season metadata.
// Returns true and sets entry.Role to EpisodeFile if the file is a video without bonus metadata
// and has episode/season metadata in the file or parent directory.
func classifyEpisodeFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Bonus == "" {
		if entry.MediaInfo.Episode != nil || entry.MediaInfo.Season != nil || (entry.Parent != nil && entry.Parent.MediaInfo.Season != nil) {
			entry.Role = metadata.EpisodeFile
			return true
		}
	}
	return false
}

// classifyMovieFile classifies a file as MovieFile if it lacks series metadata.
// Returns true and sets entry.Role to MovieFile if the file is a video without
// episode, season, or bonus metadata.
func classifyMovieFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Episode == nil && entry.MediaInfo.Season == nil && entry.MediaInfo.Bonus == "" {
		entry.Role = metadata.MovieFile
		return true
	}
	return false
}
