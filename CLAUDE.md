# Media Library Manager

Automatically processes media files downloaded via qBittorrent and moves them into a Jellyfin library. It parses filenames to extract metadata, verifies against TMDb, and transfers files to the correct destination paths.

## Module

```
github.com/ENIACore/media_library_manager
```

## Pipeline

Executed per entry in `cfg.TorrentPath` inside `cmd/media_library_manager/main.go`:

```
parser → remuxer → classifier → verifier → [enricher] → [resolver] → [transfer]
```

Bracketed stages are implemented but commented out in main.go.

## Packages

| Package | Responsibility |
|---|---|
| `parser` | Walks file system path, builds `*metadata.Entry` tree via extractor |
| `remuxer` | Strips subtitle tracks from MKV files in-place using mkvmerge |
| `classifier` | Assigns `metadata.EntryRole` to every node in the entry tree |
| `verifier` | Queries TMDb API to confirm title/year; sets `MediaInfo.TMDBid` |
| `enricher` | Fetches extended metadata (not yet active) |
| `resolver` | Computes `FileInfo.DestPath` for each entry (not yet active) |
| `transfer` | Moves files to `DestPath`; handles cleanup and error directory |
| `extractor` | Parses filenames (`ExtractMedia`) and file metadata via ffprobe (`ExtractFile`) |
| `classifier` | Assigns role based on children, filename patterns, and tree height |
| `patterns` | Regex pattern groups for title, season, episode, resolution, codec, etc. |
| `metadata` | Core data structures: `Entry`, `MediaInfo`, `FileInfo`, `EntryRole`, `ContentType` |
| `config` | Loads config from flags/env vars; single instance via `sync.OnceValue` |
| `logger` | `slog` wrapper |

## Core Data Structures

### `metadata.Entry`
```go
type Entry struct {
    Parent   *Entry
    Children []*Entry
    Depth    int        // 0 = root

    MediaInfo MediaInfo
    FileInfo  FileInfo
    Role      EntryRole  // set by classifier — not available before classifier runs
}
```

### `metadata.FileInfo`
```go
type FileInfo struct {
    SourcePath  string
    DestPath    string     // set by resolver
    IsDir       bool
    Ext         string     // uppercase, no dot: "MKV", "MP4", "SRT"
    ContentType ContentType

    Resolution  string     // "4K", "1080p", "720p", etc.
    Codec       string     // "hevc", "h264", "vp9", etc.
    Audio       string     // "aac", "ac3", "dts", etc.
    Language    []string   // ["ENG", "SPA"]
    Bitrate     string     // "5000 kbps"
}
```

### `metadata.MediaInfo`
```go
type MediaInfo struct {
    Title   []string   // uppercase words: ["THE", "DARK", "KNIGHT"]
    Year    *int
    Episode *int       // nil = not found, 0 = pattern found w/o number
    Season  *int
    DS      string     // Deleted Scenes
    BTS     string     // Behind the Scenes
    Bonus   string
    Edition string
    TMDBid  string     // "tmdb-12345" after verifier runs
}
```

### `metadata.EntryRole`
Files: `UnknownRole`, `SubtitleFile`, `DSFile`, `BTSFile`, `BonusFile`, `EpisodeFile`, `MovieFile`
Dirs: `SubtitleDir`, `DSDir`, `BTSDir`, `BonusDir`, `SeasonDir`, `SeriesDir`, `MovieDir`

### `metadata.ContentType`
`UnknownType`, `Video`, `Subtitle`

## Config Fields

| Field | Env Var | Default |
|---|---|---|
| `TorrentPath` | `ENIACORE_TORRENT_PATH` | `/opt/qbit/downloads` |
| `IncompletePath` | `ENIACORE_INCOMPLETE_PATH` | `/opt/qbit/downloads/temp` |
| `MoviePath` | `ENIACORE_MOVIE_PATH` | `/opt/jellyfin/media/movies` |
| `ShowPath` | `ENIACORE_SHOW_PATH` | `/opt/jellyfin/media/shows` |
| `ManagerPath` | `ENIACORE_MANAGER_PATH` | `/opt/media_manager` |
| `DryRun` | `ENIACORE_DRY_RUN` | `true` |
| `Interactive` | `ENIACORE_INTERACTIVE` | `true` |
| `TMDBApiKey` | `ENIACORE_TMDB_API_KEY` | `""` |

## Conventions

**Package entry point signature:**
```go
func DoThing(root *metadata.Entry, cfg *config.Config, logger *slog.Logger) error
```

**Logger setup:**
```go
lg := logger.With("func", "FuncName", "source", root.Source())
```

**Error format:**
```go
fmt.Errorf("packagename: descriptive message: %w", err)
```

**Tree walking:** recurse via `entry.Children`; use `entry.Height()` to limit depth; check `entry.FileInfo.IsDir` to branch on file vs directory.

**Role is not set before classifier.** Packages that run before classifier (parser, remuxer) must use `FileInfo.Ext`, `FileInfo.ContentType`, or `MediaInfo` fields instead of `entry.Role`.

**`FileInfo.Ext` is always uppercase** (e.g., `"MKV"`, `"MP4"`, `"SRT"`).

## External Tool Dependencies

- `ffprobe` — required at startup; used by extractor to probe video streams
- `mkvmerge` — required at startup; used by remuxer to strip subtitle tracks

Both are checked via `exec.LookPath` and panic at startup if missing.
