package format

import (
	"testing"

	"github.com/rajasatyajit/go-wc/pkg/wc"
)

func TestComputeWidth(t *testing.T) {
	tests := []struct {
		name     string
		results  []wc.FileResult
		totals   wc.FileResult
		metrics  wc.Metrics
		expected int
	}{
		{
			name: "minimum width is 7",
			results: []wc.FileResult{
				{Lines: 1, Words: 2, Bytes: 3},
			},
			totals:   wc.FileResult{Lines: 1, Words: 2, Bytes: 3},
			metrics:  wc.Metrics{Lines: true, Words: true, Bytes: true},
			expected: 7,
		},
		{
			name: "width based on largest number",
			results: []wc.FileResult{
				{Lines: 123456789, Words: 2, Bytes: 3},
			},
			totals:   wc.FileResult{Lines: 123456789, Words: 2, Bytes: 3},
			metrics:  wc.Metrics{Lines: true, Words: true, Bytes: true},
			expected: 9,
		},
		{
			name: "totals larger than individual results",
			results: []wc.FileResult{
				{Lines: 100, Words: 200, Bytes: 300},
				{Lines: 150, Words: 250, Bytes: 350},
			},
			totals:   wc.FileResult{Lines: 250, Words: 450, Bytes: 650},
			metrics:  wc.Metrics{Lines: true, Words: true, Bytes: true},
			expected: 7,
		},
		{
			name: "skip results with errors",
			results: []wc.FileResult{
				{Lines: 123456789, Err: nil},
				{Lines: 999999999, Err: &testError{}},
			},
			totals:   wc.FileResult{Lines: 123456789},
			metrics:  wc.Metrics{Lines: true},
			expected: 9,
		},
		{
			name: "max line metrics",
			results: []wc.FileResult{
				{MaxLineBytes: 12345, MaxLineChars: 6789},
			},
			totals:   wc.FileResult{MaxLineBytes: 12345, MaxLineChars: 6789},
			metrics:  wc.Metrics{MaxLineBytes: true, MaxLineChars: true},
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeWidth(tt.results, tt.totals, tt.metrics)
			if result != tt.expected {
				t.Errorf("ComputeWidth() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestFormatLine(t *testing.T) {
	tests := []struct {
		name     string
		result   wc.FileResult
		metrics  wc.Metrics
		width    int
		expected string
	}{
		{
			name:     "all metrics with filename",
			result:   wc.FileResult{Lines: 10, Words: 20, Chars: 30, Bytes: 40, MaxLineBytes: 50, MaxLineChars: 60, Filename: "test.txt"},
			metrics:  wc.Metrics{Lines: true, Words: true, Chars: true, Bytes: true, MaxLineBytes: true, MaxLineChars: true},
			width:    7,
			expected: "     10      20      30      40      50      60 test.txt",
		},
		{
			name:     "only lines and words",
			result:   wc.FileResult{Lines: 5, Words: 15, Filename: "file.txt"},
			metrics:  wc.Metrics{Lines: true, Words: true},
			width:    7,
			expected: "      5      15 file.txt",
		},
		{
			name:     "no filename",
			result:   wc.FileResult{Lines: 1, Words: 2},
			metrics:  wc.Metrics{Lines: true, Words: true},
			width:    7,
			expected: "      1       2",
		},
		{
			name:     "empty metrics",
			result:   wc.FileResult{Filename: "empty.txt"},
			metrics:  wc.Metrics{},
			width:    7,
			expected: "empty.txt",
		},
		{
			name:     "large numbers",
			result:   wc.FileResult{Bytes: 123456789, Filename: "big.txt"},
			metrics:  wc.Metrics{Bytes: true},
			width:    10,
			expected: " 123456789 big.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLine(tt.result, tt.metrics, tt.width)
			if result != tt.expected {
				t.Errorf("FormatLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "empty slice",
			parts:    []string{},
			expected: "",
		},
		{
			name:     "single part",
			parts:    []string{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple parts",
			parts:    []string{"hello", "world", "test"},
			expected: "hello world test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := join(tt.parts)
			if result != tt.expected {
				t.Errorf("join() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		value    uint64
		width    int
		expected string
	}{
		{
			name:     "no padding needed",
			value:    123,
			width:    3,
			expected: "123",
		},
		{
			name:     "padding needed",
			value:    42,
			width:    5,
			expected: "   42",
		},
		{
			name:     "zero value",
			value:    0,
			width:    4,
			expected: "   0",
		},
		{
			name:     "width smaller than number",
			value:    12345,
			width:    3,
			expected: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padRight(tt.value, tt.width)
			if result != tt.expected {
				t.Errorf("padRight() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// testError is a simple error implementation for testing
type testError struct{}

func (e *testError) Error() string {
	return "test error"
}