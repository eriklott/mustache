package token3

// import (
// 	"strings"
// 	"unicode/utf8"
// )

// const eof rune = -1

// type reader struct {
// 	src   string // the current string being read
// 	start int    // the tail position of the current selection
// 	pos   int    // the current position
// 	col   int    // the current column number
// 	ln    int    // the current line number
// }

// func newReader(src string) *reader {
// 	return &reader{
// 		src:   src,
// 		start: 0,
// 		pos:   0,
// 		col:   0,
// 		ln:    1,
// 	}
// }

// func (s *reader) readRune() rune {
// 	if s.pos >= len(s.src) {
// 		return eof
// 	}
// 	r, w := utf8.DecodeRuneInString(s.src[s.pos:])
// 	s.pos += w
// 	s.col += w
// 	if r == '\n' {
// 		s.col = 0
// 		s.ln++
// 	}
// 	return r
// }

// func (s *reader) peekRune() rune {
// 	if s.pos >= len(s.src) {
// 		return eof
// 	}
// 	r, _ := utf8.DecodeRuneInString(s.src[s.pos:])
// 	return r
// }

// func (s *reader) consumeString(pattern string) bool {
// 	if !s.matchString(pattern) {
// 		return false
// 	}

// 	l := len(pattern)
// 	for i := 0; i < l; i++ {
// 		b := s.src[s.pos]
// 		s.pos++
// 		s.col++
// 		if b == '\n' {
// 			s.col = 0
// 			s.ln++
// 		}
// 	}

// 	return true
// }

// func (s *reader) matchString(pattern string) bool {
// 	return strings.HasPrefix(s.src[s.pos:], pattern)
// }

// func (s *reader) text() string {
// 	return s.src[s.start:s.pos]
// }

// func (s *reader) resetText() {
// 	s.start = s.pos
// }

// func (s *reader) position() Position {
// 	return Position{
// 		Offset: s.pos,
// 		Column: s.col,
// 		Line:   s.ln,
// 	}
// }
