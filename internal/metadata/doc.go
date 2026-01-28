// Package metadata provides types for representing media file torrents
// and their metadata in a hierarchical structure.
//
// # Entry
//
// The main type, [Entry], represents a node in the media file hierarchy.
// It is passed along the media_library_manager pipeline and transformed until its destination is determined.
//
// # Pipeline
//
//	parser.Parse ◄─ metadata.Entry ─► extractor.Extract
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
// # Supported Structures
//
//	Episode File                        (root or embedded)
//	Movie File                          (root or embedded)
//
//	Subtitle Directory                  (embedded only)
//	├── Subtitle File(s)
//	└── Subtitle Directory(s)           (optional) (max depth = 1)
//
//	Bonus Directory                     (embedded only)
//	├── Bonus File(s)
//	├── Subtitle File(s)                (optional)
//	└── Bonus Directory(s)           	(optional) (max depth = 1)
//
//	Season Directory                    (root or embedded)
//	├── Episode File(s)
//	├── Subtitle File(s)                (optional)
//	└── Subtitle Directory              (optional)
//
//	Series Directory                    (root only)
//	├── Season Directory(s)
//	├── Bonus Directory                 (optional)
//	└── Subtitle Directory              (optional)
//
//	Movie Directory                     (root only)
//	├── Movie File
//	├── Subtitle File(s)                (optional)
//	├── Bonus File(s)                   (optional)
//	├── Subtitle Directory              (optional)
//	└── Bonus Directory                 (optional)
//
// Note: Subtitle File and Bonus File cannot exist at root level.
// Note: Sample directories and files are ignored
// Note: Allowed to embed 1 level of subtitle directories inside subtitle directories, typically to encapsulate specific episodes (those intermediary directories will be removed in the final output
// Note: Allowed to embed 1 level of bonus directories inside bonus directories, typically to encapsulate specific episodes (those intermediary directories will be removed in the final output
package metadata
