package useragent

import (
	"strings"
	"testing"
)

func TestGetRandomLanguage(t *testing.T) {
	lang := GetRandomLanguage()
	if lang.Primary == "" || lang.Secondary == "" {
		t.Error("GetRandomLanguage returned empty language")
	}
	if lang.Quality <= 0 || lang.Quality > 1 {
		t.Error("GetRandomLanguage returned invalid quality")
	}
}

func TestGetLanguageString(t *testing.T) {
	lang := Language{
		Primary:   "en",
		Secondary: "ru",
		Quality:   0.9,
	}
	str := GetLanguageString(lang)
	if !strings.Contains(str, "en") || !strings.Contains(str, "ru") || !strings.Contains(str, "0.9") {
		t.Error("GetLanguageString returned invalid format")
	}
}

func TestGetRandomLanguageString(t *testing.T) {
	str := GetRandomLanguageString()
	if str == "" {
		t.Error("GetRandomLanguageString returned empty string")
	}
	if !strings.Contains(str, ",") || !strings.Contains(str, ";q=") {
		t.Error("GetRandomLanguageString returned invalid format")
	}
}

func TestGetCustomLanguageString(t *testing.T) {
	str := GetCustomLanguageString("en", "ru", 0.9)
	if !strings.Contains(str, "en") || !strings.Contains(str, "ru") || !strings.Contains(str, "0.9") {
		t.Error("GetCustomLanguageString returned invalid format")
	}
}

func TestFormatQuality(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0.9, "0.9"},
		{0.95, "0.95"},
		{0.999, "0.99"},
		{0.1, "0.1"},
		{0.01, "0.01"},
	}

	for _, test := range tests {
		result := formatQuality(test.input)
		if result != test.expected {
			t.Errorf("formatQuality(%f) = %s, want %s", test.input, result, test.expected)
		}
	}
}
