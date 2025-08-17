package locale

import (
	"os"
	"testing"
)

func TestNormalizeEncoding(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UTF-8", "utf-8"},
		{"utf8", "utf-8"},
		{"UTF8", "utf-8"},
		{"C", "C"},
		{"POSIX", "POSIX"},
		{"posix", "POSIX"},
		{"c", "C"},
		{"ISO-8859-1", "iso-8859-1"},
		{"iso_8859_1", "iso-8859-1"},
		{"charset=utf-8", "utf-8"},
		{"csutf8", "utf-8"},
		{"  UTF-8  ", "utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeEncoding(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeEncoding(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{""}, ""},
		{[]string{"", "", ""}, ""},
		{[]string{"first"}, "first"},
		{[]string{"", "second"}, "second"},
		{[]string{"first", "second"}, "first"},
		{[]string{"", "", "third"}, "third"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := firstNonEmpty(tt.input...)
			if result != tt.expected {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	// Save original environment
	origLCALL := os.Getenv("LC_ALL")
	origLCCTYPE := os.Getenv("LC_CTYPE")
	origLANG := os.Getenv("LANG")

	// Clean up after test
	defer func() {
		os.Setenv("LC_ALL", origLCALL)
		os.Setenv("LC_CTYPE", origLCCTYPE)
		os.Setenv("LANG", origLANG)
	}()

	tests := []struct {
		name     string
		override string
		lcAll    string
		lcCtype  string
		lang     string
		expected Info
	}{
		{
			name:     "override takes precedence",
			override: "iso-8859-1",
			lcAll:    "en_US.UTF-8",
			expected: Info{Encoding: "iso-8859-1", IsUTF8: false, IsCOrPOSIX: false},
		},
		{
			name:     "LC_ALL takes precedence",
			lcAll:    "en_US.UTF-8",
			lcCtype:  "C",
			lang:     "de_DE.ISO-8859-1",
			expected: Info{Encoding: "utf-8", IsUTF8: true, IsCOrPOSIX: false},
		},
		{
			name:     "LC_CTYPE when LC_ALL empty",
			lcCtype:  "C",
			lang:     "en_US.UTF-8",
			expected: Info{Encoding: "C", IsUTF8: false, IsCOrPOSIX: true},
		},
		{
			name:     "LANG when others empty",
			lang:     "de_DE.ISO-8859-1",
			expected: Info{Encoding: "iso-8859-1", IsUTF8: false, IsCOrPOSIX: false},
		},
		{
			name:     "POSIX locale",
			lcAll:    "POSIX",
			expected: Info{Encoding: "C", IsUTF8: false, IsCOrPOSIX: true},
		},
		{
			name:     "default when all empty",
			expected: Info{Encoding: "utf-8", IsUTF8: true, IsCOrPOSIX: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("LC_ALL", tt.lcAll)
			os.Setenv("LC_CTYPE", tt.lcCtype)
			os.Setenv("LANG", tt.lang)

			result := Detect(tt.override)
			if result != tt.expected {
				t.Errorf("Detect() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}