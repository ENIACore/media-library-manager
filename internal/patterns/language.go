package patterns

import (
	"sync"
	"regexp"
	lingua "github.com/pemistahl/lingua-go"
)

// linguaToKey maps lingua language constants to the exact Key values in LanguagePatternGroups.
// Languages lingua detects that have no matching key return "".
// Bokmal and Nynorsk both collapse to "Norwegian"; Tagalog maps to "Filipino".
var linguaToKey = map[lingua.Language]string {
	lingua.Arabic:     "Arabic",
	lingua.Basque:     "Basque",
	lingua.Catalan:    "Catalan",
	lingua.Chinese:    "Chinese",
	lingua.Croatian:   "Croatian",
	lingua.Czech:      "Czech",
	lingua.Danish:     "Danish",
	lingua.Dutch:      "Dutch",
	lingua.English:    "English",
	lingua.Finnish:    "Finnish",
	lingua.French:     "French",
	lingua.German:     "German",
	lingua.Greek:      "Greek",
	lingua.Hebrew:     "Hebrew",
	lingua.Hungarian:  "Hungarian",
	lingua.Indonesian: "Indonesian",
	lingua.Italian:    "Italian",
	lingua.Japanese:   "Japanese",
	lingua.Korean:     "Korean",
	lingua.Malay:      "Malay",
	lingua.Bokmal:     "Norwegian",
	lingua.Nynorsk:    "Norwegian",
	lingua.Polish:     "Polish",
	lingua.Portuguese: "Portuguese",
	lingua.Romanian:   "Romanian",
	lingua.Russian:    "Russian",
	lingua.Spanish:    "Spanish",
	lingua.Swedish:    "Swedish",
	lingua.Tagalog:    "Filipino",
	lingua.Thai:       "Thai",
	lingua.Turkish:    "Turkish",
	lingua.Ukrainian:  "Ukrainian",
	lingua.Vietnamese: "Vietnamese",
}

