# go-wc: A fast, compatible wc implementation in Go

Binary name: go_wc
Module path: github.com/rajasatyajit/go-wc
License: MIT

Goals
- Match GNU coreutils wc and POSIX wc behavior by default
- Provide additional options for longest line counted in characters
- Respect system locale (LC_ALL > LC_CTYPE > LANG) for word/char semantics; support explicit --encoding override
- Be fast: aim to meet or beat GNU wc throughput on common workloads
- Pure Go, latest Go release

Status
- Initial implementation with ASCII fast path, UTF-8 path, and non-UTF encodings via x/text/encoding
- Concurrency for multiple files with ordered output
- Benchmarks scaffolding included

Install
- With Go installed: 
  go install github.com/rajasatyajit/go-wc/cmd/go_wc@latest

Usage
  go_wc [OPTIONS] [FILE...]

Options
  -c, --bytes                print the byte counts
  -m, --chars                print the character counts
  -l, --lines                print the newline counts
  -w, --words                print the word counts
  -L, --max-line-length      print the maximum display width of lines in bytes (GNU-compatible)
      --max-line-length-chars
                            print the maximum line length in characters
      --files0-from=FILE    read input file names from FILE, separated by NULs; - means standard input
      --encoding=NAME       override detected locale encoding (e.g., utf-8, iso-8859-1, shift_jis)
      --jobs, -j N          process up to N files concurrently (default: GOMAXPROCS)
      --buffer-size BYTES   set I/O buffer size (default: 1MiB)
      --help                display this help and exit
      --version             output version information and exit

Behavior
- Default metrics when none of -cmlwL are specified: lines, words, bytes (GNU/POSIX)
- Multiple files: print per-file counts and a final total line
- "-" means standard input
- Lines are counted by newline bytes (\n)
- Words are maximal sequences of non-whitespace per current locale
- -L uses bytes; --max-line-length-chars uses characters

Performance notes
- ASCII fast path uses byte lookups and minimal branching
- UTF-8 path decodes runes without allocations and classifies whitespace via unicode.IsSpace
- Non-UTF encodings decoded with x/text/encoding; bytes still counted from raw stream
- Worker pool processes independent files in parallel while preserving output order

Benchmarks
- Run micro and e2e benchmarks with:
  go test ./... -bench . -benchmem

Limitations
- Locale coverage for non-UTF encodings depends on x/text support
- Display width is not computed; -L is bytes, --max-line-length-chars is character count, which may differ from terminal column width

Contributing
- PRs welcome. Please run tests and benchmarks before submitting.

