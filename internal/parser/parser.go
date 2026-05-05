// Package parser creates [metadata.Entry] tree by recursively parsing file system path.
// Package relies on extractor package to create tree and produces output used by classifier package.  
package parser

import (
	"os"
	"fmt"
	"path/filepath"
	"github.com/ENIACore/media_library_manager/internal/metadata"
	"github.com/ENIACore/media_library_manager/internal/extractor"
	"log/slog"
)

// Parse constructs [metadata.Entry] tree starting at given path.
// Throws error if file system error occurs or no valid [metadata.Entry] object can be created.
// Returns root of [metadata.Entry] tree.
func Parse(path string, dryRun bool, logger *slog.Logger) (*metadata.Entry, error) {
	root, err := parseTree(path, nil, 0, dryRun, logger)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("Parsed tree for path %v is nil", path)
	}

	root = pruneTree(root)
	if root == nil {
		return nil, fmt.Errorf("Pruned tree for path %v is nil", path)
	}

	return root, nil
}

// parseTree recursively builds a tree of [metadata.Entry] objects from given path and is a helper function to [Parse].
// It skips creating [metadata.Entry] object when [extractor.Filter] returns true.
// It uses [extractor.ExtractPath] and [extractor.ExtractMedia] to extract media and path information related to entry.
func parseTree(path string, parent *metadata.Entry, depth int, dryRun bool, logger *slog.Logger) (*metadata.Entry, error) {

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat path %s, %w", path, err)
	}

	// If node is a file or dir to filter skip it
	if extractor.Filter(path, logger) {
		return nil, nil
	}

	mediaInfo := extractor.ExtractMedia(path, logger)
	fileInfo, err := extractor.ExtractFile(path, dryRun, logger)
	if err != nil {
		return nil, err
	}
	node := &metadata.Entry{
		Parent:    parent,
		Depth:     depth,
		MediaInfo: mediaInfo,
		FileInfo:  fileInfo,
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
		child, err := parseTree(childPath, node, depth+1, dryRun, logger)
		if err != nil {
			return nil, err
		} else if child != nil {
			children = append(children, child)
		}
	}
	node.Children = children
    
    return node, nil
}

// pruneTree recursively removes empty directories from [metadata.Entry] tree.
func pruneTree(entry *metadata.Entry) *metadata.Entry {
	if entry.FileInfo.IsDir && len(entry.Children) == 0 {
		return nil
	}

	if !entry.FileInfo.IsDir {
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
