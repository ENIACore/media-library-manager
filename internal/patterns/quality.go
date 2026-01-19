package patterns

import (
	"sync"
)

var ResolutionPatternGroups = []PatternGroup{
	{Key: `8K`, Patterns: []Pattern{`8K`, `4320[PI]?`, `7680X4320`, `FULLUHD`}},
	{Key: `4K`, Patterns: []Pattern{`4K`, `UHD`, `2160[PI]?`, `3840X2160`}},
	{Key: `2K`, Patterns: []Pattern{`2K`, `1440[PI]?`, `2560X1440`, `QHD`, `WQHD`}},
	{Key: `1080p`, Patterns: []Pattern{`1080[PI]?`, `FHD`, `1920X1080`, `FULLHD`}},
	{Key: `720p`, Patterns: []Pattern{`720[PI]?`, `1280X720`}},
	{Key: `576p`, Patterns: []Pattern{`576[PI]?`, `PAL`}},
	{Key: `480p`, Patterns: []Pattern{`480[PI]?`, `NTSC`}},
	{Key: `360p`, Patterns: []Pattern{`360[PI]?`}},
	{Key: `240p`, Patterns: []Pattern{`240[PI]?`}},
}

var CodecPatternGroups = []PatternGroup{
	{Key: `AV1`, Patterns: []Pattern{`AV1`, `SVT\.AV1`, `SVTAV1`, `AOV1`}},
	{Key: `VP9`, Patterns: []Pattern{`VP9`}},
	{Key: `VP8`, Patterns: []Pattern{`VP8`}},
	{Key: `x265`, Patterns: []Pattern{`X265`, `X\.265`, `H265`, `H\.265`, `HEVC`, `HEVC10`, `HEVC10BIT`, `H265P`}},
	{Key: `x264`, Patterns: []Pattern{`X264`, `X\.264`, `H264`, `H\.264`, `AVC`, `AVC1`, `H264P`}},
	{Key: `x263`, Patterns: []Pattern{`X263`, `X\.263`, `H263`, `H\.263`}},
	{Key: `XviD`, Patterns: []Pattern{`XVID`, `XVID\.AF`}},
	{Key: `DivX`, Patterns: []Pattern{`DIVX`, `DIV3`, `DIVX6`}},
	{Key: `MPEG-4`, Patterns: []Pattern{`MPEG\.4`, `MPEG4`, `MP4V`}},
	{Key: `MPEG-2`, Patterns: []Pattern{`MPEG\.2`, `MPEG2`, `MP2V`}},
	{Key: `MPEG-1`, Patterns: []Pattern{`MPEG\.1`, `MPEG1`, `MP1V`}},
	{Key: `VC-1`, Patterns: []Pattern{`VC\.1`, `VC1`, `WMV3`, `WVC1`}},
	{Key: `Theora`, Patterns: []Pattern{`THEORA`}},
	{Key: `ProRes`, Patterns: []Pattern{`PRORES`, `PRORES422`, `PRORES4444`, `PRORES422HQ`}},
	{Key: `DNxHD`, Patterns: []Pattern{`DNXHD`, `DNXHR`}},
}

var SourcePatternGroups = []PatternGroup{
	{Key: `Remux`, Patterns: []Pattern{`REMUX`}},
	{Key: `BluRay`, Patterns: []Pattern{`BLURAY`, `BDRIP`, `BD\.RIP`, `BR\.RIP`, `BRRIP`, `BDMV`, `BDISO`, `BD25`, `BD50`, `BD66`, `BD100`}},
	{Key: `WEB-DL`, Patterns: []Pattern{`WEB\.DL`, `WEBDL`}},
	{Key: `WEBRip`, Patterns: []Pattern{`WEBRIP`, `WEB-RIP`, `WEB\.RIP`}},
	{Key: `WEB`, Patterns: []Pattern{`WEB`}},
	{Key: `HDRip`, Patterns: []Pattern{`HDRIP`, `HD\.RIP`}},
	{Key: `DVDRip`, Patterns: []Pattern{`DVDRIP`, `DVD\.RIP`}},
	{Key: `DVD`, Patterns: []Pattern{`DVD`, `DVDSCR`, `DVD5`, `DVD9`}},
	{Key: `HDTV`, Patterns: []Pattern{`HDTV`, `HDTVRIP`, `DTTV`, `PDTV`, `SDTV`, `LDTV`}},
	{Key: `Telecine`, Patterns: []Pattern{`TELECINE`, `TC`}},
	{Key: `Telesync`, Patterns: []Pattern{`TELESYNC`, `TS`}},
	{Key: `Screener`, Patterns: []Pattern{`SCREENER`, `SCR`, `DVDSCR`, `BDSCR`}},
	{Key: `CAM`, Patterns: []Pattern{`CAMRIP`, `CAM`, `HDCAM`}},
	{Key: `Workprint`, Patterns: []Pattern{`WORKPRINT`, `WP`}},
	{Key: `PPV`, Patterns: []Pattern{`PPV`, `PPVRIP`}},
	{Key: `VODRip`, Patterns: []Pattern{`VODRIP`, `VOD`}},
	{Key: `HC`, Patterns: []Pattern{`HC`, `HCHDCAM`}},
	{Key: `Line`, Patterns: []Pattern{`LINE`}},
	{Key: `HDTS`, Patterns: []Pattern{`HDTS`, `HD\.TS`}},
	{Key: `HDTC`, Patterns: []Pattern{`HDTC`, `HD\.TC`}},
	{Key: `TVRip`, Patterns: []Pattern{`TVRIP`, `SATRIP`, `DTTVRIP`}},
}

