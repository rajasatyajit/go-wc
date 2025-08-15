package locale

import (
	"os"
	"strings"
)

type Info struct {
	Encoding    string
	IsUTF8      bool
	IsCOrPOSIX  bool
}

// Detect reads environment (LC_ALL > LC_CTYPE > LANG) and returns locale Info.
// If override is non-empty, it is used directly.
func Detect(override string) Info {
	if override != "" {
		enc := normalizeEncoding(override)
		return Info{Encoding: enc, IsUTF8: enc == "utf-8", IsCOrPOSIX: enc == "C" || enc == "POSIX"}
	}
	val := firstNonEmpty(os.Getenv("LC_ALL"), os.Getenv("LC_CTYPE"), os.Getenv("LANG"))
	if val == "" { return Info{Encoding: "utf-8", IsUTF8: true} }
	// Examples: en_US.UTF-8, C, POSIX, de_DE.ISO-8859-1
	up := val
	if up == "C" || up == "POSIX" {
		return Info{Encoding: "C", IsCOrPOSIX: true}
	}
	// split on '.' to find codeset
	enc := "utf-8"
	if i := strings.IndexByte(up, '.'); i >= 0 && i+1 < len(up) {
		enc = up[i+1:]
	}
	enc = normalizeEncoding(enc)
	return Info{Encoding: enc, IsUTF8: enc == "utf-8"}
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" { return s }
	}
	return ""
}

func normalizeEncoding(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "charset=", "")
	s = strings.ReplaceAll(s, "cs", "")
	s = strings.TrimPrefix(s, "")
	// Common aliases
	switch s {
	case "utf8": return "utf-8"
	case "c": return "C"
	case "posix": return "POSIX"
	}
	return s
}