var LanguagePatternGroups = []PatternGroup{
	{Key: `English`, Patterns: []Pattern{`ENGLISH`, `ENG`, `EN`}},
	{Key: `Spanish`, Patterns: []Pattern{`SPANISH`, `CASTELLANO`, `SPA`, `ESPAÑOL`, `ESPANOL`}},
	{Key: `Latin-American-Spanish`, Patterns: []Pattern{`LATIN.AMERICAN`, `LATINOAMERICANO`, `LATAM`, }},
	{Key: `French`, Patterns: []Pattern{`FRENCH`, `FRANCAIS`, `FRANÇAIS`, `FRA`, `FR`, `FRE`}},
	{Key: `Canadian-French`, Patterns: []Pattern{`CANADIAN`, `QUEBEC`, `QC`}},
	{Key: `German`, Patterns: []Pattern{`GERMAN`, `DEUTSCH`, `GER`, `DE`, `DEU`}},
	{Key: `Italian`, Patterns: []Pattern{`ITALIAN`, `ITA`, `ITALIANO`}},
	{Key: `Portuguese`, Patterns: []Pattern{`PORTUGUESE`, `PORTUGUES`, `PORTUGUÊS`, `POR`, `PT`}},
	{Key: `Brazilian-Portuguese`, Patterns: []Pattern{`BRAZILIAN`, `BRAZIL`, `BR`, `PORTUGUESE.BR`, `PT.BR`, `POR.BR`}},
	{Key: `Russian`, Patterns: []Pattern{`RUSSIAN`, `RUS`, `RU`}},
	{Key: `Japanese`, Patterns: []Pattern{`JAPANESE`, `JAP`, `JPN`, `JP`, `JA`}},
	{Key: `Korean`, Patterns: []Pattern{`KOREAN`, `KOR`, `KO`, `KR`}},
	{Key: `Arabic`, Patterns: []Pattern{`ARABIC`, `ARA`, `AR`}},
	{Key: `Hebrew`, Patterns: []Pattern{`HEBREW`, `HEB`, }},
	{Key: `Thai`, Patterns: []Pattern{`THAI`, `THA`, `TH`}},
	{Key: `Turkish`, Patterns: []Pattern{`TURKISH`, `TUR`, `TR`}},
	{Key: `Greek`, Patterns: []Pattern{`GREEK`, `GRE`, `EL`, `GRC`}},
	{Key: `Polish`, Patterns: []Pattern{`POLISH`, `POL`, `PL`, `POLSKI`}},
	{Key: `Hungarian`, Patterns: []Pattern{`HUNGARIAN`, `HUN`, `HU`, `MAGYAR`}},
	{Key: `Czech`, Patterns: []Pattern{`CZECH`, `CZE`, `CS`, `CES`}},
	{Key: `Chinese`, Patterns: []Pattern{`CHINESE`, `CHI`, `ZH`, `CN`}},
	{Key: `Simplified-Chinese`, Patterns: []Pattern{`SIMPLIFIED`, `HANS`, `ZH.HANS`, `ZH.CN`, `CHS`}},
	{Key: `Traditional-Chinese`, Patterns: []Pattern{`TRADITIONAL`, `HANT`, `ZH.HANT`, `ZH.TW`, `CHT`}},
	{Key: `Dutch`, Patterns: []Pattern{`DUTCH`, `DUT`, `NL`, `NLD`, `NEDERLANDS`}},
	{Key: `Danish`, Patterns: []Pattern{`DANISH`, `DAN`, `DA`, `DANSK`}},
	{Key: `Finnish`, Patterns: []Pattern{`FINNISH`, `FIN`, `FI`, `SUOMI`}},
	{Key: `Swedish`, Patterns: []Pattern{`SWEDISH`, `SWE`, `SV`, `SVENSKA`}},
	{Key: `Norwegian`, Patterns: []Pattern{`NORWEGIAN`, `NOR`, `NOB`, `NORSK`}},
	{Key: `Croatian`, Patterns: []Pattern{`CROATIAN`, `HRV`, `HR`, `HRVATSKI`}},
	{Key: `Romanian`, Patterns: []Pattern{`ROMANIAN`, `RUM`, `RO`, `RON`, `ROMANA`}},
	{Key: `Ukrainian`, Patterns: []Pattern{`UKRAINIAN`, `UKR`, `UK`}},
	{Key: `Vietnamese`, Patterns: []Pattern{`VIETNAMESE`, `VIE`, `VI`}},
	{Key: `Indonesian`, Patterns: []Pattern{`INDONESIAN`, `IND`, `ID`, `BAHASA`}},
	{Key: `Malay`, Patterns: []Pattern{`MALAY`, `MAY`, `MS`, `MSA`}},
	{Key: `Filipino`, Patterns: []Pattern{`FILIPINO`, `FIL`, `TAGALOG`, `TL`}},
	{Key: `Basque`, Patterns: []Pattern{`BASQUE`, `BAQ`, `EU`, `EUS`, `EUSKARA`}},
	{Key: `Catalan`, Patterns: []Pattern{`CATALAN`, `CAT`, `CA`, `CATALA`}},
	{Key: `Galician`, Patterns: []Pattern{`GALICIAN`, `GLG`, `GL`, `GALEGO`}},
}

var (
	GetLanguagePatternGroups = sync.OnceValue(func() []CompiledPatternGroup {
		return compilePatternGroups(LanguagePatternGroups)
	})
	GetLinguaToKey = sync.OnceValue(func() map[lingua.Language]string {
        return linguaToKey
    })
	GetSrtTimestampPattern = sync.OnceValue(func() *CompiledPattern {
		return (*CompiledPattern)(regexp.MustCompile(`^\d{2}:\d{2}:\d{2}[,\.]\d{3}\s*-->\s*\d{2}:\d{2}:\d{2}[,\.]\d{3}`))	
	})
	GetSrtSequencePattern = sync.OnceValue(func() *CompiledPattern {
		return (*CompiledPattern)(regexp.MustCompile(`^\d+$`))	
	})
	GetAssDialoguePattern = sync.OnceValue(func() *CompiledPattern {
		return (*CompiledPattern)(regexp.MustCompile(`(?i)^Dialogue:`))	
	})
	GetAssOverridePattern = sync.OnceValue(func() *CompiledPattern {
		return (*CompiledPattern)(regexp.MustCompile(`\{[^}]*\}`))	
	})
	GetSdhBracketPattern = sync.OnceValue(func() *CompiledPattern {
		return (*CompiledPattern)(regexp.MustCompile(`\[[^\]]+\]`))	
	})
	GetSdhMusicPattern = sync.OnceValue(func() *CompiledPattern {
		return (*CompiledPattern)(regexp.MustCompile(`[♪♫🎵🎶]`))	
	})
)
