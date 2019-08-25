package scanner2

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Position struct {
	Name   string
	Offset int
	Column int
	Line   int
}

// String returns a string in one of several forms:
//
//	line:column    valid position with file name
//	line           valid position with file name but no column (column == 0)
//
func (pos Position) String() string {
	name := pos.Name
	if name == "" {
		name = "<input>"
	}

	return fmt.Sprintf("%s:%d:%d", name, pos.Line, pos.Column)
}

type TokenType int

// The list of token types.
const (
	TEXT TokenType = iota
	WS
	NEWLINE
	VARIABLE
	UNESCAPED_VARIABLE
	SECTION
	INVERTED_SECTION
	SECTION_END
	PARTIAL
)

var tokenTypes = map[TokenType]string{
	TEXT:               "TEXT",
	WS:                 "WS",
	NEWLINE:            "NEWLINE",
	VARIABLE:           "VARIABLE",
	UNESCAPED_VARIABLE: "UNESCAPED_VARIABLE",
	SECTION:            "SECTION",
	INVERTED_SECTION:   "INVERTED_SECTION",
	SECTION_END:        "SECTION_END",
	PARTIAL:            "PARTIAL",
}

func (t TokenType) String() string {
	s, ok := tokenTypes[t]
	if !ok {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

type Token struct {
	Type     TokenType
	Value    string
	Position Position
}

const (
	defaultLeftDelim  = "{{"
	defaultRightDelim = "}}"
)

type Scanner struct {
	name string // the name of the template being scanned
	src  string // the current string being read
	pos  int    // the current position
	col  int    // the current num of characters
	ln   int    // the current line number

	lDelim string // current left delimeter
	rDelim string // current right delimeter

	err error // the last error returned
}

func New(name, src string) *Scanner {
	s := &Scanner{
		name:   name,
		src:    src,
		pos:    0,
		col:    1,
		ln:     1,
		lDelim: defaultLeftDelim,
		rDelim: defaultRightDelim,
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
		s.col = 1
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

func (s *Scanner) nextUntil(pat string) error {
	for {
		if s.match(pat) {
			return nil
		}
		_, err := s.nextByte()
		if err != nil {
			return err
		}
	}
}

func (s *Scanner) nextUntilAfter(pat string) error {
	for {
		if s.match(pat) {
			break
		}
		_, err := s.nextByte()
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(pat); i++ {
		s.nextByte()
	}
	return nil
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

func (s *Scanner) Scan() (Token, error) {
	if s.err != nil {
		return Token{}, s.err
	}
	t, err := s.scan()
	if err != nil {
		s.err = err
	}
	return t, err
}

func (s *Scanner) scan() (Token, error) {
	start := Position{
		Name:   s.name,
		Offset: s.pos,
		Column: s.col,
		Line:   s.ln,
	}

	if s.accept(s.lDelim) {
		ch, err := s.nextByte()
		switch {
		case err != nil:
			return Token{}, fmt.Errorf("%s: unclosed tag", start)
		case ch == '{':
			err = s.nextUntilAfter("}" + s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed tag", start)
			}

			// get the ident
			ident := s.src[start.Offset+len(s.lDelim+"{") : s.pos-len("}"+s.rDelim)]
			ident = strings.TrimSpace(ident)
			if ident == "" {
				return Token{}, fmt.Errorf("%s: empty tag", start)
			}

			token := Token{
				Type:     UNESCAPED_VARIABLE,
				Value:    ident,
				Position: start,
			}
			return token, nil

		case ch == '&':
			err = s.nextUntilAfter(s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed tag", start)
			}

			// get the ident
			ident := s.src[start.Offset+len(s.lDelim+"&") : s.pos-len(s.rDelim)]
			ident = strings.TrimSpace(ident)
			if ident == "" {
				return Token{}, fmt.Errorf("%s: empty tag", start)
			}

			token := Token{
				Type:     UNESCAPED_VARIABLE,
				Value:    ident,
				Position: start,
			}
			return token, nil

		case ch == '#':
			err = s.nextUntilAfter(s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed tag", start)
			}

			// get the ident
			ident := s.src[start.Offset+len(s.lDelim+"#") : s.pos-len(s.rDelim)]
			ident = strings.TrimSpace(ident)
			if ident == "" {
				return Token{}, fmt.Errorf("%s: empty tag", start)
			}

			token := Token{
				Type:     SECTION,
				Value:    ident,
				Position: start,
			}
			return token, nil

		case ch == '^':
			err = s.nextUntilAfter(s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed tag", start)
			}

			// get the ident
			ident := s.src[start.Offset+len(s.lDelim+"^") : s.pos-len(s.rDelim)]
			ident = strings.TrimSpace(ident)
			if ident == "" {
				return Token{}, fmt.Errorf("%s: empty tag", start)
			}

			token := Token{
				Type:     INVERTED_SECTION,
				Value:    ident,
				Position: start,
			}
			return token, nil

		case ch == '/':
			err = s.nextUntilAfter(s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed tag", start)
			}

			// get the ident
			ident := s.src[start.Offset+len(s.lDelim+"/") : s.pos-len(s.rDelim)]
			ident = strings.TrimSpace(ident)
			if ident == "" {
				return Token{}, fmt.Errorf("%s: empty tag", start)
			}

			token := Token{
				Type:     SECTION_END,
				Value:    ident,
				Position: start,
			}
			return token, nil

		case ch == '>':
			err = s.nextUntilAfter(s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed tag", start)
			}

			// get the ident
			ident := s.src[start.Offset+len(s.lDelim+">") : s.pos-len(s.rDelim)]
			ident = strings.TrimSpace(ident)
			if ident == "" {
				return Token{}, fmt.Errorf("%s: empty tag", start)
			}

			token := Token{
				Type:     PARTIAL,
				Value:    ident,
				Position: start,
			}
			return token, nil

		case ch == '!':
			err = s.nextUntilAfter(s.rDelim)
			if err != nil {
				return Token{}, fmt.Errorf("%s: unclosed comment tag", start)
			}
			return s.Scan()

		case ch == '=':
			err = s.nextUntilAfter("=" + s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed set delims tag", start)
			}

			// get the ident
			delims := s.src[start.Offset+len(s.lDelim+"=") : s.pos-len("="+s.rDelim)]
			parts := strings.Split(delims, " ")
			if len(parts) != 2 {
				return Token{}, fmt.Errorf("%s: invalid set delim tag", start)
			}

			s.lDelim = parts[0]
			s.rDelim = parts[1]
			return s.Scan()
		default:
			err = s.nextUntilAfter(s.rDelim)
			if err != nil || s.ln != start.Line {
				return Token{}, fmt.Errorf("%s: unclosed tag", start)
			}

			// get the ident
			ident := s.src[start.Offset+len(s.lDelim) : s.pos-len(s.rDelim)]
			ident = strings.TrimSpace(ident)
			if ident == "" {
				return Token{}, fmt.Errorf("%s: empty tag", start)
			}

			token := Token{
				Type:     VARIABLE,
				Value:    ident,
				Position: start,
			}
			return token, nil
		}
	}

	// scan first byte
	ch, err := s.nextByte()
	if err != nil {
		return Token{}, io.EOF
	}

	// consume newline
	if ch == '\r' {
		ch2, err := s.peekByte()
		if err == nil && ch2 == '\n' {
			s.nextByte()
			return Token{Type: NEWLINE, Value: s.src[start.Offset:s.pos], Position: start}, nil

		}
	}
	if ch == '\n' {
		return Token{Type: NEWLINE, Value: s.src[start.Offset:s.pos], Position: start}, nil
	}

	typ := WS

	for {
		if !isSpace(ch) {
			typ = TEXT
		}

		// check if we're arrive at EOF
		ch, err = s.peekByte()
		if err != nil {
			return Token{Type: typ, Value: s.src[start.Offset:s.pos], Position: start}, nil
		}

		// check if there is a new line ahead of us
		if isNewLine(ch) {
			return Token{Type: typ, Value: s.src[start.Offset:s.pos], Position: start}, nil
		}

		// check if we're at the left delim
		if s.match(s.lDelim) {
			return Token{Type: typ, Value: s.src[start.Offset:s.pos], Position: start}, nil
		}

		ch, _ = s.nextByte()
	}
}

func isNewLine(ch byte) bool {
	return ch == '\r' || ch == '\n'
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}
