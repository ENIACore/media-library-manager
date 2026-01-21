package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
	"log/slog"
)

// Classifies in order of specificity
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
			lg.Debug("Unable to to classify child", child.PathInfo.Source)
			return false
		}
	}

	entry.Role = metadata.MovieDir
	lg.Debug("Classification determined", "value", hasMovie) 
	return hasMovie
}


func classifySubtitleFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Subtitle {
		entry.Role = metadata.SubtitleFile
		return true
	}
	return false
}

func classifyBonusFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Bonus != "" {
		entry.Role = metadata.BonusFile
		return true
	}
	return false
}

func classifyEpisodeFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Bonus == "" {
		if entry.MediaInfo.Episode != nil || entry.MediaInfo.Season != nil || (entry.Parent != nil && entry.Parent.MediaInfo.Season != nil) {
			entry.Role = metadata.EpisodeFile
			return true
		}
	}
	return false
}

func classifyMovieFile(entry *metadata.Entry) bool {
	if entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Episode == nil && entry.MediaInfo.Season == nil && entry.MediaInfo.Bonus == "" {
		entry.Role = metadata.MovieFile
		return true
	}
	return false
}
