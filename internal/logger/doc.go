// Package logger provides structured logging functionality for the media library manager.
//
// # NewLogger
//
// The main function, [NewLogger], creates a multi-handler structured logger using [slog]
// that writes to separate log files based on severity level (DEBUG, INFO, WARN).
// Log files are organized by session timestamp in the manager directory.
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
//
// # Implementation
//
// The logger uses a custom multiHandler that implements the [slog.Handler] interface,
// routing log messages to multiple file handlers based on their severity level.
// Each session creates a timestamped directory containing DEBUG.log, INFO.log, and WARN.log files.
package logger
