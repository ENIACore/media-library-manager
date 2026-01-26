// Package enricher propagates metadata throughout the entry tree.
//
// # Enrich
//
// The main function, [Enrich], analyzes a classified [metadata.Entry] tree and
// propagates title and year information to all entries, ensuring consistent
// metadata across the entire hierarchy.
//
// # Enrichment Logic
//
// The enricher determines title and year at the root level, then propagates
// these values to all descendants. The resolution strategy depends on content type:
//
//	Shows:   Uses getShowTitle and getShowYear
//	Movies:  Uses getMovieTitle and getMovieYear
//
// # Precedence Rules
//
// Title and year are resolved by searching the tree with the following precedence:
//
//	Shows:   Root entry → child SeasonDir → descendant EpisodeFile → nil
//	Movies:  Root entry → child MovieFile → nil
//
// Parent entries take precedence over their children. For example, if a SeriesDir
// has a title set, it takes precedence over titles in child SeasonDir or EpisodeFile
// entries. If the root has no title, the enricher recursively searches children until
// a value is found.
//
// Once title and year are determined, they are propagated to all descendants via
// setEntryValues, ensuring subtitle files, bonus files, and their containing
// directories inherit the same metadata.
//
// # Pipeline
//
//	parser.Parse ◄─ metadata.Entry ─► extractor.ExtractMedia ─and─ extractor.ExtractPath
//	       │
//	       ▼
//	classifier.Classify
//	       │
//	       ▼
//	enricher.Enrich ◄─ propagates title/year to all entries
//	       │
//	       ▼
//	processor.Process
//	       │
//	       ▼
//	transfer.Transfer ─or─ transfer.Error
//
// # Implementation
//
// The package uses role-specific getter functions (getShowTitle, getShowYear,
// getMovieTitle, getMovieYear) to extract metadata based on content type,
// then recursively applies these values to all entries via setEntryValues.
//
// # Constraints
//
// Enrichment can only be performed on root entries with valid media roles
// (MovieFile, MovieDir, EpisodeFile, SeasonDir, SeriesDir). Attempting to
// enrich subtitle or bonus content at the root level returns an error.
package enricher
