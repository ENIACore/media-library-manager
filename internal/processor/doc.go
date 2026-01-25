// Package processor resolves destination paths for media entries based on their classification.
//
// # Resolve
//
// The main function, [Resolve], recursively processes a [metadata.Entry] tree and sets
// destination paths for all files and directories according to media library conventions.
//
// # Path Resolution
//
// The processor builds destination paths using media metadata:
//
//	Movies:      {MoviePath}/{Title}.{Year}/{Title}.{Year}.{Quality}.{Codec}.{ext}
//	Episodes:    {ShowPath}/{Title}.{Year}/S{Season}/{Title}.{Year}.S{Season}E{Episode}.{Quality}.{ext}
//	Bonus:       {BasePath}/Extras/{Title}.{Year}.{BonusType}.{Quality}.{ext}
//	Subtitles:   {BasePath}/Subtitles/{Title}.{Year}.{Language}.{ext}
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
//	processor.Process ◄─ sets PathInfo.Dest for all entries
//	       │
//	       ▼
//	transfer.Transfer ─or─ transfer.Error
//
// # Implementation
//
// The package contains role-specific resolvers (resolveMovieFile, resolveEpisodeFile, etc.)
// that are called recursively based on each entry's [metadata.Role].
// Helper functions like buildFilename and buildTitlePath handle string formatting.
package processor
