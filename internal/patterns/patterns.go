package patterns

import (
	"regexp"
	"sync"
)

type Pattern string

type CompiledPattern regexp.Regexp

type PatternGroup struct {
	Key      string
	Patterns []Pattern
}

type CompiledPatternGroup struct {
	Key      string
	Patterns []*CompiledPattern
}

var ReleaseWebsites = []Pattern{
		// YTS variants
	`YTS\.MX`,
	`YTS\.AM`,
	`YTS\.LT`,
	`YTS\.AG`,
	// PAHE variants
	`PAHE\.IN`,
	`PAHE\.PH`,
	// 1337x
	`1337X\.TO`,
	`WWW\.1337X\.TO`,
	// RARBG
	`RARBG\.TO`,
	`RARBG\.IS`,
	// The Pirate Bay
	`THEPIRATEBAY\.ORG`,
	`TPB\.AM`,
	// Nyaa
	`NYAA\.SI`,
	`NYAA\.NET`,
	// TorrentGalaxy
	`TORRENTGALAXY\.TO`,
	`TGX\.RS`,
	// Kickass Torrents
	`KICKASSTORRENTS\.TO`,
	`KAT\.AM`,
	`KATCR\.CO`,
	// EZTV
	`EZTV\.RE`,
	`EZTV\.AG`,
	`EZTV\.IO`,
	// Torrentz
	`TORRENTZ2\.EU`,
	`TORRENTZ\.EU`,
	// LimeTorrents
	`LIMETORRENTS\.INFO`,
	`LIMETORRENTS\.CO`,
	// Zooqle
	`ZOOQLE\.COM`,
	// Torlock
	`TORLOCK\.COM`,
	// TorrentDownloads
	`TORRENTDOWNLOADS\.ME`,
	// Torrenting
	`WWW\.TORRENTING\.COM`,
	`TORRENTING\.COM`,
	// TorrentCounter
	`TORRENTCOUNTER\.TO`,
	`TORRENTCOUNTER\.BE`,
	// ExtraTorrent
	`EXTRATORRENTS\.CC`,
	`EXTRATORRENT\.AG`,
	// IsoHunt
	`ISOHUNT\.TO`,
	// Monova
	`MONOVA\.ORG`,
	// Seedpeer
	`SEEDPEER\.EU`,
	// TorrentFunk
	`TORRENTFUNK\.COM`,
	// GloTorrents
	`GLOTORRENTS\.COM`,
	// BitSnoop
	`BITSNOOP\.COM`,
	// SkyTorrents
	`SKYTORRENTS\.IN`,
	// Idope
	`IDOPE\.SE`,
	// AniDex
	`ANIDEX\.INFO`,
	// AniRena
	`ANIRENA\.COM`,
	// Rutor
	`RUTOR\.INFO`,
	`RUTOR\.ORG`,
	// Rutracker
	`RUTRACKER\.ORG`,
	// Nnm-Club
	`NNM-CLUB\.ME`,
	`NNM-CLUB\.RU`,

	`TORRENT911.CC`,
}

// Website/domain patterns commonly found at start of torrent names
var WebsitePatterns = append([]Pattern{
	// Include all TLDs for patterns with explicit prefixes
	`WWW\.[A-Z0-9]+\.(COM|NET|ORG|IO|CO|TV|ME|INFO|BIZ|US|UK|CA|DE|FR|JP|CN|RU|BR|AU|IN|IT|NL|ES|PL|SE|NO|FI|DK|BE|CH|AT|CZ|PT|GR|HU|RO|SK|BG|HR|SI|LT|LV|EE|IE|LU|MT|CY)`,
	`HTTP\.[A-Z0-9]+\.(COM|NET|ORG|IO|CO|TV|ME|INFO|BIZ|US|UK|CA|DE|FR|JP|CN|RU|BR|AU|IN|IT|NL|ES|PL|SE|NO|FI|DK|BE|CH|AT|CZ|PT|GR|HU|RO|SK|BG|HR|SI|LT|LV|EE|IE|LU|MT|CY)`,
	`HTTPS\.[A-Z0-9]+\.(COM|NET|ORG|IO|CO|TV|ME|INFO|BIZ|US|UK|CA|DE|FR|JP|CN|RU|BR|AU|IN|IT|NL|ES|PL|SE|NO|FI|DK|BE|CH|AT|CZ|PT|GR|HU|RO|SK|BG|HR|SI|LT|LV|EE|IE|LU|MT|CY)`,
	// Use only safe TLDs for patterns without explicit prefix
	`[A-Z0-9]+\.(COM|NET|ORG|IO|INFO|BIZ)`,
}, ReleaseWebsites...)

var TMDBPatterns = []Pattern{
	`TMDB\.(\d+)`,
}

var IMDBPatterns = []Pattern{
	`TT(\d{7,})`,
	`IMDB\.(TT\d+)`,
}

var TVDBPatterns = []Pattern{
	`TVDB\.(\d+)`,
}

