package token2

import "strconv"

type Token int

const (
	TEXT       Token = iota // Text content between tags
	WS                      // Whitespace
	EOL                     // Newline char \n, \n\r, \r
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
	EOL:        "EOL",
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
