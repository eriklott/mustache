package token2

import (
	"io"
	"strings"
	"unicode/utf8"
)

type reader interface {
	read() (rune, error)
	peek() (rune, error)
	match(string) bool
	accept(string) bool
	text() string
	reset()
	position() Position
}

type stringReader struct {
	src   string // the current string being read
	start int    // the tail position of the current selection
	pos   int    // the current position
	col   int    // the current column number
	ln    int    // the current line number
}

func newStringReader(src string) *stringReader {
	return &stringReader{
		src:   src,
		start: 0,
		pos:   0,
		col:   0,
		ln:    1,
	}
}

func (s *stringReader) read() (rune, error) {
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

func (s *stringReader) peek() (rune, error) {
	if s.pos >= len(s.src) {
		return rune(-1), io.EOF
	}
	r, _ := utf8.DecodeRuneInString(s.src[s.pos:])
	return r, nil
}

func (s *stringReader) match(pattern string) bool {
	return strings.HasPrefix(s.src[s.pos:], pattern)
}

func (s *stringReader) accept(pattern string) bool {
	if !s.match(pattern) {
		return false
	}

	l := len(pattern)
	for i := 0; i < l; i++ {
		b := s.src[s.pos]
		s.pos++
		s.col++
		if b == '\n' {
			s.col = 0
			s.ln++
		}
	}

	return true
}

func (s *stringReader) text() string {
	return s.src[s.start:s.pos]
}

func (s *stringReader) reset() {
	s.start = s.pos
}

func (s *stringReader) position() Position {
	return Position{
		Offset: s.pos,
		Column: s.col,
		Line:   s.ln,
	}
}
