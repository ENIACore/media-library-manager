package classifier

import (
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"fmt"
)

/*
Subtitle Directory
└── Subtitle File(s)
*/
func isSubtitleDir(entry *metadata.Entry) bool {
	if !entry.IsDir {
		return false
	}
	// Subtitle directory cannot have nested directories
	if entry.Height() > 1 {
		return false;
	}

	// All files in subtitle directory should be of type subtitle
	for _, child := range entry.Children {
		if child.Type != metadata.Subtitle {
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
	if !entry.IsDir {
		return false
	}
	// Bonus directory cannot have nested directories
	if entry.Height() > 1 {
		fmt.Printf("For entry %v returning false due to height greater than 1", entry.Source)	
		return false;
	}

	// All files must be type subtitle or type video with either the directory or file containing a 'bonus' pattern
	for _, child := range entry.Children {
		if child.Type == metadata.Subtitle {
			continue
		}
		if child.Type != metadata.Video && (entry.Bonus == "" && child.Bonus == ""){
			return false
		}
	}

	return true
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
