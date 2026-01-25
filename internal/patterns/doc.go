// Package patterns provides compiled regex patterns for matching media metadata from torrent filenames.
//
// # Pattern Groups
//
// The package exports several pattern collections used by [extractor] to identify metadata:
//
//	LanguagePatternGroups    - Language detection (English, Spanish, French, etc.)
//	BonusPatternGroups       - Bonus content types (Behind.The.Scenes, Deleted.Scenes, etc.)
//	MiscPatterns             - Miscellaneous torrent metadata (quality, release groups, editions)
//
// Patterns are compiled once using [sync.OnceValue] for efficient reuse throughout the pipeline.
//
// # Pipeline
//
//	parser.Parse ◄─ metadata.Entry ─► extractor.ExtractMedia ─and─ extractor.ExtractPath
//	       │                                     │
//	       │                                     └── patterns (used here)
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
// # Usage
//
// Patterns are accessed through getter functions that return pre-compiled pattern groups:
//
//	GetLanguagePatternGroups()
//	GetBonusPatternGroups()
//	GetMiscPatterns()
package patterns
