package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
	"log/slog"
)

func ClassifyEntries(root *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "ClassifyEntries")
	log.Info("classifying root", "path", root.PathInfo.Source)

	if root.PathInfo.IsDir {
		var err error

		switch role := classifyDir(root); role {
			case metadata.SubtitleDir:
				err = classifySubtitleDir(root, logger)
			case metadata.BonusDir:
				err = classifyBonusDir(root, logger)
			case metadata.SeasonDir:
				err = classifySeasonDir(root, logger)
			case metadata.SeriesDir:
				err = classifySeriesDir(root, logger)
			case metadata.MovieDir:
				err = classifyMovieDir(root, logger)
			default:
				log.Error("unclassifiable dir", "path", root.PathInfo.Source)
				return fmt.Errorf("root entry %v is a dir that could not be classified", root.PathInfo.Source)
		}

		if err != nil {
			log.Error("classification failed", "path", root.PathInfo.Source, "err", err)
			return fmt.Errorf("failed to classify root %v: %w", root.PathInfo.Source, err)
		}

	} else {
		role := classifyFile(root)

		if role == metadata.UnknownRole {
			log.Error("unclassifiable file", "path", root.PathInfo.Source)
			return fmt.Errorf("root entry %v is a dir that could not be classified", root.PathInfo.Source)
		} else {
			root.Role = role
		}
	}
	
	log.Info("classified root", "path", root.PathInfo.Source, "role", root.Role)
	return nil
}

/*
	classifier helper functions
*/


func classifySubtitleDir(entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "classifySubtitleDir")

	for _, child := range entry.Children {
		if isSubtitleFile(child) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.SubtitleFile)
			child.Role = metadata.SubtitleFile
		} else {
			log.Debug("unknown child", "path", child.PathInfo.Source)
			entry.Role = metadata.UnknownRole
			return fmt.Errorf("entry %v could not be classified as a subtitle dir child", child.PathInfo.Source)
		}
	}


	log.Debug("classified dir", "path", entry.PathInfo.Source, "role", metadata.SubtitleDir)
	entry.Role = metadata.SubtitleDir
	return nil
}

func classifyBonusDir(entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "classifyBonusDir")

	bonus := false
	for _, child := range entry.Children {
		if isBonusFile(child) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.BonusFile)
			bonus = true
			child.Role = metadata.BonusFile
		} else if isSubtitleFile(child) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.SubtitleFile)
			child.Role = metadata.SubtitleFile
		} else {
			log.Debug("unknown child", "path", child.PathInfo.Source)
			return fmt.Errorf("entry %v could not be classified as a bonus dir child", child.PathInfo.Source)
		}
	}

	if !bonus {
		log.Debug("missing required child", "path", entry.PathInfo.Source)
		return fmt.Errorf("no required children of bonus dir found")
	}

	log.Debug("classified dir", "path", entry.PathInfo.Source, "role", metadata.BonusDir)
	entry.Role = metadata.BonusDir	
	return nil
}

func classifySeasonDir(entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "classifySeasonDir")

	episode := false
	for _, child := range entry.Children {
		if isSubtitleFile(child) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.SubtitleFile)
			child.Role = metadata.SubtitleFile
		} else if isSubtitleDir(child) {
			classifySubtitleDir(child, logger)
		} else if child.PathInfo.Type == metadata.Video && (isEpisodeFile(child) || entry.MediaInfo.Season != nil) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.EpisodeFile)
			child.Role = metadata.EpisodeFile
			episode = true
		} else {
			log.Debug("unknown child", "path", child.PathInfo.Source)
			return fmt.Errorf("entry %v could not be classified as a season dir child", child.PathInfo.Source)
		}
	}

	if !episode {
		log.Debug("missing required child", "path", entry.PathInfo.Source)
		return fmt.Errorf("no required children of season dir found")
	}

	log.Debug("classified dir", "path", entry.PathInfo.Source, "role", metadata.SeasonDir)
	entry.Role = metadata.SeasonDir	
	return nil
}

func classifySeriesDir(entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "classifySeriesDir")

	season := false
	for _, child := range entry.Children {
		if isSeasonDir(child) {
			season = true
			classifySeasonDir(child, logger)
		} else if isBonusDir(child) {
			classifyBonusDir(child, logger)
		} else if isSubtitleDir(child) {
			classifySubtitleDir(child, logger)
		} else {
			log.Debug("unknown child", "path", child.PathInfo.Source)
			return fmt.Errorf("entry %v could not be classified as a series dir child", child.PathInfo.Source)
		}
	}

	if !season {
		log.Debug("missing required child", "path", entry.PathInfo.Source)
		return fmt.Errorf("no required children of series dir found")
	}

	log.Debug("classified dir", "path", entry.PathInfo.Source, "role", metadata.SeriesDir)
	entry.Role = metadata.SeriesDir
	return nil
}

