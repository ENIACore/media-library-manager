package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
)

func ClassifyEntries(root *metadata.Entry) error {

	if root.PathInfo.IsDir {
		switch role := classifyDir(root); role {
			case metadata.SubtitleDir:
				classifySubtitleDir(root)
			case metadata.BonusDir:
				classifyBonusDir(root)
			case metadata.SeasonDir:
				//nothing
			case metadata.SeriesDir:
				//nothing
			case metadata.MovieDir:
				//nothing
			default:
				//nothing

		}
	} else {
		root.Role = classifyFile(root)
	}
	
	return nil
}

/*
	classifier helper functions
*/


func classifySubtitleDir(entry *metadata.Entry) error {
	for _, child := range entry.Children {
		if isSubtitleFile(child) {
			child.Role = metadata.SubtitleFile
		} else {
			entry.Role = metadata.Unknown
			return fmt.Errorf("file %v could not be classified as a subtitle dir child", child.PathInfo.Source)
		}
	}

	entry.Role = metadata.SubtitleDir
	return nil
}

func classifyBonusDir(entry *metadata.Entry) error {

	bonus := false
	for _, child := range entry.Children {
		if isBonusFile(child) {
			bonus = true
			child.Role = metadata.BonusFile
		} else if isSubtitleFile(child) {
			child.Role = metadata.SubtitleFile
		} else {
			return fmt.Errorf("file %v could not be classified as a bonus dir child", child.PathInfo.Source)
		}
	}

	if !bonus {
		return fmt.Errorf("no required children of bonus dir found")
	}

	entry.Role = metadata.BonusDir	
	return nil
}

/*
func classifySeasonDir(entry *metadata.Entry) error {

	episode := false
	for _, child := range entry.Children {
		if isEpisodeFile(child) {
			episode = true
			child.Role = metadata.EpisodeFile
		} else if isSubtitleFile(child) {
			child.Role = metadata.SubtitleFile
		} else if isSubtitleDir(child) {
			classifySubtitleDir(child)
		} else {
			return fmt.Errorf("file %v could not be classified as a season dir child", child.PathInfo.Source)
		}
	}

	if !episode {
		return fmt.Errorf("no required children of season dir found")
	}

	entry.Role = metadata.SeasonDir	
	return nil
}
*/

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
	return metadata.Unknown
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
	return metadata.Unknown
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
