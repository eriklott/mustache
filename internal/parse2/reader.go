package parse2

import (
	"io"
	"strings"
	"unicode/utf8"
)

// reader scans the source string, keeping track of the
// current scan position.
type reader interface {
	// readRune advances the position to the next rune, and returns the rune.
	// returns an io.EOF error at the end of document
	readRune() (rune, error)

	// peekRune returns the rune without advancing the position.
	// returns an io.EOF error at the end of document
	peekRune() (rune, error)

	// consumeNextString tests whether the next characters match the target string, and if so, reads
	// ahead that many characters.
	consumeNextString(target string) bool

	// hasNextString tests whether the next characters match the target string.
	hasNextString(target string) bool

	text()
	clearText()
}

type stringReader struct {
	src   string // the current string being read
	start int    // the tail position of the current selection
	pos   int    // the current position
	col   int    // the current column number
	ln    int    // the current line number
}

func (s *stringReader) readRune() (rune, error) {
	if s.pos >= len(s.src) {
		return rune(-1), io.EOF
	}
	r, w := utf8.DecodeRuneInString(s.src[s.pos:])
	s.pos += w
	s.col += w
	if r == '\n' {
		s.col = 0
		s.ln++
	}
	return r, nil
}

func (s *stringReader) peekRune() (rune, error) {
	if s.pos >= len(s.src) {
		return rune(-1), io.EOF
	}
	r, _ := utf8.DecodeRuneInString(s.src[s.pos:])
	return r, nil
}

func (s *stringReader) acceptString(pattern string) bool {
	if !strings.HasPrefix(s.src[s.pos:], pattern) {
		return false
	}

	for _, r := range pattern {
		w := utf8.RuneLen(r)
		s.pos += w
		s.col += w
		if r == '\n' {
			s.col = 0
			s.ln++
		}
	}
	return true
}
