package token3

import "strconv"

type Token int

// The result of Scan is one of these tokens.
const (
	// Special
	EOF     Token = iota // End of file
	ILLEGAL              // Illegal character

	// Literals
	IDENTITY   // A mustache tag identifier
	TEXT       // General text content found outside of tags
	WHITESPACE // Whitespace or tab
	NEWLINE    // \n, \n\r, \r

	// Symbols
	LDELIM                 // {{
	LDELIM_UNESCAPED       // {{{
	LDELIM_UNESCAPED_SYM   // {{&
	LDELIM_SECTION         // {{#
	LDELIM_INVERSE_SECTION // {{^
	LDELIM_SECTION_END     // {{/
	RDELIM                 // }}
	RDELIM_UNESCAPED       // }}}

)

var tokenNames = map[Token]string{
	EOF:                  "EOF",
	ILLEGAL:              "ILLEGAL",
	IDENTITY:             "IDENTITY",
	TEXT:                 "TEXT",
	WHITESPACE:           "WHITESPACE",
	NEWLINE:              "NEWLINE",
	LDELIM:               "LDELIM",
	LDELIM_UNESCAPED:     "LDELIM_UNESCAPED",
	LDELIM_UNESCAPED_SYM: "LDELIM_UNESCAPED_SYM",
	LDELIM_SECTION:       "LDELIM_SECTION",
	LDELIM_SECTION_END:   "LDELIM_SECTION_END",
	RDELIM:               "RDELIM",
	RDELIM_UNESCAPED:     "RDELIM_UNESCAPED",
}

func (t Token) String() string {
	s, ok := tokenNames[t]
	if !ok {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}
