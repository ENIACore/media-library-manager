<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Version][version-shield]][version-url]
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![AGPL License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <h3 align="center">Media Library Manager</h3>

  <p align="center">
    An automated torrent media organizer that transforms chaotic downloads into a clean, standardized library structure for media servers like Jellyfin and Plex.
    <br />
    <a href="https://github.com/ENIACore/media_library_manager"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/ENIACore/media_library_manager/issues/new?labels=bug&template=bug-report---.md">Report Bug</a> &middot;
    <a href="https://github.com/ENIACore/media_library_manager/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
        <li><a href="#key-features">Key Features</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
        <li><a href="#configuration">Configuration</a></li>
      </ul>
    </li>
    <li>
      <a href="#modes">Modes</a>
      <ul>
        <li><a href="#ingest-mode-default">Ingest</a></li>
        <li><a href="#subtitle-mode">Subtitle</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#how-it-works">How It Works</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project

Media Library Manager is a powerful automation tool designed for media enthusiasts who download content via torrents. It automatically processes downloaded files and directories, extracting metadata from filenames, classifying content, and organizing everything into a clean, standardized structure that media servers like Jellyfin and Plex can easily index.

The tool runs in two modes — **ingest** (default) for processing new downloads and **subtitle** for retroactively adding missing English subtitles to an existing library.

> **v1.0-beta** — First public release. All pipeline functionality is complete, including subtitle language detection. Documentation is incomplete, the codebase needs refinement, and test coverage is not yet in place.

The tool handles complex scenarios including:
- Mixed movie and TV show libraries
- Multi-season TV series with bonus content
- Subtitle files in various languages
- Bonus features and extras
- Duplicate file resolution
- Error handling with automatic quarantine

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

* [![Go][Go-badge]][Go-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Key Features

- **Two Operating Modes**: Ingest mode for new downloads; subtitle mode for backfilling missing English subtitles
- **Intelligent Pattern Recognition**: Extracts titles, years, seasons, episodes, quality, codecs, and more from filenames
- **Automatic Classification**: Identifies movies, TV shows, subtitles, and bonus content
- **Jellyfin-Standard Naming**: Generates filenames and directory structures that conform to Jellyfin naming conventions
- **TMDb Verification by ID**: Confirms and enriches titles using TMDb ID lookups for precise matching
- **MKV Subtitle Stripping**: Removes all embedded subtitle tracks from MKV files; plaintext tracks (ASS, SSA, SRT, etc.) are extracted to standalone `.srt` files
- **Subtitle Language Detection**: Detects the actual language of subtitle files and names them accordingly
- **Flexible Structure Support**: Handles various torrent directory structures
- **Bonus Content Handling**: Properly categorizes extras, behind-the-scenes, and bonus features
- **Conflict Resolution**: Automatically resolves duplicate filenames
- **Error Quarantine**: Failed entries are moved to an error directory for manual review
- **Dry Run Mode**: Test operations without moving files
- **Comprehensive Logging**: Detailed logs for debugging and monitoring

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

* Go 1.25.5 or higher
  ```sh
  go version
  ```

### Installation

#### Quick Install (Recommended)

Run this command to download and install the latest `mlm` binary:

```bash
curl -fsSL https://raw.githubusercontent.com/ENIACore/media-library-manager/main/install.py -o /tmp/mlm-install.py && sudo python3 /tmp/mlm-install.py; rm -f /tmp/mlm-install.py
```

The script will automatically:
- Detect your platform (Linux/macOS, amd64/arm64)
- Download the latest release binary
- Install to `/usr/local/bin/mlm`
- Make it executable

After installation, you can run `mlm` from anywhere in your terminal.

#### Manual Installation

If you prefer to build from source:

1. Clone the repo
   ```sh
   git clone https://github.com/ENIACore/media_library_manager.git
   cd media_library_manager
   ```

2. Build and install using Make
   ```sh
   make install
   ```

   Or build manually:
   ```sh
   go build -o mlm cmd/media_library_manager/main.go
   sudo cp mlm /usr/local/bin/
   ```

#### Build Commands

The project includes a Makefile for convenient builds:

```sh
make build      # Build for current platform
make install    # Build and install to /usr/local/bin
make release    # Build for all platforms (Linux, macOS, Windows)
make test       # Run all tests
make clean      # Remove build artifacts
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Configuration

The application can be configured using command-line flags or environment variables. Environment variables take precedence over defaults, and command-line flags take precedence over both.

#### Environment Variables

```sh
# Both modes
export ENIACORE_MODE="ingest"
export ENIACORE_MOVIE_PATH="/path/to/movies"
export ENIACORE_SHOW_PATH="/path/to/shows"
export ENIACORE_MANAGER_PATH="/path/to/manager"
export ENIACORE_LOG_STDOUT="true"
export ENIACORE_DRY_RUN="false"
export ENIACORE_TMDB_API_KEY="your_tmdb_key"

# Ingest mode
export ENIACORE_TORRENT_PATH="/path/to/downloads"
export ENIACORE_INCOMPLETE_PATH="/path/to/downloads/temp"
export ENIACORE_INTERACTIVE="true"
export ENIACORE_LIMIT="10"

# Subtitle mode
export ENIACORE_OS_API_KEY="your_opensubtitles_api_key"
export ENIACORE_OS_USER_AGENT="your_app_user_agent"
export ENIACORE_OS_USER="your_opensubtitles_username"
export ENIACORE_OS_PASS="your_opensubtitles_password"
```

#### Command-Line Flags

```sh
# Both modes
mlm \
  -mode="ingest" \
  -movie-path="/opt/jellyfin/media/movies" \
  -show-path="/opt/jellyfin/media/shows" \
  -manager-path="/opt/media_manager" \
  -log-stdout=true \
  -dry-run=false \
  -tmdb-api-key="your_tmdb_key"

# Ingest mode flags
mlm \
  -torrent-path="/opt/qbit/downloads" \
  -incomplete-path="/opt/qbit/downloads/temp" \
  -interactive=true \
  -limit=10

# Subtitle mode flags
mlm -mode=subtitle \
  -os-api-key="your_opensubtitles_api_key" \
  -os-user-agent="your_app_user_agent" \
  -os-user="your_opensubtitles_username" \
  -os-pass="your_opensubtitles_password"
```

#### Default Values

**Both modes**

| Parameter | Env Var | Default | Description |
|-----------|---------|---------|-------------|
| `mode` | `ENIACORE_MODE` | `ingest` | Operating mode: `ingest` or `subtitle` |
| `movie-path` | `ENIACORE_MOVIE_PATH` | `/opt/jellyfin/media/movies` | Destination for movie files |
| `show-path` | `ENIACORE_SHOW_PATH` | `/opt/jellyfin/media/shows` | Destination for TV show files |
| `manager-path` | `ENIACORE_MANAGER_PATH` | `/opt/media_manager` | Program directory (for logs and errors) |
| `log-stdout` | `ENIACORE_LOG_STDOUT` | `true` | Log to standard output |
| `dry-run` | `ENIACORE_DRY_RUN` | `true` | Run without moving files |
| `tmdb-api-key` | `ENIACORE_TMDB_API_KEY` | `""` | TMDb API read access token or v3 key |

**Ingest mode**

| Parameter | Env Var | Default | Description |
|-----------|---------|---------|-------------|
| `torrent-path` | `ENIACORE_TORRENT_PATH` | `/opt/qbit/downloads` | Path to downloaded torrents |
| `incomplete-path` | `ENIACORE_INCOMPLETE_PATH` | `/opt/qbit/downloads/temp` | Path to incomplete torrents (skipped) |
| `interactive` | `ENIACORE_INTERACTIVE` | `true` | Allow interactive correction during processing |
| `limit` | `ENIACORE_LIMIT` | `10` | Max movies/series to process per run (0 = unlimited) |

**Subtitle mode**

| Parameter | Env Var | Default | Description |
|-----------|---------|---------|-------------|
| `os-api-key` | `ENIACORE_OS_API_KEY` | `""` | OpenSubtitles REST API key |
| `os-user-agent` | `ENIACORE_OS_USER_AGENT` | `""` | OpenSubtitles user agent |
| `os-user` | `ENIACORE_OS_USER` | `""` | OpenSubtitles username (enables authenticated downloads) |
| `os-pass` | `ENIACORE_OS_PASS` | `""` | OpenSubtitles password |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MODES -->
## Modes

### Ingest Mode (default)

Ingest mode processes new torrent downloads and moves them into the Jellyfin library. It runs the full pipeline: parse → remux → classify → verify → enrich → resolve → transfer.

**What it does:**
- Parses filenames and directory structures to extract metadata
- Strips all embedded subtitle tracks from MKV files; plaintext tracks (ASS, SSA, SRT, etc.) are extracted to standalone `.srt` files alongside the video
- Classifies each entry (movie, episode, subtitle, bonus content, etc.)
- Verifies titles against TMDb using TMDb ID for precise matching
- Formats filenames and paths according to Jellyfin naming conventions
- Transfers files to their destination in the movie or show library

```sh
mlm -dry-run=false
# equivalent to:
mlm -mode=ingest -dry-run=false
```

### Subtitle Mode

Subtitle mode scans an existing Jellyfin library for media files that are missing an English `.srt` subtitle, then fetches and downloads matching subtitles from OpenSubtitles. It processes entries in batches of 80 per run to stay within API rate limits.

**What it does:**
- Walks the configured movie and show paths
- Identifies media files that have no English `.srt` subtitle present
- Queries OpenSubtitles for a matching subtitle using the TMDb ID
- Downloads and saves the `.srt` file next to the media file
- Stops after processing 80 entries (re-run to continue)

```sh
mlm -mode=subtitle -dry-run=false
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

### Basic Usage

Run ingest mode with dry run enabled (default):
```sh
mlm
```

Or from source:
```sh
go run cmd/media_library_manager/main.go
```

### Ingest — Production

Disable dry run to actually move files:
```sh
mlm -dry-run=false
```

### Subtitle — Backfill Missing Subtitles

Scan the library and download up to 80 missing English subtitles:
```sh
mlm -mode=subtitle -dry-run=false
```

### Custom Paths

Specify custom paths for your setup:
```sh
mlm \
  -torrent-path="/mnt/downloads" \
  -movie-path="/mnt/media/movies" \
  -show-path="/mnt/media/shows" \
  -dry-run=false
```

### Example Output Structure

**Input:**
```
downloads/
├── The.Matrix.1999.1080p.BluRay.x264.DTS/
│   ├── The.Matrix.1999.1080p.BluRay.x264.DTS.mkv
│   ├── Subs/
│   │   └── English.srt
│   └── Extras/
│       └── Behind.The.Scenes.mkv
```

**Output:**
```
movies/
└── The.Matrix.1999/
    ├── The.Matrix.1999.1080p.x264.BluRay.DTS.mkv
    ├── Subtitles/
    │   └── The.Matrix.1999.English.srt
    └── Extras/
        └── The.Matrix.1999.Behind.The.Scenes.mkv
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- HOW IT WORKS -->
## How It Works

### Ingest Pipeline

The ingest pipeline runs in six stages:

#### 1. Parser
- Recursively scans the torrent directory
- Creates a tree structure of Entry objects representing files and directories
- Filters out unwanted files (samples, NFO files, etc.)
- Extracts initial metadata from filenames using pattern matching

#### 2. Remuxer
- Processes every MKV file in the tree
- Strips all embedded subtitle tracks using `mkvmerge`
- Plaintext subtitle formats (ASS, SSA, SRT, etc.) are extracted to standalone `.srt` files next to the video before being stripped, so no subtitle data is lost
- Image-based subtitle formats (PGS, VOBSUB) are discarded

#### 3. Classifier
- Traverses the Entry tree and assigns roles to each entry
- Classifies files as: MovieFile, EpisodeFile, SubtitleFile, or BonusFile
- Classifies directories as: MovieDir, SeriesDir, SeasonDir, SubtitleDir, or BonusDir
- Uses pattern recognition and tree structure analysis

#### 4. Verifier
- Queries the TMDb API to confirm and enrich each title
- Matches via TMDb ID for precise, unambiguous lookups
- Sets the `TMDBid` field on verified entries

#### 5. Resolver
- Determines final destination paths for all entries
- Builds filenames and directory structures that conform to Jellyfin naming conventions
- Groups subtitles and extras appropriately

#### 6. Transfer
- Moves files to their final destinations
- Creates necessary directory structures
- Resolves filename conflicts automatically
- Cleans up empty source directories
- Moves failed entries to error directory

### Subtitle Pipeline

The subtitle pipeline is a lightweight scan-and-fetch loop:

1. Walks the movie and show library paths
2. Finds media files with no English `.srt` subtitle present
3. Queries OpenSubtitles using the file's TMDb ID
4. Downloads and saves the best-matching `.srt` alongside the media file
5. Stops after 80 entries — re-run to continue

### Error Handling
- Failed entries are moved to the error directory for manual review
- Comprehensive logging tracks all operations
- Dry run mode allows testing without file operations

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap

- [x] Core processing pipeline
- [x] Movie and TV show support
- [x] Subtitle and bonus content handling
- [x] Pattern-based metadata extraction
- [x] Error handling and quarantine
- [x] Readable dry run output
- [x] Jellyfin-standard naming conventions
- [x] TMDb integration for title verification (ID-based)
- [x] MKV subtitle stripping with plaintext extraction to SRT
- [x] Subtitle mode — backfill missing English subtitles via OpenSubtitles
- [x] Subtitle language detection
- [x] Released v1.0-beta
- [ ] Refactor `Reclassify` function; add TV show reclassification support to `Reclassify`
- [ ] Full documentation
- [ ] Test coverage
- [ ] Codebase refinement and cleanup

See the [open issues](https://github.com/ENIACore/media_library_manager/issues) for a full list of proposed features and known issues.

<!-- CONTRIBUTING -->
## Contributing

Contributions make the open source community an amazing place to learn, inspire, and create. Any contributions you make are greatly appreciated.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the GNU AGPL v3 License. See `LICENSE.md` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

Project Link: [https://github.com/ENIACore/media_library_manager](https://github.com/ENIACore/media_library_manager)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
[version-shield]: https://img.shields.io/badge/version-v1.0--beta-blue?style=for-the-badge
[version-url]: https://github.com/ENIACore/media_library_manager/releases/tag/v1.0-beta
[contributors-shield]: https://img.shields.io/github/contributors/ENIACore/media_library_manager.svg?style=for-the-badge
[contributors-url]: https://github.com/ENIACore/media_library_manager/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/ENIACore/media_library_manager.svg?style=for-the-badge
[forks-url]: https://github.com/ENIACore/media_library_manager/network/members
[stars-shield]: https://img.shields.io/github/stars/ENIACore/media_library_manager.svg?style=for-the-badge
[stars-url]: https://github.com/ENIACore/media_library_manager/stargazers
[issues-shield]: https://img.shields.io/github/issues/ENIACore/media_library_manager.svg?style=for-the-badge
[issues-url]: https://github.com/ENIACore/media_library_manager/issues
[license-shield]: https://img.shields.io/github/license/ENIACore/media_library_manager.svg?style=for-the-badge
[license-url]: https://github.com/ENIACore/media_library_manager/blob/main/LICENSE.md
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/linkedin_username
[Go-badge]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://golang.org/

# Releasing a New Version

## Prerequisites
- [ ] `gh` CLI installed and authenticated (`gh auth status`)
- [ ] All changes committed and pushed
- [ ] On `main` branch (`git branch --show-current`)

## Steps

### 1. Build the release binaries
```bash
make release VERSION=v0.9.9
```

### 2. Verify binaries exist
```bash
ls -lh dist/
```

Expected output:
```
dist/mlm-darwin-amd64
dist/mlm-darwin-arm64
dist/mlm-linux-amd64
dist/mlm-linux-arm64
dist/mlm-windows-amd64.exe
```

### 3. Test your local platform binary
```bash
# macOS arm64
./dist/mlm-darwin-arm64 -help

# macOS amd64
./dist/mlm-darwin-amd64 -help

# Linux amd64
./dist/mlm-linux-amd64 -help
```

### 4. Tag the commit
```bash
git tag v0.9.9
git push origin v0.9.9
```

### 5. Create the GitHub release
```bash
./release.sh
```

### 6. Verify the release
```bash
curl -s https://api.github.com/repos/ENIACore/media-library-manager/releases/latest \
  | python3 -c "
import sys, json
r = json.load(sys.stdin)
print('Version:', r['tag_name'])
print('Assets:')
[print(' -', a['name']) for a in r['assets']]
"
```

### 7. Test the install script
```bash
curl -fsSL https://raw.githubusercontent.com/ENIACore/media-library-manager/main/install.py \
  -o /tmp/mlm-install.py && sudo python3 /tmp/mlm-install.py; rm -f /tmp/mlm-install.py
```

## Releasing a Subsequent Version

Repeat the steps above, updating the version in `release.sh`:
```bash
VERSION="v1.0.0"
```
