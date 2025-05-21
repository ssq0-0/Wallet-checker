package useragent

import (
	"strings"
	"testing"
)

func TestGetRandomUserAgent(t *testing.T) {
	ua := GetRandomUserAgent()
	if ua == "" {
		t.Error("GetRandomUserAgent returned empty string")
	}
	if !strings.Contains(ua, "Mozilla/5.0") {
		t.Error("GetRandomUserAgent returned invalid User-Agent")
	}
}

func TestGetPlatformSpecificUserAgent(t *testing.T) {
	ua := GetPlatformSpecificUserAgent()
	if ua == "" {
		t.Error("GetPlatformSpecificUserAgent returned empty string")
	}
	if !strings.Contains(ua, "Mozilla/5.0") {
		t.Error("GetPlatformSpecificUserAgent returned invalid User-Agent")
	}
}

func TestGetSecChUa(t *testing.T) {
	secChUa := GetSecChUa()
	if secChUa == "" {
		t.Error("GetSecChUa returned empty string")
	}
	if !strings.Contains(secChUa, "Google Chrome") {
		t.Error("GetSecChUa returned invalid value")
	}
}

func TestGetPlatform(t *testing.T) {
	platform := GetPlatform()
	if platform == "" {
		t.Error("GetPlatform returned empty string")
	}
	if !strings.Contains(platform, `"`) {
		t.Error("GetPlatform returned invalid value")
	}
}

func TestGetCurrentPlatform(t *testing.T) {
	platform := getCurrentPlatform()
	if platform == "" {
		t.Error("getCurrentPlatform returned empty string")
	}
	validPlatforms := map[string]bool{
		"Macintosh": true,
		"Windows":   true,
		"Linux":     true,
		"Unknown":   true,
	}
	if !validPlatforms[platform] {
		t.Errorf("getCurrentPlatform returned invalid platform: %s", platform)
	}
}
