package parser

import (
	"os"
	"fmt"
	"path/filepath"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"log/slog"
)

func ParseTree(path string, parent *metadata.Entry, depth int, logger *slog.Logger) (*metadata.Entry, error) {

    info, err := os.Stat(path)
    if err != nil {
		return nil, fmt.Errorf("stat path %s, %w", path, err)
    }

    node := &metadata.Entry{
        Parent:		parent,
		Depth:		depth,
		MediaInfo: extractor.ExtractMedia(path, logger),
		PathInfo: extractor.ExtractPath(path, logger),
    }

	// If node is invalid file (txt, jpg, etc), skip it
	if node.PathInfo.Type == metadata.UnknownType && !node.PathInfo.IsDir {
		return nil, nil
	} 
	if !info.IsDir() {
		return node, nil
	} 

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read dir %s, %w", path, err)
	}

	children := make([]*metadata.Entry, 0, len(entries))
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		child, err := ParseTree(childPath, node, depth + 1, logger)
		if err != nil {
			return nil, err
		} else if child != nil {
			children = append(children, child)
		}
	}
	node.Children = children
    
    return node, nil
}
