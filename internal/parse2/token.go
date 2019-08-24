package parse2

import "strconv"

type token int

const (
	tILLEGAL    token = iota // Illegal character
	tEOF                     // End of file
	tTEXT                    // Text content between tags
	tWS                      // Whitespace
	tEOL                     // Newline char \n, \n\r, \r
	tVAR                     // Variable tag
	tRAWVAR                  // Unescaped variable tag
	tSECTION                 // Section tag
	tISECTION                // Inverted section tag
	tSECTIONEND              // Section end tag
	tPARTIAL                 // Partial tag

)

var tokenNames = map[token]string{
	tILLEGAL:    "ILLEGAL",
	tEOF:        "EOF",
	tTEXT:       "TEXT",
	tWS:         "WS",
	tEOL:        "EOL",
	tVAR:        "VAR",
	tRAWVAR:     "RAWVAR",
	tSECTION:    "SECTION",
	tISECTION:   "ISECTION",
	tSECTIONEND: "SECTIONEND",
	tPARTIAL:    "PARTIAL",
}

func (t token) String() string {
	s, ok := tokenNames[t]
	if !ok {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}
