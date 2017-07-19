// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package token

import (
	"strings"
	"unicode/utf8"
)

const eof rune = -1

type selectionReader interface {
	readRune() rune
	peekRune() rune
	hasPrefix(string) bool
	readPrefix(string)
	selection() string
	resetSelection()
	line() int
	column() int
}

type stringSelectionReader struct {
	src   string
	start int // the tail position of the current selection
	pos   int // the current position
	col   int // the current column number
	ln    int // the current line number
}

func newStringSelectionReader(src string) *stringSelectionReader {
	return &stringSelectionReader{
		src: src,
		ln:  1,
	}
}

func (s *stringSelectionReader) readRune() rune {
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

func (s *stringSelectionReader) peekRune() rune {
	if s.pos >= len(s.src) {
		return eof
	}
	r, _ := utf8.DecodeRuneInString(s.src[s.pos:])
	return r
}

func (s *stringSelectionReader) hasPrefix(prefix string) bool {
	return strings.HasPrefix(s.src[s.pos:], prefix)
}

func (s *stringSelectionReader) readPrefix(prefix string) {
	if s.hasPrefix(prefix) {
		s.pos += len(prefix)
	}
}

func (s *stringSelectionReader) selection() string {
	return s.src[s.start:s.pos]
}

func (s *stringSelectionReader) resetSelection() {
	s.start = s.pos
}

func (s *stringSelectionReader) line() int {
	return s.ln
}

func (s *stringSelectionReader) column() int {
	return s.col
}
