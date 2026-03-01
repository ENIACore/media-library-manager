package metadata

// MediaInfo describes the movie, show, or episode the related file or directory pertains too.
// nil or "" indicates missing pattern or information in file or directory.
type MediaInfo struct {
	Title      []string
	Year       *int
	Episode    []int  // List of episode numbers (e.g., [1,2,3] for episodes 1-3)
	Season     *int   // 0 = pattern found w/o number, >0 = season number
	Resolution string
	Codec      string
	Source     string
	Audio      string
	Language   []string
	DS         string // Deleted scenes
	BTS	       string // Behind the scenes
	Bonus      string // Less specific bonus (i.e featurettes, extras ect)
	Edition	   string
}

// PathInfo describes the file or directory related information.
// All paths are full paths and not relative.
// [ContentType] has a value of [UnknownType] for directories.
type PathInfo struct {
	Dest   	string
	Source 	string

	Ext   string
	Type  ContentType
	IsDir bool
}
