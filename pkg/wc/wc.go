package wc

import (
	"bufio"
	"io"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/rajasatyajit/go-wc/pkg/wc/locale"
)

// Metrics selects which counters to compute
 type Metrics struct {
	Lines         bool
	Words         bool
	Bytes         bool
	Chars         bool
	MaxLineBytes  bool
	MaxLineChars  bool
 }

// Options control scanning behavior
 type Options struct {
	BufferSize int
	Locale     locale.Info
 }

// FileResult holds counts for a single file
 type FileResult struct {
	Index         int
	Filename      string
	Lines         uint64
	Words         uint64
	Bytes         uint64
	Chars         uint64
	MaxLineBytes  uint64
	MaxLineChars  uint64
	Err           error
	Duration      time.Duration
 }

// CountReader processes counts from an io.Reader
 func CountReader(r *bufio.Reader, m Metrics, opt Options) FileResult {
	buf := make([]byte, opt.BufferSize)
	var res FileResult
	prevSpace := true
	var curLineBytes uint64
	var curLineChars uint64
	localeInfo := opt.Locale
	asciiMode := localeInfo.IsCOrPOSIX || localeInfo.IsUTF8 // start in ASCII fast path when possible
	carry := make([]byte, 0, 4)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			res.Bytes += uint64(n)

			if asciiMode {
				// If in ASCII mode, check for any non-ASCII to potentially switch
				if !localeInfo.IsCOrPOSIX {
					for _, b := range chunk {
						if b >= 0x80 {
							asciiMode = false
							break
						}
					}
				}
				if asciiMode {
					// Process with ASCII fast path
					for _, b := range chunk {
					if m.Lines && b == '\n' {
						res.Lines++
						if m.MaxLineBytes && curLineBytes > res.MaxLineBytes {
							res.MaxLineBytes = curLineBytes
						}
						if m.MaxLineChars && curLineChars > res.MaxLineChars {
							res.MaxLineChars = curLineChars
						}
						curLineBytes = 0
						curLineChars = 0
					} else {
							if m.MaxLineBytes {
								curLineBytes++
							}
							if m.MaxLineChars {
								curLineChars++
							}
						}
						// word counting in ASCII space
						if m.Words {
							isSpace := asciiSpace[b]
							if !isSpace && prevSpace {
								res.Words++
							}
							prevSpace = isSpace
						}
					}
					// ASCII mode: chars equals bytes if requested
					if m.Chars {
						res.Chars += uint64(n)
					}
					continue
				}
			}

			// UTF-8 or multibyte path: use rune decoding
			data := append(carry, chunk...)
			carry = carry[:0]
			for len(data) > 0 {
				r, size := utf8.DecodeRune(data)
				if r == utf8.RuneError && size == 1 {
					// invalid byte; count as one char and advance one
					if m.Chars {
						res.Chars++
					}
					if m.MaxLineBytes {
						curLineBytes++
					}
					if m.MaxLineChars {
						curLineChars++
					}
					b := data[0]
					if m.Lines && b == '\n' {
						res.Lines++
						if m.MaxLineBytes && curLineBytes > res.MaxLineBytes {
							res.MaxLineBytes = curLineBytes
						}
						if m.MaxLineChars && curLineChars > res.MaxLineChars {
							res.MaxLineChars = curLineChars
						}
						curLineBytes = 0
						curLineChars = 0
					}
					data = data[1:]
					if m.Words {
						sp := asciiSpace[b]
						if !sp && prevSpace {
							res.Words++
						}
						prevSpace = sp
					}
					continue
				}

				if m.Chars {
					res.Chars++
				}
				if m.Words {
					sp := unicode.IsSpace(r)
					if !sp && prevSpace {
						res.Words++
					}
					prevSpace = sp
				}
				if m.Lines {
					// lines counted by raw '\n' byte, but we can infer from rune if newline
					if r == '\n' {
						res.Lines++
						if m.MaxLineBytes && curLineBytes > res.MaxLineBytes {
							res.MaxLineBytes = curLineBytes
						}
						if m.MaxLineChars && curLineChars > res.MaxLineChars {
							res.MaxLineChars = curLineChars
						}
						curLineBytes = 0
						curLineChars = 0
					} else {
						if m.MaxLineBytes {
							curLineBytes += uint64(size)
						}
						if m.MaxLineChars {
							curLineChars++
						}
					}
				} else {
					// not counting lines, still need to advance max len counters per byte/char
					if m.MaxLineBytes {
						curLineBytes += uint64(size)
					}
					if m.MaxLineChars {
						curLineChars++
					}
				}

				data = data[size:]
			}
			// keep any partial for the next read
			if len(chunk) > 0 {
				// Any leftover in data are partial rune bytes (0..3)
				if len(data) > 0 {
					carry = append(carry, data...)
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			res.Err = err
			break
		}
	}
	// EOF: finalize max line metrics (for last line without trailing newline)
	if res.Err == nil {
		if m.MaxLineBytes && curLineBytes > res.MaxLineBytes {
			res.MaxLineBytes = curLineBytes
		}
		if m.MaxLineChars && curLineChars > res.MaxLineChars {
			res.MaxLineChars = curLineChars
		}
	}
	return res
 }

// CountBytes is a helper to count from an in-memory byte slice efficiently
 func CountBytes(b []byte, m Metrics, opt Options) FileResult {
	br := bufio.NewReaderSize(&bytesReader{b: b}, opt.BufferSize)
	return CountReader(br, m, opt)
 }

// bytesReader avoids allocations like bytes.NewReader for small code
 type bytesReader struct { b []byte; off int }

 func (r *bytesReader) Read(p []byte) (int, error) {
	if r.off >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.off:])
	r.off += n
	return n, nil
 }

