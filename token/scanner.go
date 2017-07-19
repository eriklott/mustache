// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package token

import (
	"strings"
	"unicode"
)

const (
	DefaultLeftDelim  = "{{"
	DefaultRightDelim = "}}"
)

// state represents the possible lexing states
type state int

const (
	stateText state = iota
	stateTag
	stateCommentTag
	stateSetDelimTag
)

type Scanner interface {
	// Next returns the next
	Next() Token
	// The following methods all refer to the most recent token returned by Next.
	// Text returns the original string representation of the
	Text() string
	// Pos reports the position or the most recent token.
	Pos() Position
}

func NewScanner(filename string, src string, lDelim, rDelim string) Scanner {
	return &scanner{
		r:          newStringSelectionReader(src),
		file:       filename,
		lDelim:     lDelim,
		rDelim:     rDelim,
		nextLDelim: lDelim,
		nextRDelim: rDelim,
		state:      stateText,
	}
}

type scanner struct {
	r          selectionReader
	file       string // name of the file being lexed, if any
	lDelim     string // left delim
	rDelim     string // right delim
	nextLDelim string // next left delim
	nextRDelim string // next right delim
	state      state  // current lexing state
}

// Text returns the literal string of the most recent returned token
func (s *scanner) Text() string {
	return s.r.selection()
}

// Pos reports the position or the most recent token.
func (s *scanner) Pos() Position {
	return Position{
		File: s.file,
		Line: s.r.line(),
		Col:  s.r.column(),
	}
}

// Next returns the next token in the src file.
func (s *scanner) Next() Token {
	s.r.resetSelection()
	token := ILLEGAL
	switch s.state {
	case stateText:
		token, s.state = s.lexText()
	case stateTag:
		token, s.state = s.lexTag()
	case stateCommentTag:
		token, s.state = s.lexCommentTag()
	case stateSetDelimTag:
		token, s.state = s.lexSetDelimTag()
	}
	return token
}

func (s *scanner) lexText() (Token, state) {
	// if at the left delim, consume it and switch to tag state
	if s.r.hasPrefix(s.lDelim) {
		s.r.readPrefix(s.lDelim)
		return LDELIM, stateTag
	}

	ch := s.r.readRune()

	// check for end of file
	if ch == eof {
		return EOF, stateText
	}

	// check for end of line
	if isEndOfLine(ch) {
		if ch == '\r' {
			ch2 := s.r.peekRune()
			if ch2 == '\n' {
				s.r.readRune()
			}
		}
		return EOL, stateText
	}

	t := PAD
	if !isSpace(ch) {
		t = TEXT
	}
	for {
		if s.r.hasPrefix(s.lDelim) {
			return t, stateText
		}

		ch2 := s.r.peekRune()

		if ch2 == eof {
			return t, stateText
		}

		if isEndOfLine(ch2) {
			return t, stateText
		}

		if !isSpace(ch2) {
			t = TEXT
		}

		s.r.readRune()
	}
}

func (s *scanner) lexTag() (Token, state) {
	if s.atRightDelim() {
		s.r.readPrefix(s.rDelim)
		return RDELIM, stateText
	}

	switch ch := s.r.readRune(); {
	case ch == eof:
		return EOF, stateText
	case isKey(ch):
		for {
			ch2 := s.r.peekRune()
			if !isKey(ch2) {
				return KEY, stateTag
			}
			s.r.readRune()
		}
	case isSpace(ch):
		for {
			ch2 := s.r.peekRune()
			if !isSpace(ch2) {
				return WS, stateTag
			}
			s.r.readRune()
		}
	case ch == '!':
		return COMMENT, stateCommentTag
	case ch == '=':
		return SETDELIM, stateSetDelimTag
	case ch == '.':
		return DOT, stateTag
	case ch == '^':
		return ISECTION, stateTag
	case ch == '>':
		return PARTIAL, stateTag
	case ch == '#':
		return SECTION, stateTag
	case ch == '/':
		return SECTIONEND, stateTag
	case ch == '&':
		return UNESC, stateTag
	case ch == '{':
		return LUNESC, stateTag
	case ch == '}':
		return RUNESC, stateTag
	default:
		return ILLEGAL, stateTag
	}
}

func (s *scanner) lexCommentTag() (Token, state) {

	if s.atRightDelim() {
		s.r.readPrefix(s.rDelim)
		return RDELIM, stateText
	}

	ch := s.r.readRune()
	if ch == eof {
		return EOF, stateText
	}

	for {
		if s.atRightDelim() {
			return TEXT, stateCommentTag
		}
		ch := s.r.readRune()
		if ch == eof {
			return TEXT, stateCommentTag
		}
	}
}

func (s *scanner) lexSetDelimTag() (Token, state) {
	if s.atRightDelim() {
		s.r.readPrefix(s.rDelim)
		s.lDelim = s.nextLDelim
		s.rDelim = s.nextRDelim
		return RDELIM, stateText
	}

	ch := s.r.readRune()

	if ch == eof {
		return EOF, stateText
	}

	if ch == '=' {
		return SETDELIM, stateSetDelimTag
	}

Loop:
	for {
		if s.atRightDelim() {
			break
		}
		switch ch := s.r.peekRune(); {
		case ch == eof:
			break Loop
		case ch == '=':
			break Loop
		default:
			s.r.readRune()
		}
	}

	rawDelims := s.r.selection()
	delims := strings.SplitN(strings.TrimSpace(rawDelims), " ", 2)
	if len(delims) != 2 {
		return ILLEGAL, stateSetDelimTag
	}
	s.nextLDelim = strings.TrimSpace(delims[0])
	s.nextRDelim = strings.TrimSpace(delims[1])
	return TEXT, stateSetDelimTag
}

// Hepers

func (s *scanner) atRightDelim() bool {
	unescapedRightDelim := "}" + s.rDelim
	if s.r.hasPrefix(unescapedRightDelim) {
		return false
	}
	return s.r.hasPrefix(s.rDelim)
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(ch rune) bool {
	return ch == '\r' || ch == '\n'
}

// isKey reports whether r is an alphabetic, digit, or underscore.
func isKey(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

// isSpace reports whether r is a space character.
func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}
