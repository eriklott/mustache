package scanner

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/eriklott/mustache/internal/token"
)

type state int

const (
	stateText state = iota
	stateTag
	stateCommentTag
	stateSetDelimsTag
)

const (
	defaultLeftDelim  = "{{"
	defaultRightDelim = "}}"
)

type Scanner struct {
	src string // the current string being read
	pos int    // the current position
	col int    // the current num of characters
	ln  int    // the current line number

	startPos int // the start position of the current selection
	startCol int // the start number of characters of the current selection
	startLn  int // the start line number of the current selection

	lDelim string // current left delimeter
	rDelim string // current right delimeter

	nextLDelim string // next left delimeter
	nextRDelim string // next right delimeter

	state state // the scanning state
}

func New(src string) *Scanner {
	s := &Scanner{
		src:        src,
		pos:        0,
		col:        0,
		ln:         1,
		startPos:   0,
		startCol:   0,
		startLn:    1,
		lDelim:     defaultLeftDelim,
		rDelim:     defaultRightDelim,
		nextLDelim: defaultLeftDelim,
		nextRDelim: defaultRightDelim,
		state:      stateText,
	}
	return s
}

// Advance one byte forward and return the byte. Returns an io.EOF error when at the end of the string.
func (s *Scanner) nextByte() (byte, error) {
	if s.pos >= len(s.src) {
		return 0, io.EOF
	}
	b := s.src[s.pos]
	s.pos++
	s.col++
	if b == '\n' {
		s.col = 0
		s.ln++
	}
	return b, nil
}

// Return the next byte without advancing. Returns an io.EOF error when at the end of the string.
func (s *Scanner) peekByte() (byte, error) {
	if s.pos >= len(s.src) {
		return 0, io.EOF
	}
	b := s.src[s.pos]
	return b, nil
}

