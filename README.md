<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
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
    <a href="https://github.com/ENIACore/media_library_manager/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
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

- **Intelligent Pattern Recognition**: Extracts titles, years, seasons, episodes, quality, codecs, and more from filenames
- **Automatic Classification**: Identifies movies, TV shows, subtitles, and bonus content
- **Standardized Naming**: Generates clean, consistent filenames following media server conventions
- **Flexible Structure Support**: Handles various torrent directory structures
- **Subtitle Management**: Organizes subtitle files by season and language
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
export ENIACORE_TORRENT_PATH="/path/to/downloads"
export ENIACORE_INCOMPLETE_PATH="/path/to/downloads/temp"
export ENIACORE_MOVIE_PATH="/path/to/movies"
export ENIACORE_SHOW_PATH="/path/to/shows"
export ENIACORE_MANAGER_PATH="/path/to/manager"
export ENIACORE_LOG_STDOUT="true"
export ENIACORE_DRY_RUN="false"
```

#### Command-Line Flags

```sh
mlm \
  -torrent-path="/opt/qbit/downloads" \
  -incomplete-path="/opt/qbit/downloads/temp" \
  -movie-path="/opt/jellyfin/media/movies" \
  -show-path="/opt/jellyfin/media/shows" \
  -manager-path="/opt/media_manager" \
  -log-stdout=true \
  -dry-run=false
```

#### Default Values

| Parameter | Default Value | Description |
|-----------|--------------|-------------|
| `torrent-path` | `/opt/qbit/downloads` | Path to downloaded torrents |
| `incomplete-path` | `/opt/qbit/downloads/temp` | Path to incomplete torrents (will be skipped) |
| `movie-path` | `/opt/jellyfin/media/movies` | Destination for movie files |
| `show-path` | `/opt/jellyfin/media/shows` | Destination for TV show files |
| `manager-path` | `/opt/media_manager` | Program directory (for logs and errors) |
| `log-stdout` | `true` | Log to standard output |
| `dry-run` | `true` | Run without moving files |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

### Basic Usage

Run with default configuration (dry run enabled):
```sh
mlm
```

Or if running from source:
```sh
go run cmd/media_library_manager/main.go
```

### Production Usage

Disable dry run to actually move files:
```sh
mlm -dry-run=false
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

The Media Library Manager uses a five-stage processing pipeline:

### 1. Parser
- Recursively scans the torrent directory
- Creates a tree structure of Entry objects representing files and directories
- Filters out unwanted files (samples, NFO files, etc.)
- Extracts initial metadata from filenames using pattern matching

### 2. Classifier
- Traverses the Entry tree and assigns roles to each entry
- Classifies files as: MovieFile, EpisodeFile, SubtitleFile, or BonusFile
- Classifies directories as: MovieDir, SeriesDir, SeasonDir, SubtitleDir, or BonusDir
- Uses pattern recognition and tree structure analysis

### 3. Enricher
- Propagates contextual information throughout the tree
- Passes titles and years from parent to child entries
- Enriches episode files with season information
- Handles intermediary subtitle and bonus directories

### 4. Resolver
- Determines final destination paths for all entries
- Builds standardized filenames with proper capitalization
- Organizes content according to media server conventions
- Groups subtitles and extras appropriately

### 5. Transfer
- Moves files to their final destinations
- Creates necessary directory structures
- Resolves filename conflicts automatically
- Cleans up empty source directories
- Moves failed entries to error directory

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
- [ ] Readable dry run output
- [ ] API integration (TMDB, TVDB) for metadata enrichment
- [ ] Custom pattern configuration
- [ ] Advanced duplicate detection
- [ ] Advanced collision avoidance
- [ ] Advanced aggregator
- [ ] Advanced language detection (subtitles)

See the [open issues](https://github.com/ENIACore/media_library_manager/issues) for a full list of proposed features and known issues.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


## TODO
- Add better dry run output
- Fix Test batch 4 (1) failing test
- Fix Test batch 1, "Walk The Line" Line is being recognized as an Audio value
- Add advanced aggregator (e.g If year missing from one show, combine under the show with the year ect...)
- Detect duplicates
- TMDB integration to ensure movie title correctness
- Language detection for subtitles to correct file names

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
