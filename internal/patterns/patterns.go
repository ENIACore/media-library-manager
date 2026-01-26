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

// Website/domain patterns commonly found at start of torrent names
var WebsitePatterns = []Pattern{
	// Include all TLDs for patterns with explicit prefixes
	`WWW\.[A-Z0-9]+\.(COM|NET|ORG|IO|CO|TV|ME|INFO|BIZ|US|UK|CA|DE|FR|JP|CN|RU|BR|AU|IN|IT|NL|ES|PL|SE|NO|FI|DK|BE|CH|AT|CZ|PT|GR|HU|RO|SK|BG|HR|SI|LT|LV|EE|IE|LU|MT|CY)`,
	`HTTP\.[A-Z0-9]+\.(COM|NET|ORG|IO|CO|TV|ME|INFO|BIZ|US|UK|CA|DE|FR|JP|CN|RU|BR|AU|IN|IT|NL|ES|PL|SE|NO|FI|DK|BE|CH|AT|CZ|PT|GR|HU|RO|SK|BG|HR|SI|LT|LV|EE|IE|LU|MT|CY)`,
	`HTTPS\.[A-Z0-9]+\.(COM|NET|ORG|IO|CO|TV|ME|INFO|BIZ|US|UK|CA|DE|FR|JP|CN|RU|BR|AU|IN|IT|NL|ES|PL|SE|NO|FI|DK|BE|CH|AT|CZ|PT|GR|HU|RO|SK|BG|HR|SI|LT|LV|EE|IE|LU|MT|CY)`,
	// Use only safe TLDs for patterns without explicit prefix
	`[A-Z0-9]+\.(COM|NET|ORG|IO|INFO|BIZ)`,
}


// Un-useful metadata of file pertaining to typical patterns found in torrent files
var MiscPatterns = []Pattern{
	// === UNUSED QUALITY INDICATORS ===
	// HDR variants
	`HDR`, `HDR10`, `HDR10PLUS`, `HDR10\+`, `DOLBY\.VISION`, `DOLBYVISION`, `DV`, `HLG`,
	// Bit depth
	`10BIT`, `10\.BIT`, `8BIT`, `8\.BIT`, `12BIT`, `12\.BIT`,
	// Color
	`SDR`,

	// === EDITION / VERSION ===
	`REMASTERED`, `REMASTER`,
	`EXTENDED`, `EXTENDED\.CUT`, `EXTENDED\.EDITION`,
	`UNRATED`,
	`UNCUT`,
	`DIRECTORS\.CUT`, `DC`,
	`THEATRICAL`, `THEATRICAL\.CUT`,
	`CRITERION`, `CC`,
	`SPECIAL\.EDITION`, `SE`,
	`ANNIVERSARY`, `ANNIVERSARY\.EDITION`,
	`COLLECTORS`, `COLLECTORS\.EDITION`, `CE`,
	`LIMITED`, `LIMITED\.EDITION`,
	`PROPER`,
	`REPACK`, `RERIP`,
	`REAL`,
	`RETAIL`,
	`FINAL\.CUT`,
	`IMAX`,
	`OPEN\.MATTE`, `OPENMATTE`,
	`3D`, `HSBS`, `HOU`, `HALF\.SBS`, `FULL\.SBS`,

	// === RELEASE INFO ===
	`INTERNAL`, `INT`,
	`NFO`, `NFOFIX`,
	`SAMPLE`,
	`PROOF`,
	`READNFO`, `READ\.NFO`,
	`DIRFIX`,
	`NFOFIX`,
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
	`CRITERION`,

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
}

var LanguagePatternGroups = []PatternGroup{
	{Key: `English`, Patterns: []Pattern{`ENGLISH`, `ENG`, `EN`}},
	{Key: `Spanish`, Patterns: []Pattern{`SPANISH`, `CASTELLANO`, `SPA`, `ES`, `ESPAÑOL`}},
	{Key: `French`, Patterns: []Pattern{`FRENCH`, `FRA`, `FR`}},
	{Key: `German`, Patterns: []Pattern{`GERMAN`, `DEUTSCH`, `GER`, `DE`, `GERMAN`}},
	{Key: `Italian`, Patterns: []Pattern{`ITALIAN`, `ITA`, `ITALIANO`}},
	{Key: `Portuguese`, Patterns: []Pattern{`PORTUGUESE`, `PORTUGUES`, `POR`, `PT`}},
	{Key: `Brazilian Portuguese`, Patterns: []Pattern{`BRAZILIAN`, `BRAZIL`, `BR`, `PORTUGUESE.BR`, `PT.BR`}},
	{Key: `Russian`, Patterns: []Pattern{`RUSSIAN`, `RUS`, `RU`}},
	{Key: `Japanese`, Patterns: []Pattern{`JAPANESE`, `JAP`, `JPN`, `JP`, `JA`}},
	{Key: `Korean`, Patterns: []Pattern{`KOREAN`, `KOR`, `KO`, `KR`}},
	{Key: `Arabic`, Patterns: []Pattern{`ARABIC`, `ARA`, `AR`}},
	{Key: `Hebrew`, Patterns: []Pattern{`HEBREW`, `HEB`, `HE`}},
	{Key: `Thai`, Patterns: []Pattern{`THAI`, `THA`, `TH`}},
	{Key: `Turkish`, Patterns: []Pattern{`TURKISH`, `TUR`, `TR`}},
	{Key: `Greek`, Patterns: []Pattern{`GREEK`, `GRE`, `EL`}},
	{Key: `Polish`, Patterns: []Pattern{`POLISH`, `POL`, `PL`, `POLSKI`}},
	{Key: `Hungarian`, Patterns: []Pattern{`HUNGARIAN`, `HUN`, `HU`, `MAGYAR`}},
	{Key: `Czech`, Patterns: []Pattern{`CZECH`, `CZE`, `CS`}},
	{Key: `Chinese`, Patterns: []Pattern{`CHINESE`, `CHI`, `ZH`}},
}

var BonusPatternGroups = []PatternGroup{
	{Key: `Behind.The.Scenes`, Patterns: []Pattern{
		`BEHIND\.THE\.SCENE[S]?`,
		`BTS`,
		`MAKING\.OF`,
		`MAKING`,
		`THE\.MAKING\.OF`,
	}},
	{Key: `Deleted.Scenes`, Patterns: []Pattern{
		`DELETED\.SCENE[S]?`,
		`DELETED`,
		`EXTENDED\.SCENE[S]?`,
		`ALTERNATE\.SCENE[S]?`,
		`ADDITIONAL\.SCENE[S]?`,
	}},
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
		`QA`,
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
		`DOC[S]?`,
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


var (
	GetLanguagePatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(LanguagePatternGroups)
	})
	GetBonusPatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(BonusPatternGroups)
	})
	GetMiscPatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(MiscPatterns)
	})
	GetWebsitePatterns = sync.OnceValue(func() []*CompiledPattern {
		return compilePatterns(WebsitePatterns)
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
