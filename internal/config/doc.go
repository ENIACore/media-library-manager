// Package config provides configuration management for the media library manager.
//
// # Config
//
// The main type, [Config], holds all application configuration including paths
// for torrents, movies, shows, and manager data, plus runtime flags like DryRun.
//
// # Load
//
// The [Load] function is a singleton that returns the global configuration,
// loaded once using [sync.OnceValue]. Configuration follows a precedence order:
//
//	command-line flags > environment variables > defaults
//
// # Environment Variables
//
//	ENIACORE_TORRENT_PATH    - Path to downloaded torrents
//	ENIACORE_MOVIE_PATH      - Path to movie library
//	ENIACORE_SHOW_PATH       - Path to show library
//	ENIACORE_MANAGER_PATH    - Path to program directory (logs, errors)
//	ENIACORE_DRY_RUN         - Run without moving files (true/false)
//
// # Pipeline
//
// Configuration is used throughout the entire pipeline to determine paths and behavior.
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
//	processor.Process ◄─ config (determines destination paths)
//	       │
//	       ▼
//	transfer.Transfer ─or─ transfer.Error ◄─ config (determines dry run behavior)
package config
