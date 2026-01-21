package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
	"log/slog"
)

// Classifies in order of specificity
func Classify(root *metadata.Entry, logger *slog.Logger) error {

	if root.PathInfo.IsDir {
		if classifySubtitleDir(root, logger) || classifyBonusDir(root, logger) || classifySeasonDir(root, logger) || classifySeriesDir(root, logger) || classifyMovieDir(root, logger) {
			return nil
		}
	} else {
		if classifySubtitleFile(root) || classifyBonusFile(root) || classifyEpisodeFile(root) || classifyMovieFile(root) {
			return nil
		} 
	}
	return fmt.Errorf("Failed to classify root entry %v", root.PathInfo.Source)
}

func classifySubtitleDir(entry *metadata.Entry, logger *slog.Logger) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 1 {
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
	return hasSubtitle
}

func classifyBonusDir(entry *metadata.Entry, logger *slog.Logger) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 1 {
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
	return hasBonus
}

func classifySeasonDir(entry *metadata.Entry, logger *slog.Logger) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
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
	return hasEpisode
}

func classifySeriesDir(entry *metadata.Entry, logger *slog.Logger) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 3 {
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
	return hasSeason
}

func classifyMovieDir(entry *metadata.Entry, logger *slog.Logger) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		return false
	}
	if entry.MediaInfo.Season != nil || entry.MediaInfo.Episode != nil {
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
			return false
		}
	}

	entry.Role = metadata.MovieDir
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
