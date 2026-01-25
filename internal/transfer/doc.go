// Package transfer handles moving media files to their final destinations or error directory.
//
// # Transfer
//
// The main function, [Transfer], moves successfully processed media entries
// from the torrent directory to the appropriate media library (movies or shows).
// Files are moved to paths determined by [processor.Process].
//
// # Error
//
// The [Error] function moves failed entries to the manager's error directory
// for manual review. This function is called when parsing, classification,
// enrichment, or processing fails.
//
// # Conflict Resolution
//
// If a destination file already exists, the package automatically resolves conflicts
// by appending a numeric suffix (_1, _2, etc.) to the filename.
//
// # Pipeline
//
//	parser.Parse ◄─ metadata.Entry ─► extractor.ExtractMedia ─and─ extractor.ExtractPath
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
//	transfer.Transfer ─or─ transfer.Error ◄─ moves files to final destination
//
// # Dry Run
//
// When [config.Config.DryRun] is true, both Transfer and Error return immediately
// without moving any files, allowing safe validation of the processing pipeline.
package transfer
