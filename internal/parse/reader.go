// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parse

import (
	"strings"
	"unicode/utf8"
)

const eof rune = -1

type runeReader interface {
	read() rune
	peek() rune
	hasPrefix(string) bool
	text() string
	reset()
	line() int
	column() int
}

type stringRuneReader struct {
	src   string
	start int // the tail position of the current selection
	pos   int // the current position
	ln    int // the current line number
	col   int // the current column number
}

func newRuneReader(src string) *stringRuneReader {
	return &stringRuneReader{
		src:   src,
		start: 0,
		pos:   0,
		ln:    1,
		col:   0,
	}
}

func (s *stringRuneReader) read() rune {
	if s.pos >= len(s.src) {
		return eof
	}
	r, w := utf8.DecodeRuneInString(s.src[s.pos:])
	s.pos += w
	s.col += w
	if r == '\n' {
		s.col = 0
		s.ln++
	}
	return r
}

func (s *stringRuneReader) peek() rune {
	if s.pos >= len(s.src) {
		return eof
	}
	r, _ := utf8.DecodeRuneInString(s.src[s.pos:])
	return r
}

func (s *stringRuneReader) hasPrefix(prefix string) bool {
	return strings.HasPrefix(s.src[s.pos:], prefix)
}

func (s *stringRuneReader) text() string {
	return s.src[s.start:s.pos]
}

func (s *stringRuneReader) reset() {
	s.start = s.pos
}

func (s *stringRuneReader) line() int {
	return s.ln
}

func (s *stringRuneReader) column() int {
	return s.col
}
