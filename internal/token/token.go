package token

import "strconv"

type Token int

// The list of tokens.
const (
	// Special tokens
	EOF     Token = iota // End of file
	ILLEGAL              // Illegal character

	// Literals
	IDENT   // A mustache tag identifier
	TEXT    // General text content found outside of tags
	WS      // Whitespace or tab
	NEWLINE // \n, \n\r, \r
	DOT     //.
	DELIMS  // A set of new delims e.g. << >> . Found inside the SetDelims tag
	COMMENT // The comment text found within a comment tag

	// Symbols
	LDELIM                 // {{
	LDELIM_UNESCAPED       // {{{
	LDELIM_UNESCAPED_SYM   // {{&
	LDELIM_SECTION         // {{#
	LDELIM_INVERSE_SECTION // {{^
	LDELIM_SECTION_END     // {{/
	LDELIM_PARTIAL         // {{>
	LDELIM_COMMENT         // {{!
	LDELIM_SETDELIM        // {{=
	RDELIM                 // }}
	RDELIM_UNESCAPED       // }}}
	RDELIM_SETDELIM        // =}}

)

var tokenNames = map[Token]string{
	EOF:                    "EOF",
	ILLEGAL:                "ILLEGAL",
	IDENT:                  "IDENT",
	TEXT:                   "TEXT",
	WS:                     "WS",
	NEWLINE:                "NEWLINE",
	DOT:                    "DOT",
	DELIMS:                 "DELIMS",
	COMMENT:                "COMMENT",
	LDELIM:                 "LDELIM",
	LDELIM_UNESCAPED:       "LDELIM_UNESCAPED",
	LDELIM_UNESCAPED_SYM:   "LDELIM_UNESCAPED_SYM",
	LDELIM_SECTION:         "LDELIM_SECTION",
	LDELIM_INVERSE_SECTION: "LDELIM_INVERSE_SECTION",
	LDELIM_SECTION_END:     "LDELIM_SECTION_END",
	LDELIM_PARTIAL:         "LDELIM_PARTIAL",
	LDELIM_COMMENT:         "LDELIM_COMMENT",
	LDELIM_SETDELIM:        "LDELIM_SETDELIM",
	RDELIM:                 "RDELIM",
	RDELIM_UNESCAPED:       "RDELIM_UNESCAPED",
	RDELIM_SETDELIM:        "RDELIM_SETDELIM",
}

func (t Token) String() string {
	s, ok := tokenNames[t]
	if !ok {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}
