package format

import (
	"strconv"

	"github.com/rajasatyajit/go-wc/pkg/wc"
)

// ComputeWidth decides the minimum column width required for alignment
func ComputeWidth(results []wc.FileResult, totals wc.FileResult, m wc.Metrics) int {
	max := uint64(0)
	for _, r := range results {
		if r.Err != nil { continue }
		if m.Lines && r.Lines > max { max = r.Lines }
		if m.Words && r.Words > max { max = r.Words }
		if m.Chars && r.Chars > max { max = r.Chars }
		if m.Bytes && r.Bytes > max { max = r.Bytes }
		if m.MaxLineBytes && r.MaxLineBytes > max { max = r.MaxLineBytes }
		if m.MaxLineChars && r.MaxLineChars > max { max = r.MaxLineChars }
	}
	if m.Lines && totals.Lines > max { max = totals.Lines }
	if m.Words && totals.Words > max { max = totals.Words }
	if m.Chars && totals.Chars > max { max = totals.Chars }
	if m.Bytes && totals.Bytes > max { max = totals.Bytes }
	if m.MaxLineBytes && totals.MaxLineBytes > max { max = totals.MaxLineBytes }
	if m.MaxLineChars && totals.MaxLineChars > max { max = totals.MaxLineChars }
	w := len(strconv.FormatUint(max, 10))
	if w < 7 { w = 7 }
	return w
}

// FormatLine formats a single file result
func FormatLine(r wc.FileResult, m wc.Metrics, width int) string {
	// Order: lines, words, chars, bytes, max-line-bytes, max-line-chars
	parts := make([]string, 0, 6)
	if m.Lines { parts = append(parts, padRight(r.Lines, width)) }
	if m.Words { parts = append(parts, padRight(r.Words, width)) }
	if m.Chars { parts = append(parts, padRight(r.Chars, width)) }
	if m.Bytes { parts = append(parts, padRight(r.Bytes, width)) }
	if m.MaxLineBytes { parts = append(parts, padRight(r.MaxLineBytes, width)) }
	if m.MaxLineChars { parts = append(parts, padRight(r.MaxLineChars, width)) }
	if r.Filename != "" { parts = append(parts, r.Filename) }
	return join(parts)
}

func join(parts []string) string {
	if len(parts) == 0 { return "" }
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " " + parts[i]
	}
	return out
}

func padRight(v uint64, width int) string {
	s := strconv.FormatUint(v, 10)
	for len(s) < width { s = " " + s }
	return s
}

