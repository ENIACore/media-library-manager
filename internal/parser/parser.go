package parser

import (
	"os"
	"fmt"
	"path/filepath"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"log/slog"
)

// Parse constructs a tree representation of the filesystem starting at path.
// Recursively walks the directory structure, extracting media and path metadata for each entry.
// Returns the root Entry after pruning empty directories, or an error if parsing fails.
func Parse(path string, logger *slog.Logger) (*metadata.Entry, error) {
	root, err := parseTree(path, nil, 0, logger)
	if err != nil {
		return nil, err
	}
	root = pruneTree(root)
	return root, nil
}

// parseTree recursively builds a tree of Entry nodes from a filesystem path.
// Extracts metadata for files and directories, skipping invalid file types (UnknownType non-directories).
// Returns nil for skipped entries, allowing the tree to filter them out during construction.
func parseTree(path string, parent *metadata.Entry, depth int, logger *slog.Logger) (*metadata.Entry, error) {

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
		child, err := parseTree(childPath, node, depth + 1, logger)
		if err != nil {
			return nil, err
		} else if child != nil {
			children = append(children, child)
		}
	}
	node.Children = children
    
    return node, nil
}

// pruneTree removes empty directories from the entry tree.
// Recursively processes children first, then removes the entry if it's an empty directory.
// Returns nil if the entry should be pruned, otherwise returns the entry with pruned children.
func pruneTree(entry *metadata.Entry) *metadata.Entry {
	if entry.PathInfo.IsDir && len(entry.Children) == 0 {
		return nil
	}

	if !entry.PathInfo.IsDir {
		return entry
	}

	children := make([]*metadata.Entry, 0, len(entry.Children))
	for _, child := range entry.Children {
		child = pruneTree(child)
		if child != nil {
			children = append(children, child)
		}
	}

	entry.Children = children
	if len(entry.Children) == 0 {
		return nil
	}
	return entry
}
