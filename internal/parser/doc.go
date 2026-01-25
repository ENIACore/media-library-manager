// Package parser builds a hierarchical tree of metadata entries from filesystem paths.
//
// # Parse
//
// The main function, [Parse], recursively walks a file or directory path and creates
// a tree of [metadata.Entry] nodes. Each node contains media metadata extracted by
// [extractor.ExtractMedia] and path information from [extractor.ExtractPath].
//
// # Tree Building
//
// The parser performs two operations:
//
//	1. parseTree    - Recursively builds the entry tree from the filesystem
//	2. pruneTree    - Removes empty directories and invalid files from the tree
//
// Invalid files (non-media files like .txt, .jpg) are skipped during parsing.
//
// # Pipeline
//
//	parser.Parse ◄─ builds metadata.Entry tree from filesystem
//	       │
//	       ▼
//	classifier.Classify
//	       │
//	       ▼
//	enricher.Enrich
//	       │
//	       ▼
//	processor.Process
//	       │
//	       ▼
//	transfer.Transfer ─or─ transfer.Error
//
// # Entry Structure
//
// Each [metadata.Entry] maintains parent-child relationships, depth information,
// and tracks whether it represents a file or directory. This tree structure is
// passed through the entire pipeline for classification, enrichment, and processing.
package parser
