package scanner

import (
	"fmt"
	"strconv"
)

// Position represents a location in the source string.
type Position struct {
	Offset int
	Column int
	Line   int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token int

const (
	TEXT       Token = iota // Text content between tags
	WS                      // Whitespace
	NEWLINE                 // Newline char \n, \n\r, \r
	VAR                     // Variable tag
	UVAR                    // Unescaped variable tag
	SECTION                 // Section tag
	ISECTION                // Inverted section tag
	SECTIONEND              // Section end tag
	PARTIAL                 // Partial tag
)

var tokenNames = map[Token]string{
	TEXT:       "TEXT",
	WS:         "WS",
	NEWLINE:    "NEWLINE",
	VAR:        "VAR",
	UVAR:       "UVAR",
	SECTION:    "SECTION",
	ISECTION:   "ISECTION",
	SECTIONEND: "SECTIONEND",
	PARTIAL:    "PARTIAL",
}

func (t Token) String() string {
	s, ok := tokenNames[t]
	if !ok {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

type Scanner struct {
	src
}

func (s *Scanner) Next() (Token, error) {}

func (s *Scanner) Text() string {}

func (s *Scanner) Pos() Position {}
