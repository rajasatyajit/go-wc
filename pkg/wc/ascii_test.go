package wc

import "testing"

func TestASCIISpaceTable(t *testing.T) {
	// Test that ASCII space characters are correctly identified
	expectedSpaces := []byte{'\t', '\n', '\v', '\f', '\r', ' '}
	
	for _, b := range expectedSpaces {
		if !asciiSpace[b] {
			t.Errorf("Expected byte %d (%q) to be marked as space", b, b)
		}
	}
	
	// Test some non-space characters
	nonSpaces := []byte{'a', 'A', '0', '!', '@', '#', 'z', 'Z', '9'}
	for _, b := range nonSpaces {
		if asciiSpace[b] {
			t.Errorf("Expected byte %d (%q) to NOT be marked as space", b, b)
		}
	}
	
	// Test all 256 possible byte values to ensure table is complete
	spaceCount := 0
	for i := 0; i < 256; i++ {
		if asciiSpace[i] {
			spaceCount++
		}
	}
	
	// Should have exactly 6 space characters
	if spaceCount != 6 {
		t.Errorf("Expected exactly 6 space characters, got %d", spaceCount)
	}
}

func TestASCIISpaceTableCompleteness(t *testing.T) {
	// Verify the table has the correct size
	if len(asciiSpace) != 256 {
		t.Errorf("ASCII space table should have 256 entries, got %d", len(asciiSpace))
	}
	
	// Test boundary values
	testCases := []struct {
		byte     byte
		expected bool
		name     string
	}{
		{0, false, "null"},
		{8, false, "backspace"},
		{9, true, "tab"},
		{10, true, "newline"},
		{11, true, "vertical tab"},
		{12, true, "form feed"},
		{13, true, "carriage return"},
		{14, false, "shift out"},
		{31, false, "unit separator"},
		{32, true, "space"},
		{33, false, "exclamation"},
		{127, false, "delete"},
		{128, false, "high bit set"},
		{255, false, "max byte value"},
	}
	
	for _, tc := range testCases {
		if asciiSpace[tc.byte] != tc.expected {
			t.Errorf("Byte %d (%s): expected %v, got %v", tc.byte, tc.name, tc.expected, asciiSpace[tc.byte])
		}
	}
}

// Benchmark the ASCII space lookup
func BenchmarkASCIISpaceLookup(b *testing.B) {
	testBytes := []byte("hello world\ttest\nline\r\n")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, bt := range testBytes {
			_ = asciiSpace[bt]
		}
	}
}