package metadata

import (
	"path/filepath"
	"strconv"
)

// Entry represents a file or directory in a tree datastructure.
// It contains media, path, and role information used by other packages.
type Entry struct {
	Parent   *Entry
	Children []*Entry
	Depth    int // Depth depth from root (root = 0)

	MediaInfo MediaInfo
	FileInfo  FileInfo
	Role      EntryRole
}

func (entry *Entry) Source() string {
	if entry == nil || entry.FileInfo.SourcePath == "" {
		return "no source"
	}
	return filepath.Base(entry.FileInfo.SourcePath)
}

func (entry *Entry) Dest() string {
	if entry == nil || entry.FileInfo.DestPath == "" {
		return "no destination"
	}
	return filepath.Base(entry.FileInfo.DestPath)
}

func (entry *Entry) Episode() string {
	if entry == nil || entry.MediaInfo.Episode == nil {
		return "not set"
	}
	return strconv.Itoa(*entry.MediaInfo.Episode)
}

func (entry *Entry) Season() string {
	if entry == nil || entry.MediaInfo.Season == nil {
		return "not set"
	}
	return strconv.Itoa(*entry.MediaInfo.Season)
}

func (entry *Entry) TMDB() string {
	if entry == nil || entry.MediaInfo.TMDBid == 0 {
		return ""
	}
	return "tmdb-" + strconv.Itoa(entry.MediaInfo.TMDBid)
}

// Height returns the maximum depth from this entry to its deepest descendant, starting at 0.
func (entry *Entry) Height() int {
	if entry.Children == nil {
		return 0
	}

	maxHeight := 0
	for _, child := range entry.Children {
		maxHeight = max(maxHeight, child.Height())
	}
	return maxHeight + 1
}
