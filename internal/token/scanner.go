package token

import (
	"io"
	"strings"
)

type state int

const (
	stateText state = iota
	stateTagSymbol
	stateTagInside
	stateCommentTagInside
)

const (
	defaultLeftDelim  = "{{"
	defaultRightDelim = "}}"
)

type Scanner struct {
	src    string
	pos    int
	column int
	line   int

	startPos    int
	startColumn int
	startLine   int

	leftDelim      string
	rightDelim     string
	nextLeftDelim  string
	nextRightDelim string

	state state
}

func NewScanner(src string) *Scanner {
	s := &Scanner{
		src:            src,
		pos:            0,
		column:         0,
		line:           1,
		startPos:       0,
		startColumn:    0,
		startLine:      1,
		leftDelim:      defaultLeftDelim,
		rightDelim:     defaultRightDelim,
		nextLeftDelim:  defaultLeftDelim,
		nextRightDelim: defaultRightDelim,
		state:          stateText,
	}
	return s
}

// read returns the next byte in the source. Returns io.EOF at end of string
func (s *Scanner) read() (byte, error) {
	if s.pos >= len(s.src) {
		return 0, io.EOF
	}
	b := s.src[s.pos]
	s.pos++
	s.column++
	if b == '\n' {
		s.column = 0
		s.line++
	}
	return b, nil
}

// peek return the next byte in the source without advancing the position.
func (s *Scanner) peek() (byte, error) {
	if s.pos >= len(s.src) {
		return 0, io.EOF
	}
	return s.src[s.pos], nil
}

func (s *Scanner) match(str string) bool {
	return strings.HasPrefix(s.src[s.pos:], str)
}

func (s *Scanner) accept(str string) bool {
	if !s.match(str) {
		return false
	}
	for i := 0; i < len(str); i++ {
		s.read()
	}
	return true
}

func (s *Scanner) resetSelection() {
	s.startPos = s.pos
	s.startColumn = s.column
	s.startLine = s.line
}

func (s *Scanner) Text() string {
	return s.src[s.startPos:s.pos]
}

func (s *Scanner) Position() Position {
	p := Position{
		Offset: s.startPos,
		Column: s.startColumn,
		Line:   s.startLine,
	}
	return p
}

func (s *Scanner) Next() Token {
	s.resetSelection()
	var token Token
	switch s.state {
	case stateText:
		token, s.state = s.scanText()
	case stateTagSymbol:
		token, s.state = s.scanTagSymbol()
	}
	return token
}

func (s *Scanner) scanText() (Token, state) {
	// read left delim tag if we're next to it.
	if s.accept(s.leftDelim) {
		if s.accept("{") {
			return ULDELIM, stateTagSymbol
		}
		return LDELIM, stateTagSymbol
	}

	// check for end of line
	if s.accept("\n") || s.accept("\r\n") {
		return EOL, stateText
	}

	b, err := s.read()

	// check for end of file
	if err != nil {
		return EOF, stateText
	}

	tok := WS

	for {
		if !isSpace(b) {
			tok = TEXT
		}

		_, err := s.peek()

		if err != nil || s.match("\n") || s.match("\r\n") || s.match(s.leftDelim) {
			return tok, stateText
		}

		b, _ = s.read()
	}
}

func (s *Scanner) scanTagSymbol() (Token, state) {
	// read right delim tag if we're next to it.
	if s.accept("}" + s.rightDelim) {
		return URDELIM, stateText
	}
	if s.accept(s.rightDelim) {
		return RDELIM, stateText
	}

	b, err := s.peek()
	if err != nil {
		return EOF, stateText
	}
	switch b {
	case '#':
		s.read()
		return SECTION, stateTagInside
	case '^':
		s.read()
		return ISECTION, stateTagInside
	case '/':
		s.read()
		return SECTIONEND, stateTagInside
	case '&':
		s.read()
		return UNESC, stateTagInside
	case '!':
		s.read()
		return COMMENT, stateCommentTagInside
	case '=':
		s.read()
		return SETDELIM, stateCommentTagInside
	}

}

// package token

// type state int

// const (
// 	stateText state = iota
// 	stateTagSymbol
// )

// const (
// 	defaultLeftDelim  = "{{"
// 	defaultRightDelim = "}}"
// )

// type Scanner struct {
// 	reader     *reader
// 	leftDelim  string
// 	rightDelim string
// 	state      state
// }

// func NewScanner(src string) *Scanner {
// 	s := &Scanner{
// 		reader:     newReader(src),
// 		leftDelim:  defaultLeftDelim,
// 		rightDelim: defaultRightDelim,
// 		state:      stateText,
// 	}
// 	return s
// }

// func (s *Scanner) Next() Token {
// 	s.reader.resetText()
// 	var token Token
// 	switch s.state {
// 	case stateText:
// 		token, s.state = s.scanText()
// 	case stateTagSymbol:
// 		token, s.state = s.scanTagSymbol()
// 	}
// 	return token
// }

// func (s *Scanner) Text() string {
// 	return s.reader.text()
// }

// func (s *Scanner) Pos() Position {
// 	return s.reader.position()
// }

// func (s *Scanner) scanText() (Token, state) {
// 	// read left delim tag if we're next to it.
// 	if s.reader.consumeString(s.leftDelim) {
// 		if s.reader.consumeString("{") {
// 			return ULDELIM, stateTagSymbol
// 		}
// 		return LDELIM, stateTagSymbol
// 	}

// 	// check for end of line
// 	if s.reader.consumeString("\n") || s.reader.consumeString("\r\n") {
// 		return EOL, stateText
// 	}

// 	ch := s.reader.readRune()

// 	// check for end of file
// 	if ch == eof {
// 		return EOF, stateText
// 	}

// 	tok := WS

// 	for {
// 		if !isSpace(ch) {
// 			tok = TEXT
// 		}

// 		if s.reader.matchString(s.leftDelim) {
// 			return tok, stateText
// 		}

// 		ch = s.reader.peekRune()

// 		if ch == eof {
// 			return tok, stateText
// 		}

// 		if isEndOfLine(ch) {
// 			return tok, stateText
// 		}

// 		ch = s.reader.readRune()
// 	}

// }

// func (s *Scanner) scanTagSymbol() (Token, state) {
// 	// consume unexpected right delim if we're next to it
// 	if s.reader.consumeString("}" + s.rightDelim) {
// 		return URDELIM, stateText
// 	}
// 	if s.reader.consumeString(s.rightDelim) {
// 		return RDELIM, stateText
// 	}
// 	return RDELIM, stateText
// 	// ch := s.reader.readRune()
// 	// switch ch {
// 	// case eof:
// 	// 	return EOF, stateText
// 	// case '!':
// 	// 	return COMMENT, stateTag
// 	// case '&':
// 	// 	return UNESC, stateTag
// 	// case '#':
// 	// 	return SECTION, stateTag
// 	// case '^'

// 	// }

// }

// func isEndOfLine(ch rune) bool {
// 	return ch == '\r' || ch == '\n'
// }

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}
