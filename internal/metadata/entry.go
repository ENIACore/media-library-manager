package metadata

// Entry represents a node in the media file hierarchy.
//
// Entry forms a tree structure where each node contains metadata
// about a file or directory and its relationship to other entries.
type Entry struct {
	Parent   *Entry
	Children []*Entry
	Depth    int // depth from root (root = 0)

	MediaInfo MediaInfo
	PathInfo  PathInfo
	Role      EntryRole
}

// Height returns the maximum depth from this entry to its deepest descendant.
// Returns 0 for leaf nodes (entries with no children).
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
