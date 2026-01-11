package metadata

type Entry struct {
	Parent		*Entry
	Children	[]*Entry
	Depth		int			// Root level entry should be Depth 0

	MediaInfo	MediaInfo
	PathInfo	PathInfo
}

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