// IdentificationPatterns are db identifier tokens used as terminators.
var IdentificationPatterns = append(append(append([]Pattern{}, TMDBPatterns...), IMDBPatterns...), TVDBPatterns...)

// Patterns that indicate directory or file is a sample and should be ignored
var SamplePatterns = []Pattern{
	`SAMPLE[S]?`,
	`PREVIEW[S]?`,
}

var JunkPatterns = []Pattern{
	`ETRG`,
}

// MiscPatterns are noise tokens in torrent names with no useful metadata value.
var MiscPatterns = append([]Pattern{
	`NEW`, `ESUB`,

	// === UNUSED QUALITY INDICATORS ===
	// HDR variants
	`HDR`, `HDR10`, `HDR10PLUS`, `HDR10\+`, `DOLBY\.VISION`, `DOLBYVISION`, `DV`, `HLG`,
	// Bit depth
	`10BIT`, `10\.BIT`, `8BIT`, `8\.BIT`, `12BIT`, `12\.BIT`,
	// Color
	`SDR`,

	// === RELEASE FIX FLAGS ===
	`PROPER`,
	`REPACK`, `RERIP`,
	`REAL`,
	`RETAIL`,
	`REISSUE`,

	// === RELEASE INFO ===
	`INTERNAL`, `INT`,
	`NFO`, `NFOFIX`,
	`SAMPLE`,
	`PROOF`,
	`READNFO`, `READ\.NFO`,
	`DIRFIX`,
	`SYNCFIX`,
	`SAMPLEFIX`,
	`SUBBED`, `SUBS`, `SUB`,
	`DUBBED`, `DUB`,
	`HARDCODED`, `HC`,
	`MULTISUBS`, `MULTI\.SUBS`, `MULTISUB`,
	`MULTI`, `MULTILANG`, `MULTI\.LANG`, `MULTi`,

	// === SCENE / P2P RELEASE GROUPS ===
	`YIFY`, `YTS`, `YTS\.MX`, `YTS\.AM`, `YTS\.LT`, `YTS\.AG`,
	`RARBG`,
	`ETRG`, `ETTV`, `ETHD`,
	`PSA`, `PSARIPS`,
	`GALAXYRG`, `GALAXY\.RG`, `GALAXYTV`,
	`SPARKS`,
	`GECKOS`,
	`AMIABLE`,
	`DRONES`,
	`FGT`,
	`EVO`,
	`CMRG`,
	`TIGOLE`, `QXRTRIGOLE`,
	`FLUX`,
	`NTG`,
	`EPSILON`,
	`PLAYNOW`,
	`HDETG`,
	`DIMENSION`,
	`LOL`,
	`KILLERS`,
	`AVS`,
	`SVA`,
	`FLEET`,
	`NTEB`,
	`PAHE`, `PAHE\.IN`, `PAHE\.PH`,
	`MKVKING`,
	`ION10`,
	`AMZN`,
	`NF`,
	`HULU`,
	`DSNP`,
	`ATVP`,
	`PCOK`,
	`HMAX`, `HBO`,
	`MAX`,
	`PMTP`,
	`CRAV`,
	`SHO`,
	`STAN`,

	// === TV SPECIFIC ===
	`COMPLETE`, `COMPLETE\.SERIES`,
	`MINISERIES`, `MINI\.SERIES`,
	`PILOT`,
	`FINALE`,
	`HDTV`,

	// === MISC NOISE ===
	`XXX`,
	`HINDI\.DUBBED`, `TAMIL\.DUBBED`, `TELUGU\.DUBBED`,
	`CONVERT`,
	`COLORIZED`,
	`RESTORED`,
	`AI\.UPSCALE`, `UPSCALED`, `AI\.ENHANCED`,

	// === STUDIOS ===
	`PARAMOUNT`,
	`AMAZON`,
	`AMZN`,
}, ReleaseWebsites...)

var BehindTheScenesPatternGroups = []PatternGroup{
	{Key: `Behind.The.Scenes`, Patterns: []Pattern{
		`BEHIND\.THE\.SCENE[S]?`,
		`BTS`,
		`MAKING\.OF`,
		`THE\.MAKING\.OF`,
	}},
}

var DeletedScenesPatternGroups = []PatternGroup{
	{Key: `Deleted.Scenes`, Patterns: []Pattern{
		`DELETED\.SCENE[S]?`,
		`DELETED`,
		`EXTENDED\.SCENE[S]?`,
		`ALTERNATE\.SCENE[S]?`,
		`ALTERNATE\.ENDING[S]?`,
		`ADDITIONAL\.SCENE[S]?`,
	}},
}

