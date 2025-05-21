package useragent

import (
	"fmt"
	"math/rand"
	"strings"
)

type Language struct {
	Primary   string
	Secondary string
	Quality   float64
}

var (
	CommonLanguages = []string{
		"en", "ru", "es", "fr", "de", "it", "pt", "nl", "pl", "uk",
		"ja", "ko", "zh", "ar", "hi", "tr", "vi", "th", "id", "ms",
	}

	LanguageWeights = map[string]float64{
		"en": 0.9,
		"ru": 0.9,
		"es": 0.8,
		"fr": 0.8,
		"de": 0.8,
		"it": 0.7,
		"pt": 0.7,
		"nl": 0.7,
		"pl": 0.7,
		"uk": 0.7,
		"ja": 0.6,
		"ko": 0.6,
		"zh": 0.6,
		"ar": 0.5,
		"hi": 0.5,
		"tr": 0.5,
		"vi": 0.5,
		"th": 0.5,
		"id": 0.5,
		"ms": 0.5,
	}

	DefaultLanguages = []Language{
		{Primary: "en", Secondary: "ru", Quality: 0.9},
		{Primary: "ru", Secondary: "en", Quality: 0.9},
		{Primary: "en", Secondary: "es", Quality: 0.8},
		{Primary: "en", Secondary: "fr", Quality: 0.8},
		{Primary: "en", Secondary: "de", Quality: 0.8},
		{Primary: "ru", Secondary: "uk", Quality: 0.8},
		{Primary: "en", Secondary: "ja", Quality: 0.7},
		{Primary: "en", Secondary: "ko", Quality: 0.7},
		{Primary: "en", Secondary: "zh", Quality: 0.7},
	}
)

func GetRandomLanguage() Language {
	return DefaultLanguages[rand.Intn(len(DefaultLanguages))]
}

func GetLanguageString(lang Language) string {
	return strings.Join([]string{
		lang.Primary,
		lang.Secondary + ";q=" + formatQuality(lang.Quality),
	}, ",")
}

func GetRandomLanguageString() string {
	return GetLanguageString(GetRandomLanguage())
}

func GetCustomLanguageString(primary, secondary string, quality float64) string {
	return GetLanguageString(Language{
		Primary:   primary,
		Secondary: secondary,
		Quality:   quality,
	})
}

func formatQuality(q float64) string {
	q = float64(int(q*100)) / 100
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", q), "0"), ".")
}
