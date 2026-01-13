package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

func ClassifyEntries(root *metadata.Entry) error {
	return nil
}

// returns classification of entry or Unknown otherwise
func classifyEntryFile(entry *metadata.Entry) metadata.EntryRole {

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

// returns classification of entry or Unknown otherwise
func classifyEntryDir(entry *metadata.Entry) metadata.EntryRole {
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

// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isMovieFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Episode == nil && entry.MediaInfo.Season == nil && entry.MediaInfo.Bonus == ""
}

// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isEpisodeFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Video && (entry.MediaInfo.Episode != nil || entry.MediaInfo.Season != nil) && entry.MediaInfo.Bonus == ""
}

// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isSubtitleFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Subtitle
}

// Helper function, not completely accurate by itself and must be used in conjunction with other functions to accureately determine file type
func isBonusFile(entry *metadata.Entry) bool {
	return entry.PathInfo.Type == metadata.Video && entry.MediaInfo.Bonus != ""
}

/*
Subtitle Directory
└── Subtitle File(s)
*/
// Returns true if atleast one subtitle file and all child entries are subtitle files
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

/*
Bonus Directory
├── Bonus File(s)
└── Subtitle File(s) (optional)
*/
// Returns true if atleast one bonus file, and all child entries are bonus or subtitle files
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

/*
Movie Directory
├── Movie File
├── Subtitle File(s) (optional)
├── Bonus File(s) (optional)
├── Subtitle Directory (optional)
└── Bonus Directory (optional)
*/
// Returns true if there is exactly one movie file, and all child entries are movie file, subtitle file, bonus file, subtitle directory or bonus directory 
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

/*
Season Directory
├── Episode File(s)
├── Subtitle File(s) (optional)
└── Subtitle Directory (optional)
*/
func isSeasonDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 2 {
		return false
	}

	episode := false
	for _, child := range entry.Children {
		if isSubtitleFile(child) || isSubtitleDir(child) {
			continue
		}

		if isEpisodeFile(child) {
			episode = true
		} else {
			return false	
		}
	}

	return entry.MediaInfo.Season != nil || episode
}

/*
Series Directory
├── Season Directory(s)
├── Bonus Directory (optional)
└── Subtitle Directory (optional)
*/
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
