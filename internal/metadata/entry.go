package metadata

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