func classifyMovieDir(entry *metadata.Entry, logger *slog.Logger) error {
	log := logger.With("func", "classifyMovieDir")

	movie := false
	for _, child := range entry.Children {
		if isSubtitleFile(child) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.SubtitleFile)
			child.Role = metadata.SubtitleFile
		} else if isSubtitleDir(child) {
			classifySubtitleDir(child, logger)
		} else if isBonusFile(child) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.BonusFile)
			child.Role = metadata.BonusFile
		} else if isBonusDir(child) {
			classifyBonusDir(child, logger)
		} else if isMovieFile(child) {
			log.Debug("classified child", "path", child.PathInfo.Source, "role", metadata.MovieFile)
			movie = true
			child.Role = metadata.MovieFile
		} else {
			log.Debug("unknown child", "path", child.PathInfo.Source)
			return fmt.Errorf("entry %v could not be classified as a movie dir child", child.PathInfo.Source)
		}
	}

	if !movie {
		log.Debug("missing required child", "path", entry.PathInfo.Source)
		return fmt.Errorf("no required children of movie dir found")
	}

	log.Debug("classified dir", "path", entry.PathInfo.Source, "role", metadata.MovieDir)
	entry.Role = metadata.MovieDir
	return nil
}

func classifyFile(entry *metadata.Entry) metadata.EntryRole {

	// 1. Completely distinct type
	if isSubtitleFile(entry) {
		return metadata.SubtitleFile
	}

	// 2. Special-case video types
	if isBonusFile(entry) {
		return metadata.BonusFile
	}

	// 3. Episodic content (more specific than movie)
	if isEpisodeFile(entry) {
		return metadata.EpisodeFile
	}

	// 4. Default video type
	if isMovieFile(entry) {
		return metadata.MovieFile
	}

	// 5. Fallback
	return metadata.UnknownRole
}

func classifyDir(entry *metadata.Entry) metadata.EntryRole {
	// 1. Completely distinct type
	if isSubtitleDir(entry) {
		return metadata.SubtitleDir
	}

	// 2. Special-case video types
	if isBonusDir(entry) {
		return metadata.BonusDir
	}

	// 3. Episodic content (more specific than movie)
	if isSeasonDir(entry) {
		return metadata.SeasonDir
	}

	// 4. Episodic content wrapped by directory (more specific than season)
	if isSeriesDir(entry) {
		return metadata.SeriesDir
	}

	// 5. Default directory type
	if isMovieDir(entry) {
		return metadata.MovieDir
	}

	// 6. Fallback
	return metadata.UnknownRole
}


/*
	file helpers
*/


// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isSubtitleFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Subtitle
}

// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isBonusFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Bonus != ""
}

// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isEpisodeFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Video && (entry.MediaInfo.Episode != nil || entry.MediaInfo.Season != nil) && entry.MediaInfo.Bonus == ""
}

// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isMovieFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Episode == nil && entry.MediaInfo.Season == nil && entry.MediaInfo.Bonus == ""
}


/*
	dir helpers
*/

func isSubtitleDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 1 {
		return false
	}

	subtitle := false
	for _, child := range entry.Children {
		if isSubtitleFile(child) {
			subtitle = true
		} else {
			return false
		}
	}

	return subtitle
}

func isBonusDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 1 {
		return false
	}

	bonus := false
	for _, child := range entry.Children {
		if isSubtitleFile(child) {
			continue
		}

		if child.PathInfo.Type == metadata.Video && (isBonusFile(child) || entry.MediaInfo.Bonus != "") {
			bonus = true
		} else {
			return false
		}
	}

	return bonus
}

func isSeasonDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		return false
	}

	episode := false
	for _, child := range entry.Children {
		if isSubtitleFile(child) || isSubtitleDir(child) {
			continue
		}

		if child.PathInfo.Type == metadata.Video && (isEpisodeFile(child) || entry.MediaInfo.Season != nil) {
			episode = true
		} else {
			return false
		}
	}

	return entry.MediaInfo.Season != nil || episode
}

func isSeriesDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 3 {
		return false
	}
	
	season := false
	for _, child := range entry.Children {
		if isSubtitleDir(child) || isBonusDir(child) {
			continue
		}

		if isSeasonDir(child) {
			season = true
		} else {
			return false
		}
	}

	return season
}

func isMovieDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		return false
	}
	if entry.MediaInfo.Season != nil || entry.MediaInfo.Episode != nil {
		return false
	}

	movie := false
	for _, child := range entry.Children {
		if isSubtitleFile(child) || isBonusFile(child) {
			continue
		}
		if child.PathInfo.IsDir && (isBonusDir(child) || isSubtitleDir(child)) {
			continue
		}
		if isMovieFile(child) && !movie {
			movie = true
		} else {
			return false
		}
	}
	return movie
}
