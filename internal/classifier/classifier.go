package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
)

/*
Subtitle Directory
└── Subtitle File(s)
*/
func isSubtitleDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir {
		return false
	}
	// Subtitle directory cannot have nested directories
	if entry.Height() > 1 {
		return false;
	}

	// All files in subtitle directory should be of type subtitle
	for _, child := range entry.Children {
		if child.PathInfo.Type != metadata.Subtitle {
			return false;
		}
	}

	return true
}

/*
Bonus Directory
├── Bonus File(s)
└── Subtitle File(s) (optional)
*/
func isBonusDir(entry *metadata.Entry) bool {
	if !entry.PathInfo.IsDir || entry.Height() > 1 {
		return false
	}

	hasBonus := false
	for _, child := range entry.Children {
		switch child.PathInfo.Type {
			case metadata.Subtitle:
				// Subtitles are allowed, continue
			case metadata.Video:
				// Videos must have bonus pattern on either file or parent dir
				if entry.MediaInfo.Bonus == "" && child.MediaInfo.Bonus == "" {
					return false
				}
				hasBonus = true
			default:
				return false
		}
	}

	return hasBonus
}

/*
Movie Directory
├── Movie File
├── Subtitle File(s) (optional)
├── Bonus File(s) (optional)
├── Subtitle Directory (optional)
└── Bonus Directory (optional)
*/
func isMovieDir(entry *metadata.Entry) bool {
	return false
}

/*
Season Directory
├── Episode File(s)
├── Subtitle File(s) (optional)
└── Subtitle Directory (optional)
*/
func isSeasonDir(entry *metadata.Entry) bool {
	return false
}

/*
Series Directory
├── Season Directory(s)
├── Bonus Directory (optional)
└── Subtitle Directory (optional)
*/
func isSeriesDir(entry *metadata.Entry) bool {
	return false
}
