package main

import (
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCfg  cliConfig
		expectedRem  []string
		expectError  bool
	}{
		{
			name: "default config",
			args: []string{},
			expectedCfg: cliConfig{
				jobs:    runtime.GOMAXPROCS(0),
				bufSize: 1 * 1024 * 1024,
			},
			expectedRem: []string{},
		},
		{
			name: "count bytes short flag",
			args: []string{"-c", "file.txt"},
			expectedCfg: cliConfig{
				countBytes: true,
				jobs:       runtime.GOMAXPROCS(0),
				bufSize:    1 * 1024 * 1024,
			},
			expectedRem: []string{"file.txt"},
		},
		{
			name: "count bytes long flag",
			args: []string{"--bytes", "file.txt"},
			expectedCfg: cliConfig{
				countBytes: true,
				jobs:       runtime.GOMAXPROCS(0),
				bufSize:    1 * 1024 * 1024,
			},
			expectedRem: []string{"file.txt"},
		},
		{
			name: "multiple flags",
			args: []string{"-l", "-w", "-c", "file1.txt", "file2.txt"},
			expectedCfg: cliConfig{
				countLines: true,
				countWords: true,
				countBytes: true,
				jobs:       runtime.GOMAXPROCS(0),
				bufSize:    1 * 1024 * 1024,
			},
			expectedRem: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "max line length flags",
			args: []string{"-L", "--max-line-length-chars"},
			expectedCfg: cliConfig{
				countMaxBytes: true,
				countMaxChars: true,
				jobs:          runtime.GOMAXPROCS(0),
				bufSize:       1 * 1024 * 1024,
			},
			expectedRem: []string{},
		},
		{
			name: "custom jobs and buffer size",
			args: []string{"-j", "4", "--buffer-size", "2048"},
			expectedCfg: cliConfig{
				jobs:    4,
				bufSize: 2048,
			},
			expectedRem: []string{},
		},
		{
			name: "files0-from and encoding",
			args: []string{"--files0-from", "filelist.txt", "--encoding", "utf-8"},
			expectedCfg: cliConfig{
				files0From: "filelist.txt",
				encoding:   "utf-8",
				jobs:       runtime.GOMAXPROCS(0),
				bufSize:    1 * 1024 * 1024,
			},
			expectedRem: []string{},
		},
		{
			name: "help flag",
			args: []string{"--help"},
			expectedCfg: cliConfig{
				showHelp: true,
				jobs:     runtime.GOMAXPROCS(0),
				bufSize:  1 * 1024 * 1024,
			},
			expectedRem: []string{},
		},
		{
			name: "version flag",
			args: []string{"--version"},
			expectedCfg: cliConfig{
				showVer: true,
				jobs:    runtime.GOMAXPROCS(0),
				bufSize: 1 * 1024 * 1024,
			},
			expectedRem: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, rem, err := parseArgs(tt.args)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if !reflect.DeepEqual(cfg, tt.expectedCfg) {
				t.Errorf("Config mismatch:\ngot:  %+v\nwant: %+v", cfg, tt.expectedCfg)
			}
			
			if !reflect.DeepEqual(rem, tt.expectedRem) {
				t.Errorf("Remaining args mismatch:\ngot:  %v\nwant: %v", rem, tt.expectedRem)
			}
		})
	}
}

func TestReadFiles0From(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "single file",
			content:  "file1.txt\x00",
			expected: []string{"file1.txt"},
		},
		{
			name:     "multiple files",
			content:  "file1.txt\x00file2.txt\x00file3.txt\x00",
			expected: []string{"file1.txt", "file2.txt", "file3.txt"},
		},
		{
			name:     "files without trailing null",
			content:  "file1.txt\x00file2.txt",
			expected: []string{"file1.txt", "file2.txt"},
		},
		{
			name:     "empty entries filtered out",
			content:  "file1.txt\x00\x00file2.txt\x00",
			expected: []string{"file1.txt", "file2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpFile, err := os.CreateTemp("", "test_files0_")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			
			// Write test content
			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()
			
			// Test the function
			result, err := readFiles0From(tmpFile.Name())
			if err != nil {
				t.Fatalf("readFiles0From failed: %v", err)
			}
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Result mismatch:\ngot:  %v\nwant: %v", result, tt.expected)
			}
		})
	}
}

func TestReadFiles0FromStdin(t *testing.T) {
	// Test reading from stdin (represented by "-")
	content := "file1.txt\x00file2.txt\x00"
	
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	
	// Save original stdin and restore after test
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()
	os.Stdin = r
	
	// Write content to pipe in a goroutine
	go func() {
		defer w.Close()
		w.WriteString(content)
	}()
	
	// Test the function
	result, err := readFiles0From("-")
	if err != nil {
		t.Fatalf("readFiles0From failed: %v", err)
	}
	
	expected := []string{"file1.txt", "file2.txt"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Result mismatch:\ngot:  %v\nwant: %v", result, expected)
	}
}

func TestReadFiles0FromNonexistentFile(t *testing.T) {
	_, err := readFiles0From("/nonexistent/file")
	if err == nil {
		t.Error("Expected error for nonexistent file, but got none")
	}
}

// Test helper functions and edge cases
func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("Version should not be empty")
	}
}

func TestUsageDoesNotPanic(t *testing.T) {
	// Just ensure usage() doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("usage() panicked: %v", r)
		}
	}()
	
	// Capture output to avoid cluttering test output
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	usage()
	
	w.Close()
	os.Stdout = origStdout
	
	// Read the output to ensure it's not empty
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	
	if !strings.Contains(output, "go_wc") {
		t.Error("Usage output should contain 'go_wc'")
	}
}

// Benchmark tests
func BenchmarkParseArgs(b *testing.B) {
	args := []string{"-l", "-w", "-c", "file1.txt", "file2.txt", "file3.txt"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseArgs(args)
	}
}

func BenchmarkReadFiles0From(b *testing.B) {
	// Create a temporary file with test content
	content := strings.Repeat("file.txt\x00", 100)
	tmpFile, err := os.CreateTemp("", "bench_files0_")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	tmpFile.WriteString(content)
	tmpFile.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readFiles0From(tmpFile.Name())
	}
}