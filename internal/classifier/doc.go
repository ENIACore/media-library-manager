// Package classifier determines the role of media entries in the metadata tree.
//
// # Classify
//
// The main function, [Classify], analyzes a [metadata.Entry] tree and assigns
// each entry a [metadata.Role] based on its type, content, and structure.
//
// # Classification Roles
//
// Entries are classified into one of the following roles:
//
//	Files:        SubtitleFile, BonusFile, EpisodeFile, MovieFile
//	Directories:  SubtitleDir, BonusDir, SeasonDir, SeriesDir, MovieDir
//
// # Classification Logic
//
// Classification is performed in order of specificity, examining:
//
//	- File extension and type (video, subtitle)
//	- Presence of season/episode numbers
//	- Presence of bonus content indicators
//	- Directory structure and children types
//	- Tree height (depth from root)
//
// # Pipeline
//
//	parser.Parse ◄─ metadata.Entry ─► extractor.ExtractMedia ─and─ extractor.ExtractPath
//	       │
//	       ▼
//	classifier.Classify ◄─ assigns Role to each entry
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
// # Implementation
//
// The package uses role-specific classification functions (classifyMovieFile,
// classifyEpisodeFile, etc.) that validate entries against expected patterns
// and structural constraints defined in [metadata] package documentation.
package classifier
