package metadata

// Entry represents a node in the media file hierarchy, forming a tree where each node contains metadata about file or directory
type Entry struct {
	Parent   *Entry
	Children []*Entry
	Depth    int // Depth depth from root (root = 0)

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
