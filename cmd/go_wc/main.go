package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rajasatyajit/go-wc/pkg/wc"
	"github.com/rajasatyajit/go-wc/pkg/wc/format"
	"github.com/rajasatyajit/go-wc/pkg/wc/locale"
)

const version = "0.1.0"

// cliConfig holds parsed CLI options
type cliConfig struct {
	countBytes bool
	countChars bool
	countLines bool
	countWords bool
	countMaxBytes bool
	countMaxChars bool

	files0From string
	encoding   string
	jobs       int
	bufSize    int
	showHelp   bool
	showVer    bool
}

func parseArgs(args []string) (cliConfig, []string, error) {
	var cfg cliConfig
	fs := flag.NewFlagSet("go_wc", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.BoolVar(&cfg.countBytes, "c", false, "")
	fs.BoolVar(&cfg.countBytes, "bytes", false, "")
	fs.BoolVar(&cfg.countChars, "m", false, "")
	fs.BoolVar(&cfg.countChars, "chars", false, "")
	fs.BoolVar(&cfg.countLines, "l", false, "")
	fs.BoolVar(&cfg.countLines, "lines", false, "")
	fs.BoolVar(&cfg.countWords, "w", false, "")
	fs.BoolVar(&cfg.countWords, "words", false, "")
	fs.BoolVar(&cfg.countMaxBytes, "L", false, "")
	fs.BoolVar(&cfg.countMaxBytes, "max-line-length", false, "")
	fs.BoolVar(&cfg.countMaxChars, "max-line-length-chars", false, "")

	fs.StringVar(&cfg.files0From, "files0-from", "", "")
	fs.StringVar(&cfg.encoding, "encoding", "", "")
	fs.IntVar(&cfg.jobs, "jobs", runtime.GOMAXPROCS(0), "")
	fs.IntVar(&cfg.jobs, "j", runtime.GOMAXPROCS(0), "")
	fs.IntVar(&cfg.bufSize, "buffer-size", 1*1024*1024, "")
	fs.BoolVar(&cfg.showHelp, "help", false, "")
	fs.BoolVar(&cfg.showVer, "version", false, "")

	if err := fs.Parse(args); err != nil {
		return cfg, nil, err
	}
	rem := fs.Args()
	return cfg, rem, nil
}

func usage() {
	fmt.Println("go_wc - compatible and fast wc implementation in pure Go")
	fmt.Println("Usage: go_wc [OPTIONS] [FILE...]")
	fmt.Println("Options:")
	fmt.Println("  -c, --bytes                 print the byte counts")
	fmt.Println("  -m, --chars                 print the character counts")
	fmt.Println("  -l, --lines                 print the newline counts")
	fmt.Println("  -w, --words                 print the word counts")
	fmt.Println("  -L, --max-line-length       print the maximum line length in bytes")
	fmt.Println("      --max-line-length-chars print the maximum line length in characters")
	fmt.Println("      --files0-from=FILE      read input file names from FILE, separated by NULs; - means standard input")
	fmt.Println("      --encoding=NAME         override detected locale encoding (e.g., utf-8)")
	fmt.Println("  -j, --jobs N                process up to N files concurrently (default: GOMAXPROCS)")
	fmt.Println("      --buffer-size BYTES     set I/O buffer size (default: 1MiB)")
	fmt.Println("      --help                  display this help and exit")
	fmt.Println("      --version               output version information and exit")
}

func main() {
	cfg, files, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		usage()
		os.Exit(1)
	}
	if cfg.showHelp {
		usage()
		return
	}
	if cfg.showVer {
		fmt.Printf("go_wc version %s\n", version)
		return
	}

	metrics := wc.Metrics{}
	if !(cfg.countBytes || cfg.countChars || cfg.countLines || cfg.countWords || cfg.countMaxBytes || cfg.countMaxChars) {
		// default: lines, words, bytes
		metrics.Lines = true
		metrics.Words = true
		metrics.Bytes = true
	} else {
		metrics.Bytes = cfg.countBytes
		metrics.Chars = cfg.countChars
		metrics.Lines = cfg.countLines
		metrics.Words = cfg.countWords
		metrics.MaxLineBytes = cfg.countMaxBytes
		metrics.MaxLineChars = cfg.countMaxChars
	}

	// Build file list possibly augmented by --files0-from
	inputs := make([]string, 0, len(files)+8)
	inputs = append(inputs, files...)
	if cfg.files0From != "" {
		names, ferr := readFiles0From(cfg.files0From)
		if ferr != nil {
			fmt.Fprintln(os.Stderr, ferr)
			os.Exit(1)
		}
		inputs = append(inputs, names...)
	}
	if len(inputs) == 0 {
		inputs = []string{"-"}
	}

	loc := locale.Detect(cfg.encoding)

	opts := wc.Options{BufferSize: cfg.bufSize, Locale: loc}

	// Prepare jobs and worker pool
	type job struct {
		idx  int
		name string
	}
	jobs := make(chan job)
	results := make(chan wc.FileResult)
	var wg sync.WaitGroup
	workerCount := cfg.jobs
	if workerCount < 1 {
		workerCount = 1
	}

	stdinOnce := sync.Once{}
	var stdinData []byte // If stdin referenced multiple times, slurp once.
	stdinErr := error(nil)

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			var fr wc.FileResult
			start := time.Now()
			if j.name == "-" {
				stdinOnce.Do(func() {
					stdinData, stdinErr = io.ReadAll(bufio.NewReaderSize(os.Stdin, opts.BufferSize))
				})
				if stdinErr != nil {
					fr = wc.FileResult{Filename: j.name, Err: stdinErr}
				} else {
					fr = wc.CountBytes(stdinData, metrics, opts)
					fr.Filename = j.name
				}
			} else {
				f, e := os.Open(j.name)
				if e != nil {
					fr = wc.FileResult{Filename: j.name, Err: e}
				} else {
					fr = wc.CountReader(bufio.NewReaderSize(f, opts.BufferSize), metrics, opts)
					fr.Filename = j.name
					_ = f.Close()
				}
			}
			fr.Duration = time.Since(start)
			fr.Index = j.idx
			results <- fr
		}
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker()
	}
	go func() {
		for i, name := range inputs {
			jobs <- job{idx: i, name: name}
		}
		close(jobs)
	}()

	// Collect and print in order
	pending := make(map[int]wc.FileResult)
	next := 0
	var exitCode int

	all := make([]wc.FileResult, 0, len(inputs))
	for range inputs {
		res := <-results
		if res.Err != nil {
			exitCode = 1
		}
		pending[res.Index] = res
		for {
			if pr, ok := pending[next]; ok {
				all = append(all, pr)
				delete(pending, next)
				next++
			} else {
				break
			}
		}
	}
	wg.Wait()

	// Compute totals and formatting
	var totals wc.FileResult
	multiple := len(inputs) > 1
	for _, r := range all {
		if r.Err == nil {
			totals.Lines += r.Lines
			totals.Words += r.Words
			totals.Bytes += r.Bytes
			totals.Chars += r.Chars
			if r.MaxLineBytes > totals.MaxLineBytes {
				totals.MaxLineBytes = r.MaxLineBytes
			}
			if r.MaxLineChars > totals.MaxLineChars {
				totals.MaxLineChars = r.MaxLineChars
			}
		}
	}

	// Determine column width based on all results and totals
	width := format.ComputeWidth(all, totals, metrics)

	// Print results
	for _, r := range all {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "go_wc: %s: %v\n", r.Filename, r.Err)
			continue
		}
		fmt.Println(format.FormatLine(r, metrics, width))
	}
	if multiple {
		totals.Filename = "total"
		fmt.Println(format.FormatLine(totals, metrics, width))
	}

	os.Exit(exitCode)
}

func readFiles0From(path string) ([]string, error) {
	var r io.Reader
	if path == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(filepath.Clean(path))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(data), "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

