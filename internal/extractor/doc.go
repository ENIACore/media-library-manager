// Package extractor provides extraction functions for media metadata and filesystem metadata of files and/or directories
//
// # ExtractMedia
//
// One of two main functions, [ExtractMedia] and [ExtractPath], that extract metadata from file or directory path.
// Used to help [parser.Parse] create [metadata.Entry] objects, to be passed along the media_library_manager pipeline.
//
// # ExtractPath
//
// One of two main functions, [ExtractMedia] and [ExtractPath], that extract metadata from file or directory path.
// Used to help [parser.Parse] create [metadata.Entry] objects, to be passed along the media_library_manager pipeline.
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
//	transfer.Transfer ─or─ transfer.Error
package extractor
