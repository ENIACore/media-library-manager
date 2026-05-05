package patterns

import (
	"sync"
)

var ResolutionPatternGroups = []PatternGroup{
	// Ordered highest to lowest to avoid e.g. "4K" matching inside "4320"
	{Key: `8K`, Patterns: []Pattern{`8K`, `4320[PI]?`, `7680X4320`, `FULLUHD`}},
	{Key: `4K`, Patterns: []Pattern{`4K`, `UHD`, `2160[PI]?`, `3840X2160`}},
	{Key: `2K`, Patterns: []Pattern{`2K`, `1440[PI]?`, `2560X1440`, `QHD`, `WQHD`}},
	{Key: `1080p`, Patterns: []Pattern{`1080[PI]?`, `FHD`, `1920X1080`, `FULLHD`}},
	{Key: `720p`, Patterns: []Pattern{`720[PI]?`, `1280X720`, `HD`}},
	{Key: `576p`, Patterns: []Pattern{`576[PI]?`, `PAL`}},
	{Key: `480p`, Patterns: []Pattern{`480[PI]?`, `NTSC`, `SD`}},
	{Key: `360p`, Patterns: []Pattern{`360[PI]?`}},
	{Key: `240p`, Patterns: []Pattern{`240[PI]?`}},
}

var CodecPatternGroups = []PatternGroup{
	// AV1 before VP9/VP8 (no overlap but kept grouped by family)
	{Key: `AV1`, Patterns: []Pattern{`AV1`, `SVT\.AV1`, `SVTAV1`, `AOV1`}},
	{Key: `VP9`, Patterns: []Pattern{`VP9`}},
	{Key: `VP8`, Patterns: []Pattern{`VP8`}},
	// x265/HEVC variants before x264/AVC — HEVC10 before HEVC so it doesn't partially match
	{Key: `x265`, Patterns: []Pattern{`HEVC10BIT`, `HEVC10`, `X265`, `X\.265`, `H265`, `H\.265`, `HEVC`, `H265P`}},
	{Key: `x264`, Patterns: []Pattern{`X264`, `X\.264`, `H264`, `H\.264`, `AVC1`, `AVC`, `H264P`}},
	{Key: `x263`, Patterns: []Pattern{`X263`, `X\.263`, `H263`, `H\.263`}},
	{Key: `XviD`, Patterns: []Pattern{`XVID\.AF`, `XVID`}},
	{Key: `DivX`, Patterns: []Pattern{`DIVX6`, `DIVX`, `DIV3`}},
	// MPEG variants: specific before generic (MPEG.4 before MPEG4 before MP4V)
	{Key: `MPEG-4`, Patterns: []Pattern{`MPEG\.4`, `MPEG4`, `MP4V`}},
	{Key: `MPEG-2`, Patterns: []Pattern{`MPEG\.2`, `MPEG2`, `MP2V`}},
	{Key: `MPEG-1`, Patterns: []Pattern{`MPEG\.1`, `MPEG1`, `MP1V`}},
	{Key: `VC-1`, Patterns: []Pattern{`VC\.1`, `VC1`, `WMV3`, `WVC1`}},
	{Key: `Theora`, Patterns: []Pattern{`THEORA`}},
	// ProRes sub-variants before base ProRes
	{Key: `ProRes`, Patterns: []Pattern{`PRORES4444`, `PRORES422HQ`, `PRORES422`, `PRORES`}},
	{Key: `DNxHD`, Patterns: []Pattern{`DNXHR`, `DNXHD`}},
}

var SourcePatternGroups = []PatternGroup{
	// Remux is highest quality — first
	{Key: `Remux`, Patterns: []Pattern{`REMUX`}},
	// Blu-ray disc formats — raw disc images before rips
	{Key: `BluRay`, Patterns: []Pattern{`BLURAY`, `BDMV`, `BDISO`, `BD25`, `BD50`, `BD66`, `BD100`, `BD`, `BLU\.RAY`}},
	// BDRip: encode direct from BD disc (better than BRRip)
	{Key: `BDRip`, Patterns: []Pattern{`BDRIP`, `BD\.RIP`}},
	// BRRip: re-encode of an existing BDRip (lower quality than BDRip)
	{Key: `BRRip`, Patterns: []Pattern{`BRRIP`, `BR\.RIP`}},
	// WEB sources — WEB-DL and WEBRip before bare WEB to avoid shadowing
	{Key: `WEB-DL`, Patterns: []Pattern{`WEB\.DL`, `WEBDL`}},
	{Key: `WEBRip`, Patterns: []Pattern{`WEBRIP`, `WEB-RIP`, `WEB\.RIP`}},
	{Key: `WEB`, Patterns: []Pattern{`WEB`}},
	{Key: `HDRip`, Patterns: []Pattern{`HDRIP`, `HD\.RIP`}},
	// DVD: DVDScr and DVDRip before bare DVD patterns
	{Key: `DVDScr`, Patterns: []Pattern{`DVDSCR`}},
	{Key: `DVDRip`, Patterns: []Pattern{`DVDRIP`, `DVD\.RIP`}},
	{Key: `DVD`, Patterns: []Pattern{`DVD5`, `DVD9`, `DVD`}},
	// TV sources — HDTV before the others, then specific variants
	{Key: `HDTV`, Patterns: []Pattern{`HDTVRIP`, `HDTV`}},
	{Key: `PDTV`, Patterns: []Pattern{`PDTV`}},
	{Key: `SDTV`, Patterns: []Pattern{`SDTV`}},
	{Key: `LDTV`, Patterns: []Pattern{`LDTV`}},
	{Key: `DTTV`, Patterns: []Pattern{`DTTV`}},
	{Key: `TVRip`, Patterns: []Pattern{`DTTVRIP`, `TVRIP`}},
	{Key: `SATRip`, Patterns: []Pattern{`SATRIP`}},
	// PPV / VOD
	{Key: `PPV`, Patterns: []Pattern{`PPVRIP`, `PPV`}},
	{Key: `VODRip`, Patterns: []Pattern{`VODRIP`, `VOD`}},
	// Cam sources — HDCAM before CAM so it doesn't get swallowed
	{Key: `HDCAM`, Patterns: []Pattern{`HDCAM`}},
	{Key: `CAM`, Patterns: []Pattern{`CAMRIP`, `CAM`}},
	// Screeners — medium-specific before generic
	{Key: `BDScr`, Patterns: []Pattern{`BDSCR`}},
	{Key: `DVDScr`, Patterns: []Pattern{`DVDSCR`}}, // already listed above but included for Screener hierarchy clarity
	{Key: `Screener`, Patterns: []Pattern{`SCREENER`, `SCR`}},
	// HDTS/HDTC before TS/TC so bare TS/TC don't shadow them
	{Key: `HDTS`, Patterns: []Pattern{`HDTS`, `HD\.TS`}},
	{Key: `Telesync`, Patterns: []Pattern{`TELESYNC`, `TS`}},
	{Key: `HDTC`, Patterns: []Pattern{`HDTC`, `HD\.TC`}},
	{Key: `Telecine`, Patterns: []Pattern{`TELECINE`, `TC`}},
	// HCHDCAM before HC so HC doesn't shadow it
	{Key: `HCHDCAM`, Patterns: []Pattern{`HCHDCAM`}},
	{Key: `HC`, Patterns: []Pattern{`HC`}},
	{Key: `Workprint`, Patterns: []Pattern{`WORKPRINT`, `WP`}},
	//{Key: `Line`, Patterns: []Pattern{`LINE`}},
}

