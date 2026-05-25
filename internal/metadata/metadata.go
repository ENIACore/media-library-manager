package metadata

import (
	"strconv"
)

// MediaInfo describes the movie, show, or episode the related file or directory pertains too.
// nil or "" indicates missing pattern or information in file or directory.
type MediaInfo struct {
	Title      []string
	Year       *int
	Episode    *int   // 0 = pattern found w/o number, >0 = episode number
	Season     *int   // 0 = pattern found w/o number, >0 = season number

	DS         string // Deleted scenes
	BTS	       string // Behind the scenes
	Bonus      string // Less specific bonus (i.e featurettes, extras ect)
	Edition	   string
	Source     string

	TMDBid	   int
}

// PathInfo describes the file or directory related information.
// All paths are full paths and not relative.
// [ContentType] has a value of [UnknownType] for directories.
type FileInfo struct {
	DestPath 	string
	SourcePath 	string

	IsDir 		bool
	Ext   		string
	ContentType ContentType

	Resolution 	string
	Codec      	string
	Audio      	string
	AspectRatio	string
	Language	[]string
	Bitrate		string
	BitDepth	string
}

func (info *MediaInfo) YearString() string {
	if info.Year == nil {
        return "nil"
    }
    return strconv.Itoa(*info.Year)
}

func (info *MediaInfo) EpisodeString() string {
	if info.Episode == nil {
        return "nil"
    }
    return strconv.Itoa(*info.Episode)
}

func (info *MediaInfo) SeasonString() string {
	if info.Season == nil {
        return "nil"
    }
    return strconv.Itoa(*info.Season)
}
