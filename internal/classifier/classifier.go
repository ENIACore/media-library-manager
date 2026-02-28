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

	lg.Error("Unable to classify entry", "entry", root.PathInfo.Source)
	return fmt.Errorf("Failed to classify root entry %v", root.PathInfo.Source)
}

// classifySubtitleDir recursively determines if entry is [metadata.SubtitleDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySubtitleDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySubtitleDir")
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	hasSubtitle := false
	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			hasSubtitle = true
		} else if classifySubtitleDir(child, logger)  { // Allows for intermediary directory, multiple interm=ediaries prevented with Height() limit
			hasSubtitle = true
		} else {
			return false
		}
	}

	entry.Role = metadata.SubtitleDir
	lg.Debug("Classification determined", "value", hasSubtitle) 
	return hasSubtitle
}

// Handles DS/BTS/Bonus directories and files
func classifyExtrasDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifyBonusDir")
	if !entry.PathInfo.IsDir || entry.Height() > 10 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	var childRole metadata.EntryRole

	// DS/BTS/Bonus directories HAVE to have the indicator in directory name OR
	// DS/BTS/Bonus directories can be a season directory WITH a parent DS/BTS/Bonus directory
	if entry.MediaInfo.DS != "" || (entry.MediaInfo.Season != nil && entry.Parent != nil && entry.Parent.MediaInfo.DS != "") {
		entry.Role = metadata.DSDir
		childRole = metadata.DSFile
	} else if entry.MediaInfo.BTS != "" || (entry.MediaInfo.Season != nil && entry.Parent != nil && entry.Parent.MediaInfo.BTS != "") {
		entry.Role = metadata.BTSDir
		childRole = metadata.BTSFile
	} else if entry.MediaInfo.Bonus != "" || (entry.MediaInfo.Season != nil && entry.Parent != nil && entry.Parent.MediaInfo.Bonus != "") {
		entry.Role = metadata.BonusDir
		childRole = metadata.BonusFile
	} else {
		lg.Debug("Directory and parent directory have no extra pattern", "source", entry.PathInfo.Source)
		return false
	}

	for _, child := range entry.Children {
		if classifySubtitleFile(child) {
			continue
		} else if classifySubtitleDir(child, logger) {
			continue
		}

		if child.PathInfo.IsDir {
			if classifyExtrasDir(child, logger) {
				continue
			} else {
				return false
			}
		} else {
			child.Role = childRole
		}
	}

	return true
}

// classifySeasonDir recursively determines if entry is [metadata.SeasonDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySeasonDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeasonDir")
	if !entry.PathInfo.IsDir || entry.Height() > 3 {
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

// classifySeriesDir recursively determines if entry is [metadata.SeriesDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
func classifySeriesDir(entry *metadata.Entry, logger *slog.Logger) bool {
	lg := logger.With("func", "classifySeriesDir")
	if !entry.PathInfo.IsDir || entry.Height() > 3 {
		lg.Debug("Invalid height or dir status", "IsDir", entry.PathInfo.IsDir, "height", entry.Height())
		return false
	}

	hasSeason := false
	for _, child := range entry.Children {
		if classifyExtrasDir(child, logger) {
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

// classifyMovieDir recursively determines if entry is [metadata.MovieDir].
// It assigns Role field for entry and all children; whether successful or not.
// Returns true if successful classification, false otherwise.
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
		if classifyDSFile(child) {
			continue
		} 
		if classifyBTSFile(child) {
			continue
		} 
		if classifyBonusFile(child) {
			continue
		} 
		if classifyExtrasDir(child, logger) {
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
		// If file has bonus information
		if entry.MediaInfo.DS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
		// If parent has bonus information
		if entry.Parent != nil && entry.Parent.MediaInfo.DS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
	}
	return false
}

func classifyBTSFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video {
		// If file has bonus information
		if entry.MediaInfo.BTS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
		// If parent has bonus information
		if entry.Parent != nil && entry.Parent.MediaInfo.BTS != "" {
			entry.Role = metadata.BonusFile
			return true
		}
	}
	return false

}

func classifyBonusFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video {
		// If file has bonus information
		if entry.MediaInfo.Bonus != "" {
			entry.Role = metadata.BonusFile
			return true
		}
		// If parent has bonus information
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
		// If file has season information
		if entry.MediaInfo.Episode != nil || entry.MediaInfo.Season != nil {
			entry.Role = metadata.EpisodeFile
			return true
		}
		// If parent has season information
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
