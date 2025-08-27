// NeuroScript Version: 0.3.0
// File version: 3
// Purpose: Updates function to use the centralized sentinel error from the lang package.
// filename: neuroscript/pkg/nsio/cleaner.go
// nlines: 55
// risk_rating: LOW

package nsio

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// CleanNS trims or rejects suspicious bytes before ANTLR sees the script.
// * Keeps TAB, LF and CR (converted to LF). Drops every other < 0x20.
// * Verifies UTF-8; replacement chars are an error (caller decides what to do).
// * Removes BOM if present.
// * Collapses CRLF -> LF.
// Returns cleaned bytes or an error.
func CleanNS(r io.Reader, maxBytes int) ([]byte, error) {
	br := bufio.NewReader(io.LimitReader(r, int64(maxBytes+1)))
	var out bytes.Buffer

	// header BOM?
	bom, _ := br.Peek(3)
	if bytes.Equal(bom, []byte{0xEF, 0xBB, 0xBF}) {
		_, _ = br.Discard(3)
	}

	for {
		rn, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read rune: %w", err)
		}
		if rn == '\r' { // normalize CR or CRLF -> LF
			next, _ := br.Peek(1)
			if len(next) == 1 && next[0] == '\n' {
				_, _ = br.Discard(1)
			}
			rn = '\n'
		}

		if rn == '\t' || rn == '\n' {
			out.WriteRune(rn)
			continue
		}

		if rn < 0x20 {
			// silently drop other ASCII controls
			continue
		}

		if !utf8.ValidRune(rn) || rn == utf8.RuneError {
			return nil, lang.ErrInvalidUTF8
		}

		// optionally: ban bidi overrides / zero-width; comment out to allow
		switch rn {
		case '\u200B', '\u200C', '\u200D', '\u2060', '\u202A', '\u202B',
			'\u202C', '\u202D', '\u202E':
			return nil, fmt.Errorf("disallowed invisible control U+%04X", rn)
		}

		out.WriteRune(rn)

		if out.Len() > maxBytes {
			return nil, fmt.Errorf("script exceeds %d bytes after cleaning: %w", maxBytes, lang.ErrResourceExhaustion)
		}
	}

	return out.Bytes(), nil
}
