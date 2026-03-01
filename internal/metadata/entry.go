package metadata

import (
	"path/filepath"
	"strings"
	"strconv"
)

// Entry represents a file or directory in a tree datastructure.
// It contains media, path, and role information used by other packages.
type Entry struct {
	Parent   *Entry
	Children []*Entry
	Depth    int // Depth depth from root (root = 0)

	MediaInfo MediaInfo
	PathInfo  PathInfo
	Role      EntryRole
}

func (entry *Entry) Source() string {
	if entry == nil || entry.MediaInfo.Source == "" {
		return "no source"
	}
	return filepath.Base(entry.MediaInfo.Source)
}

func (entry *Entry) Episode() string {
	if entry == nil || len(entry.MediaInfo.Episode) == 0 {
		return "not set"
	}
	episodes := make([]string, len(entry.MediaInfo.Episode))
	for i, ep := range entry.MediaInfo.Episode {
    	episodes[i] = strconv.Itoa(ep)
	}
	return strings.Join(episodes, ", ")
}

func (entry *Entry) Season() string {
	if entry == nil || entry.MediaInfo.Season == nil {
		return "not set"
	}
	return strconv.Itoa(*entry.MediaInfo.Season)
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
