# Jellyfin Naming Conventions

> Derived from Jellyfin documentation and internal resolver logic.
> Separators used throughout: `.` (period) between all filename components.

---

## Movies

### Directory
```
<Title>.<Year>/
```
```
The.Dark.Knight.2008/
```

### Movie File
```
<Title>.<Year>.<Resolution>.<Codec>.<Source>.<Audio>.<Edition>.<Ext>
```
```
The.Dark.Knight.2008.1080p.x264.BluRay.DTS.mkv
The.Dark.Knight.2008.2160p.x265.BluRay.TrueHD.mkv
The.Dark.Knight.2008.1080p.x264.BluRay.DTS.Director's.Cut.mkv
```

> `Resolution`, `Codec`, `Source`, `Audio`, and `Edition` are all optional. Include what is relevant.

### External Subtitles
Subtitles must be placed **directly alongside** the video file (not in a subdirectory). The base filename must match the video file exactly.

```
<Title>.<Year>.<Resolution>.<Codec>.<Source>.<Audio>.<Lang>[.<Flags>].<Ext>
```
```
The.Dark.Knight.2008.1080p.x264.BluRay.DTS.en.srt
The.Dark.Knight.2008.1080p.x264.BluRay.DTS.en.sdh.srt
The.Dark.Knight.2008.1080p.x264.BluRay.DTS.default.en.forced.ass
```

| Flag | Meaning |
|------|---------|
| `default` | Default track |
| `forced` / `foreign` | Forced/foreign track |
| `sdh` / `cc` / `hi` | Hearing impaired |

> `hi` alone resolves as Hindi language. Pair it with another language code (e.g. `en.hi`) to tag as hearing impaired.

### Full Movie Directory Example
```
The.Dark.Knight.2008/
├── The.Dark.Knight.2008.1080p.x264.BluRay.DTS.mkv
├── The.Dark.Knight.2008.1080p.x264.BluRay.DTS.en.srt
├── The.Dark.Knight.2008.1080p.x264.BluRay.DTS.en.sdh.srt
├── behind the scenes/
│   └── Finding The Score.mp4
└── extras/
    └── Home Recreation.mp4
```

---

## TV Shows

### Series Directory
```
<Title>.<Year>/
```
```
Breaking.Bad.2008/
```

> Year = **first air year** of the series. Used for disambiguation only (e.g. two shows with the same name). Does not change between seasons.

### Season Directory
Season number is zero-padded to two digits. `Season 00` is reserved for specials.

```
S<NN>/
```
```
S00/   ← Specials
S01/
S02/
```

### Episode File
```
<Title>.<Year>.S<SS>E<EE>.<Resolution>.<Codec>.<Source>.<Audio>.<Ext>
```
```
Breaking.Bad.2008.S01E01.1080p.x264.BluRay.DTS.mkv
Breaking.Bad.2008.S01E01E02.1080p.x264.BluRay.DTS.mkv   ← Multi-episode
```

### External Subtitles
Same rule as movies — placed alongside the video file, base filename must match.
```
Breaking.Bad.2008.S01E01.1080p.x264.BluRay.DTS.en.srt
Breaking.Bad.2008.S01E01.1080p.x264.BluRay.DTS.en.sdh.srt
```

### Full Series Directory Example
```
Breaking.Bad.2008/
├── S00/
│   └── Breaking.Bad.2008.S00E01.mkv   ← Special
├── S01/
│   ├── Breaking.Bad.2008.S01E01.1080p.x264.BluRay.DTS.mkv
│   ├── Breaking.Bad.2008.S01E01.1080p.x264.BluRay.DTS.en.srt
│   ├── Breaking.Bad.2008.S01E02.1080p.x264.BluRay.DTS.mkv
│   ├── deleted scenes/
│   │   └── Deleted Scene Title.mp4
│   └── behind the scenes/
│       └── BTS Title.mp4
└── interviews/
    └── Interview With The Director.mp4
```

---

## Extras

Extras are placed in named subdirectories. Supported folder names:

| Folder | Type |
|--------|------|
| `behind the scenes` | BTS footage |
| `deleted scenes` | Deleted/cut scenes |
| `interviews` | Cast/crew interviews |
| `trailers` | Trailers |
| `featurettes` | Featurettes |
| `shorts` | Short films |
| `scenes` | Scenes |
| `samples` | Samples |
| `clips` | Clips |
| `extras` | Generic catch-all |
| `other` | Generic catch-all |
| `theme-music` | Theme audio (`theme.mp3` / `theme.flac`) |
| `backdrops` | Backdrop video |

Extras filenames use readable title casing with spaces (not dot-separated):
```
behind the scenes/
└── Finding The Score.mp4
```

---

## Reserved Characters

The following characters **must not** appear in any file or directory name:

```
< > : " / \ | ? *
```
