package token2

import "strings"

const (
	defaultLeftDelim  = "{{"
	defaultRightDelim = "}}"
)

type Scanner struct {
	src        string // the current string being read
	pos        int    // the current position
	column     int    // the current column number
	line       int    // the current line number
	leftDelim  string
	rightDelim string
}

func NewScanner(src string) *Scanner {
	s := &Scanner{
		src:        src,
		pos:        0,
		column:     0,
		line:       0,
		leftDelim:  defaultLeftDelim,
		rightDelim: defaultRightDelim,
	}
	return s
}

func (s *Scanner) Next() (Token, string, error) {
	// Scan tag
	startPos := s.pos
	if strings.HasPrefix(s.src[s.pos:], s.leftDelim) {
		sym := s.src[s.pos:]
	}
}