var AudioPatternGroups = []PatternGroup{
	{Key: `Atmos`, Patterns: []Pattern{`ATMOS`, `DOLBY-ATMOS`, `DOLBY\.ATMOS`, `DOLBYATMOS`}},
	{Key: `DTS-X`, Patterns: []Pattern{`DTSX`, `DTS\.X`, `DTS`}},
	{Key: `DTS-HD MA`, Patterns: []Pattern{`DTS\.HD\.MA`, `DTSHD-MA`, `DTSHD\.MA`, `DTS\.HD`, `DTSHD`}},
	{Key: `DTS-MA`, Patterns: []Pattern{`DTS\.MA`, `DTSMA`}},
	{Key: `DTS-ES`, Patterns: []Pattern{`DTS\.ES`, `DTSES`}},
	{Key: `DTS`, Patterns: []Pattern{`DTS`}},
	{Key: `TrueHD`, Patterns: []Pattern{`TRUEHD`, `TRUE\.HD`}},
	{Key: `DD+`, Patterns: []Pattern{`DDP`, `E\.AC\.3`, `EAC3`, `DD\.PLUS`, `DDPLUS`}},
	{Key: `DD`, Patterns: []Pattern{`DD`, `AC3`, `DOLBY-DIGITAL`, `DOLBY\.DIGITAL`, `DOLBYDIGITAL`}},
	{Key: `AAC`, Patterns: []Pattern{`AAC`, `HE\.AAC`, `HEAAC`}},
	{Key: `FLAC`, Patterns: []Pattern{`FLAC`}},
	{Key: `MP3`, Patterns: []Pattern{`MP3`}},
	{Key: `LPCM`, Patterns: []Pattern{`LPCM`, `PCM`}},
	{Key: `Ogg`, Patterns: []Pattern{`OGG`, `VORBIS`}},
	{Key: `Opus`, Patterns: []Pattern{`OPUS`}},
	{Key: `5.1`, Patterns: []Pattern{`5\.1`, `51`, `6CH`}},
	{Key: `7.1`, Patterns: []Pattern{`7\.1`, `71`, `8CH`}},
	{Key: `2.0`, Patterns: []Pattern{`2\.0`, `20`, `STEREO`, `2CH`}},
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
		res := make([]CompiledPatternGroup, len(ResolutionPatternGroups) + len(CodecPatternGroups) + len(SourcePatternGroups) + len(AudioPatternGroups))

		for i, group := range GetResolutionPatternGroups() {
			res[i] = CompiledPatternGroup{
				Key: group.Key,
				Patterns: group.Patterns,
			}
		}
		for i, group := range GetCodecPatternGroups() {
			res[i] = CompiledPatternGroup{
				Key: group.Key,
				Patterns: group.Patterns,
			}
		}
		for i, group := range GetSourcePatternGroups() {
			res[i] = CompiledPatternGroup{
				Key: group.Key,
				Patterns: group.Patterns,
			}
		}
		for i, group := range GetAudioPatternGroups() {
			res[i] = CompiledPatternGroup{
				Key: group.Key,
				Patterns: group.Patterns,
			}
		}

		return res
	})
)
