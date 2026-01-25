// Package main is the entry point for the media library manager application.
//
// # Overview
//
// The media library manager processes downloaded torrent files and organizes them
// into a structured media library suitable for media servers like Jellyfin or Plex.
//
// # Processing Pipeline
//
// For each entry in the torrent directory, the application executes:
//
//	1. parser.Parse          - Build metadata tree from filesystem
//	2. classifier.Classify   - Determine role of each entry (movie, episode, etc.)
//	3. enricher.Enrich       - Fetch additional metadata from external APIs
//	4. processor.Process     - Resolve destination paths based on classification
//	5. transfer.Transfer     - Move files to media library (or error directory on failure)
//
// # Configuration
//
// The application uses [config.Load] to load configuration from environment variables
// and command-line flags. See [config] package documentation for available options.
//
// # Logging
//
// Structured logging is provided by [logger.NewLogger], which writes to separate
// log files by severity level in the manager directory.
//
// # Error Handling
//
// If any stage of processing fails, the entry is moved to the error directory
// using [transfer.Error] for manual review. The pipeline continues with the next entry.
//
// # Dry Run Mode
//
// When --dry-run flag is enabled, the application validates processing logic without
// moving any files, logging what would happen instead.
package main
