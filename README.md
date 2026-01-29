# media_library_manager

# TODO
- Add subtitle patterns (classifier)
- Add differentiation of episode title and series title on episode files
- Allow configuration of "movies" and "shows" subdir names


# To comment (in order)
- parser
- classifier
- enricher
- resolver
- transfer
- main

# Temporary

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
