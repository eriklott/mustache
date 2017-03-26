// Package token lexes mustache template strings, transforming them into a
// stream of tokens. Clients should use the mustache package to parse and
// render templates rather than this one, which provides shared internal data
// structures not intended for general use.
package token

import "strconv"

// Token is the set of lexical tokens of the Mustache template language.
type Token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF

	// Literals
	KEY  // The tag identifier
	TEXT // Text content between tags
	PAD  // Empty text content between tags
	EOL  // Newline char \n, \n\r, \r
	WS   // Whitespace

	// Tag Symbols
	LDELIM     // {{
	LUNESC     // {
	RDELIM     // }}
	RUNESC     // }
	SECTION    // #
	ISECTION   // ^
	SECTIONEND // /
	PARTIAL    // >
	SETDELIM   // =
	UNESC      // &
	COMMENT    // !
	DOT        // !
)

var tokens = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	KEY:        "KEY",
	TEXT:       "TEXT",
	PAD:        "PAD",
	EOL:        "EOL",
	WS:         "WS",
	LDELIM:     "LDELIM",
	RDELIM:     "RDELIM",
	LUNESC:     "{",
	RUNESC:     "}",
	SECTION:    "#",
	ISECTION:   "^",
	SECTIONEND: "/",
	PARTIAL:    ">",
	SETDELIM:   "=",
	UNESC:      "&",
	COMMENT:    "!",
	DOT:        ".",
}

// String returns the string corresponding to the token.
// For tag symbols, the string is the actual token character
// sequence (e.g., for the token Section, the string is
// "#"). For all other tokens, the string corresponds to the token
// constant name (e.g. for the token KEY, the string is "KEY").
func (t Token) String() string {
	s := ""
	if 0 <= t && t < Token(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}