var BonusPatternGroups = []PatternGroup{
	{Key: `Featurette`, Patterns: []Pattern{
		`FEATURETTE[S]?`,
		`FEATURE[S]?`,
		`SHORT[S]?`,
		`MINI\.FEATURE[S]?`,
	}},
	{Key: `Interview`, Patterns: []Pattern{
		`INTERVIEW[S]?`,
		`CAST\.INTERVIEW[S]?`,
		`Q\.?AND\.?A`,
		//`QA`,
	}},
	{Key: `Blooper`, Patterns: []Pattern{
		`BLOOPER[S]?`,
		`GAG\.?REEL[S]?`,
		`OUTTAKE[S]?`,
	}},
	{Key: `Trailer`, Patterns: []Pattern{
		`TRAILER[S]?`,
		`TEASER[S]?`,
		`PROMO[S]?`,
		`TV\.SPOT[S]?`,
		`TVSPOT[S]?`,
	}},
	{Key: `Commentary`, Patterns: []Pattern{
		`COMMENTARY`,
		`AUDIO\.COMMENTARY`,
		`DIRECTOR[S]?\.COMMENTARY`,
	}},
	{Key: `Documentary`, Patterns: []Pattern{
		`DOCUMENTARY`,
		`DOCUMENTARIES`,
		//`DOC[S]?`,
	}},
	{Key: `Extra`, Patterns: []Pattern{
		`EXTRA[S]?`,
		`BONUS`,
		`BONUS\.CONTENT`,
		`BONUS\.MATERIAL[S]?`,
		`SPECIAL\.FEATURE[S]?`,
		`SUPPLEMENTAL[S]?`,
		`SUPPLEMENT[S]?`,
	}},
}

var EditionPatternGroups = []PatternGroup{
	{Key: `Remastered`, Patterns: []Pattern{`REMASTERED`, `REMASTER`}},
	{Key: `Extended`, Patterns: []Pattern{`EXTENDED`, `EXTENDED\.CUT`, `EXTENDED\.EDITION`}},
	{Key: `Unrated`, Patterns: []Pattern{`UNRATED`}},
	{Key: `Uncut`, Patterns: []Pattern{`UNCUT`}},
	{Key: `DirectorsCut`, Patterns: []Pattern{`DIRECTOR[S]?\.CUT[S]?`, `DC`}},
	{Key: `Theatrical`, Patterns: []Pattern{`THEATRICAL`, `THEATRICAL\.CUT`}},
	{Key: `Criterion`, Patterns: []Pattern{`CRITERION`}},
	{Key: `SpecialEdition`, Patterns: []Pattern{`SPECIAL\.EDITION`}},
	{Key: `Anniversary`, Patterns: []Pattern{`ANNIVERSARY`, `ANNIVERSARY\.EDITION`}},
	{Key: `Collectors`, Patterns: []Pattern{`COLLECTORS`, `COLLECTORS\.EDITION`}},
	{Key: `Limited`, Patterns: []Pattern{`LIMITED`, `LIMITED\.EDITION`}},
	{Key: `FinalCut`, Patterns: []Pattern{`FINAL\.CUT`}},
	{Key: `IMAX`, Patterns: []Pattern{`IMAX`}},
	{Key: `OpenMatte`, Patterns: []Pattern{`OPEN\.MATTE`, `OPENMATTE`}},
	{Key: `3D`, Patterns: []Pattern{`3D`, `HSBS`, `HOU`, `HALF\.SBS`, `FULL\.SBS`}},
}

// SubtitleModifierPatternGroups matches subtitle modifier tokens in filename segments.
// Forced is listed first so it appears before SDH in the Language array when both are present.
var SubtitleModifierPatternGroups = []PatternGroup{
	{Key: "Forced", Patterns: []Pattern{`FORCED`}},
	{Key: "SDH", Patterns: []Pattern{`SDH`}},
}

var (
	GetBehindTheScenesPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
        return compilePatternGroups(BehindTheScenesPatternGroups)
    })
    GetDeletedScenesPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
        return compilePatternGroups(DeletedScenesPatternGroups)
    })
    GetBonusPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
        return compilePatternGroups(BonusPatternGroups)
    })
	GetEditionPatterns = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(EditionPatternGroups)
	})
	GetSamplePatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(SamplePatterns)
	})
	GetJunkPatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(JunkPatterns)
	})
	GetMiscPatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(MiscPatterns)
	})
	GetWebsitePatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(WebsitePatterns)
	})
	GetIdentificationPatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(IdentificationPatterns)
	})
	GetTMDBPatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(TMDBPatterns)
	})
	GetSubtitleModifierPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(SubtitleModifierPatternGroups)
	})
)

func compilePatternGroups(patternGroups []PatternGroup) []CompiledPatternGroup {
	res := make([]CompiledPatternGroup, len(patternGroups))
	for i, group := range patternGroups {
		patterns := make([]*CompiledPattern, len(group.Patterns))
		for j, pattern := range group.Patterns {
			patterns[j] = (*CompiledPattern)(regexp.MustCompile(string(pattern)))
		}
		res[i] = CompiledPatternGroup{
			Key:      group.Key,
			Patterns: patterns,
		}
	}
	return res
}

func compilePatterns(patterns []Pattern) []*CompiledPattern {
	result := make([]*CompiledPattern, len(patterns))
	for i, pattern := range patterns {
		result[i] = (*CompiledPattern)(regexp.MustCompile(string(pattern)))
	}
	return result
}
