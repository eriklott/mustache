// package token

// import "strconv"

// type Token int

// const (
// 	ILLEGAL    Token = iota // Illegal character
// 	EOF                     // End of file
// 	TEXT                    // Text content between tags
// 	WS                      // Whitespace
// 	EOL                     // Newline char \n, \n\r, \r
// 	VAR                     // Variable tag
// 	RAWVAR                  // Unescaped variable tag
// 	SECTION                 // Section tag
// 	ISECTION                // Inverted section tag
// 	SECTIONEND              // Section end tag
// 	PARTIAL                 // Partial tag

// )

// var tokenNames = map[Token]string{
// 	ILLEGAL:    "ILLEGAL",
// 	EOF:        "EOF",
// 	TEXT:       "TEXT",
// 	WS:         "WS",
// 	EOL:        "EOL",
// 	VAR:        "VAR",
// 	RAWVAR:     "RAWVAR",
// 	SECTION:    "SECTION",
// 	ISECTION:   "ISECTION",
// 	SECTIONEND: "SECTIONEND",
// 	PARTIAL:    "PARTIAL",
// }

// func (t Token) String() string {
// 	s, ok := tokenNames[t]
// 	if !ok {
// 		s = "token(" + strconv.Itoa(int(t)) + ")"
// 	}
// 	return s
// }

package token

import "strconv"

type Token int

const (
	// Special
	ILLEGAL Token = iota
	EOF

	// Literals
	KEY  // The tag identifier
	TEXT // Text content between tags
	WS   // Whitespace
	EOL  // Newline char \n, \n\r, \r

	// Symbols
	LDELIM     // {{
	RDELIM     // }}
	ULDELIM    // {{{
	URDELIM    // }}}
	SECTION    // #
	ISECTION   // ^
	SECTIONEND // /
	PARTIAL    // >
	SETDELIM   // =
	UNESC      // &
	COMMENT    // !
)

var tokenNames = map[Token]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	KEY:        "KEY",
	TEXT:       "TEXT",
	WS:         "WS",
	EOL:        "EOL",
	LDELIM:     "LDELIM",
	RDELIM:     "RDELIM",
	ULDELIM:    "ULDELIM",
	URDELIM:    "URDELIM",
	SECTION:    "#",
	ISECTION:   "^",
	SECTIONEND: "/",
	PARTIAL:    ">",
	SETDELIM:   "=",
	UNESC:      "&",
	COMMENT:    "!",
}

func (t Token) String() string {
	s, ok := tokenNames[t]
	if !ok {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}