var AudioPatternGroups = []PatternGroup{
	// Atmos sits on top of TrueHD in practice but is a separate tag
	{Key: `Atmos`, Patterns: []Pattern{`DOLBY\.ATMOS`, `DOLBY-ATMOS`, `DOLBYATMOS`, `ATMOS`}},
	// DTS family: most specific sub-formats first, bare DTS last
	{Key: `DTS-HD MA`, Patterns: []Pattern{`DTS\.HD\.MA`, `DTSHD-MA`, `DTSHD\.MA`, `DTS\.HD`, `DTSHD`}},
	{Key: `DTS-MA`, Patterns: []Pattern{`DTS\.MA`, `DTSMA`}},
	{Key: `DTS-ES`, Patterns: []Pattern{`DTS\.ES`, `DTSES`}},
	{Key: `DTS-X`, Patterns: []Pattern{`DTS\.X`, `DTSX`}},
	{Key: `DTS`, Patterns: []Pattern{`DTS`}},
	// TrueHD
	{Key: `TrueHD`, Patterns: []Pattern{`TRUEHD`, `TRUE\.HD`}},
	// Dolby Digital family: DD+ before DD so DD doesn't shadow it
	{Key: `DD+`, Patterns: []Pattern{`E\.AC\.3`, `EAC3`, `DD\.PLUS`, `DDPLUS`, `DDP`}},
	{Key: `DD`, Patterns: []Pattern{`DOLBY-DIGITAL`, `DOLBY\.DIGITAL`, `DOLBYDIGITAL`, `AC3`, `DD`}},
	// AAC variants before bare AAC
	{Key: `AAC`, Patterns: []Pattern{`HE\.AAC`, `HEAAC`, `AAC`}},
	{Key: `FLAC`, Patterns: []Pattern{`FLAC`}},
	{Key: `MP3`, Patterns: []Pattern{`MP3`}},
	{Key: `LPCM`, Patterns: []Pattern{`LPCM`, `PCM`}},
	{Key: `Ogg`, Patterns: []Pattern{`VORBIS`, `OGG`}},
	{Key: `Opus`, Patterns: []Pattern{`OPUS`}},
	// Channel layouts — 7.1 before 5.1 before 2.0 so partial matches don't fire early
	{Key: `7.1`, Patterns: []Pattern{`7\.1`, `71`, `8CH`}},
	{Key: `5.1`, Patterns: []Pattern{`5\.1`, `51`, `6CH`}},
	{Key: `2.0`, Patterns: []Pattern{`2\.0`, `STEREO`, `2CH`, `20`}},
	{Key: `Dual`, Patterns: []Pattern{`DUAL\.AUDIO`, `DUAL`}},
}

var (
	GetResolutionPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(ResolutionPatternGroups)
	})
	GetCodecPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(CodecPatternGroups)
	})
	GetSourcePatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(SourcePatternGroups)
	})
	GetAudioPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(AudioPatternGroups)
	})
	GetQualityPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		res := make([]CompiledPatternGroup, len(ResolutionPatternGroups)+len(CodecPatternGroups)+len(SourcePatternGroups)+len(AudioPatternGroups))

		offset := 0
		for _, groups := range [][]CompiledPatternGroup{
			GetResolutionPatternGroups(),
			GetCodecPatternGroups(),
			GetSourcePatternGroups(),
			GetAudioPatternGroups(),
		} {
			for _, group := range groups {
				res[offset] = CompiledPatternGroup{Key: group.Key, Patterns: group.Patterns}
				offset++
			}
		}

		return res
	})
)
