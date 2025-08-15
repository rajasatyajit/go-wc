# GNU and POSIX wc behavior targets

Options
- -c, --bytes: print the byte counts
- -m, --chars: print the character counts
- -l, --lines: print the newline counts (count of '\n')
- -w, --words: print the word counts
- -L, --max-line-length: print the maximum line length in bytes
- --max-line-length-chars: print the maximum line length in characters (extension)
- --files0-from=FILE: read NUL-delimited file list from FILE (or '-' for stdin)
- --encoding=NAME: override detected encoding (utf-8 default; see README for supported names)
- -j, --jobs N: process up to N files concurrently
- --buffer-size BYTES: set buffer size
- --help, --version

Default behavior
- With no -cmlwL options, behave as 'wc -l -w -c': print newline, word, and byte counts.
- When multiple files are provided, print one line per file followed by a 'total' line which is the sum across readable files.
- If an input is '-', read standard input.

Output formatting
- Right-align numeric columns. Minimum width 7; widen columns to accommodate the largest value.
- Field order when multiple are selected: newline, word, character (-m), byte (-c), max-line-length (-L), then filename. The '--max-line-length-chars' field, when requested, follows the byte max-line-length.

Exit status
- 0: All files processed successfully
- 1: Any error occurred (invalid options or unreadable files). Continue processing remaining files when possible.

Word definition
- Words are maximal sequences of non-whitespace according to the active locale:
  - C/POSIX locale: whitespace is ASCII [\t, \n, \v, \f, \r, space].
  - UTF-8 or other multibyte locales: classify using Unicode whitespace (unicode.IsSpace on decoded runes).

Character and line length
- -m counts characters per active locale. In UTF-8: runes; invalid byte sequences count as a single character each (RuneError with size 1).
- -L is bytes per line (GNU behavior). '--max-line-length-chars' is characters per line (extension), both considering lines split on '\n'.

Locale and encoding detection
- Honor LC_ALL, then LC_CTYPE, then LANG to detect locale and encoding. Default to UTF-8 when unspecified.
- For non-UTF encodings supported by golang.org/x/text/encoding, decode to runes for -m/-w; byte counts reflect raw input bytes.

Stdin handling
- If '-' appears multiple times, standard input is consumed once and its result is reused for each occurrence for printing while the total remains correct.

Concurrency
- Multiple files may be processed concurrently; output order must match the input order.
- Standard input is processed synchronously.

Limits and counters
- Use 64-bit counters. Behavior for values exceeding uint64 is undefined.