// Advance one rune forward and return the rune. Returns an io.EOF error when at the end of the string.
func (s *Scanner) nextRune() (rune, error) {
	if s.pos >= len(s.src) {
		return 0, io.EOF
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

// Return the next rune without advancing. Returns an io.EOF error when at the end of the string.
func (s *Scanner) peekRune() (rune, error) {
	if s.pos >= len(s.src) {
		return 0, io.EOF
	}
	r, _ := utf8.DecodeRuneInString(s.src[s.pos:])
	return r, nil
}

// Check if the next bytes in the source equal the string.
func (s *Scanner) match(pattern string) bool {
	// return strings.HasPrefix(s.src[s.pos:], pattern)
	if s.pos+len(pattern) > len(s.src) {
		return false
	}
	for i := 0; i < len(pattern); i++ {
		if s.src[s.pos+i] != pattern[i] {
			return false
		}
	}
	return true

}

// Check if the next bytes in the source equal the string, and move foward if they do.
func (s *Scanner) accept(pattern string) bool {
	if !s.match(pattern) {
		return false
	}
	for i := 0; i < len(pattern); i++ {
		s.nextByte()
	}
	return true
}

// Reset the text selection
func (s *Scanner) reset() {
	s.startPos = s.pos
	s.startCol = s.col
	s.startLn = s.ln
}

// Returns the text value of the last scanned token.
func (s *Scanner) Text() string {
	return s.src[s.startPos:s.pos]
}

// Returns the starting position of the last scanned token.
func (s *Scanner) Position() token.Position {
	return token.Position{
		Offset: s.startPos,
		Column: s.startCol,
		Line:   s.startLn,
	}
}

func (s *Scanner) Scan() token.Token {
	s.reset()
	_, err := s.peekByte()
	if err != nil {
		return token.EOF
	}

	var t token.Token
	switch s.state {
	case stateText:
		t, s.state = s.scanText()
	case stateTag:
		t, s.state = s.scanTag()
	case stateCommentTag:
		t, s.state = s.scanCommentTag()
	case stateSetDelimsTag:
		t, s.state = s.scanSetDelimsTag()
	}
	return t
}

func (s *Scanner) scanText() (token.Token, state) {
	// parse left delim if we're beside it.
	if s.accept(s.lDelim) {
		ch, err := s.peekByte()
		if err != nil {
			return token.LDELIM, stateText
		}
		switch ch {
		case '{':
			s.nextByte()
			return token.LDELIM_UNESCAPED, stateTag
		case '&':
			s.nextByte()
			return token.LDELIM_UNESCAPED_SYM, stateTag
		case '#':
			s.nextByte()
			return token.LDELIM_SECTION, stateTag
		case '^':
			s.nextByte()
			return token.LDELIM_INVERSE_SECTION, stateTag
		case '/':
			s.nextByte()
			return token.LDELIM_SECTION_END, stateTag
		case '>':
			s.nextByte()
			return token.LDELIM_PARTIAL, stateTag
		case '!':
			s.nextByte()
			return token.LDELIM_COMMENT, stateCommentTag
		case '=':
			s.nextByte()
			return token.LDELIM_SETDELIM, stateSetDelimsTag
		default:
			return token.LDELIM, stateTag
		}
	}

	// scan first byte
	ch, err := s.nextByte()
	if err != nil {
		return token.EOF, stateText
	}

	// consume newline
	if ch == '\r' {
		ch2, err := s.peekByte()
		if err == nil && ch2 == '\n' {
			s.nextByte()
			return token.NEWLINE, stateText
		}
	}
	if ch == '\n' {
		return token.NEWLINE, stateText
	}

	tok := token.WS

	for {
		if !isSpace(ch) {
			tok = token.TEXT
		}

		// check if we're arrive at EOF
		ch, err = s.peekByte()
		if err != nil {
			return tok, stateText
		}

		// check if there is a new line ahead of us
		if isNewLine(ch) {
			return tok, stateText
		}

		// check if we're at the left delim
		if s.match(s.lDelim) {
			return tok, stateText
		}

		s.nextByte()
	}
}

func (s *Scanner) scanTag() (token.Token, state) {
	// skip whitespace and reset text selection (so that we don't capture any whitespace in current selection)
	for {
		b, _ := s.peekByte()
		if !isSpace(b) {
			s.reset()
			break
		}
		s.nextByte()
	}

	// scan right delim if we're next to it.
	if s.accept("}" + s.rDelim) {
		return token.RDELIM_UNESCAPED, stateText
	}
	if s.accept(s.rDelim) {
		return token.RDELIM, stateText
	}

	// scan
	switch ch, err := s.nextRune(); {
	case err != nil:
		return token.EOF, stateText
	case ch == '.':
		return token.DOT, stateTag
	case isIdent(ch):
		for {
			ch, _ = s.peekRune()
			if !isIdent(ch) {
				return token.IDENT, stateTag
			}
			ch, _ = s.nextRune()
		}
	default:
		return token.ILLEGAL, stateTag
	}
}

func (s *Scanner) scanCommentTag() (token.Token, state) {
	// scan right delim if we're next to it.
	if s.accept(s.rDelim) {
		return token.RDELIM, stateText
	}

	_, err := s.nextByte()

	if err != nil {
		return token.EOF, stateText
	}

	for {
		if s.match(s.rDelim) {
			return token.COMMENT, stateCommentTag
		}

		_, err = s.nextByte()
		if err != nil {
			return token.COMMENT, stateText
		}
	}
}

func (s *Scanner) scanSetDelimsTag() (token.Token, state) {
	rDelim := "=" + s.rDelim
	if s.accept(rDelim) {
		s.lDelim = s.nextLDelim
		s.rDelim = s.nextRDelim
		return token.RDELIM_SETDELIM, stateText
	}

	_, err := s.nextByte()
	if err != nil {
		return token.EOF, stateText
	}

	for {
		_, err := s.peekByte()
		if err != nil || s.match(rDelim) {
			delims := s.Text()
			parts := strings.Split(delims, " ")
			if len(parts) != 2 {
				return token.ILLEGAL, stateSetDelimsTag
			}
			s.nextLDelim = parts[0]
			s.nextRDelim = parts[1]
			return token.DELIMS, stateSetDelimsTag
		}

		s.nextByte()
	}
}

func isNewLine(ch byte) bool {
	return ch == '\r' || ch == '\n'
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

func isIdent(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}
