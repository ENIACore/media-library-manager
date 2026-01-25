package metadata

// MediaInfo contains metadata extracted from a media file or directory name.
type MediaInfo struct {
	Title      []string
	Year       *int   // nil if not found
	Episode    *int   // nil = no pattern, 0 = pattern but no number, >0 = episode number
	Season     *int   // nil = no pattern, 0 = pattern but no number, >0 = season number
	Resolution string // empty if not found
	Codec      string
	Source     string
	Audio      string
	Language   string
	Bonus      string
}

// PathInfo contains filesystem path information for an entry.
type PathInfo struct {
	Dest   string // destination path set by processor
	Source string // original source path

	Ext   string      // file extension, empty if none
	Type  ContentType // content type, UnknownType if directory or no extension
	IsDir bool
}
