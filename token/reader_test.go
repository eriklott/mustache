// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package token

import (
	"fmt"
	"strings"
	"testing"
)

type tl struct {
	tok Token
	lit string
}

func (t tl) String() string {
	return fmt.Sprintf("%s: %s", t.tok, t.lit)
}

var scanTests = []struct {
	name   string
	src    string
	tokens []tl
}{
	{"empty", "", []tl{
		tl{EOF, ""},
	}},
	{"null text", "      ", []tl{
		tl{PAD, "      "},
		tl{EOF, ""},
	}},
	{"text", "   abc   ", []tl{
		tl{TEXT, "   abc   "},
		tl{EOF, ""},
	}},
	{"new line", "\n", []tl{
		tl{EOL, "\n"},
		tl{EOF, ""},
	}},
	{"carriage return new line", "\r\n", []tl{
		tl{EOL, "\r\n"},
		tl{EOF, ""},
	}},
	{"new line and whitespace", " \n ", []tl{
		tl{PAD, " "},
		tl{EOL, "\n"},
		tl{PAD, " "},
		tl{EOF, ""},
	}},
	{"various text", "    hello  \n   goodbye   ", []tl{
		tl{TEXT, "    hello  "},
		tl{EOL, "\n"},
		tl{TEXT, "   goodbye   "},
		tl{EOF, ""},
	}},
	{"tag only", "{{}}", []tl{
		tl{LDELIM, "{{"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"alternating text and tag", "A{{}}B", []tl{
		tl{TEXT, "A"},
		tl{LDELIM, "{{"},
		tl{RDELIM, "}}"},
		tl{TEXT, "B"},
		tl{EOF, ""},
	}},
	{"var tag", "{{id}}", []tl{
		tl{LDELIM, "{{"},
		tl{KEY, "id"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"var tag whitespace", "{{ id }}", []tl{
		tl{LDELIM, "{{"},
		tl{WS, " "},
		tl{KEY, "id"},
		tl{WS, " "},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"var period", "{{.}}", []tl{
		tl{LDELIM, "{{"},
		tl{DOT, "."},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"var period", "{{person.name}}", []tl{
		tl{LDELIM, "{{"},
		tl{KEY, "person"},
		tl{DOT, "."},
		tl{KEY, "name"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"unescaped tag", "{{{id}}}", []tl{
		tl{LDELIM, "{{"},
		tl{LUNESC, "{"},
		tl{KEY, "id"},
		tl{RUNESC, "}"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"comment tag", "{{!abc}}", []tl{
		tl{LDELIM, "{{"},
		tl{COMMENT, "!"},
		tl{TEXT, "abc"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"unclosed comment tag", "{{!abc", []tl{
		tl{LDELIM, "{{"},
		tl{COMMENT, "!"},
		tl{TEXT, "abc"},
		tl{EOF, ""},
	}},
	{"unescape symbol", "{{&}}", []tl{
		tl{LDELIM, "{{"},
		tl{UNESC, "&"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"section symbol", "{{#}}", []tl{
		tl{LDELIM, "{{"},
		tl{SECTION, "#"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"inverted section symbol", "{{^}}", []tl{
		tl{LDELIM, "{{"},
		tl{ISECTION, "^"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"section end symbol", "{{/}}", []tl{
		tl{LDELIM, "{{"},
		tl{SECTIONEND, "/"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"partial symbol", "{{>}}", []tl{
		tl{LDELIM, "{{"},
		tl{PARTIAL, ">"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"illegal", "{{~}}", []tl{
		tl{LDELIM, "{{"},
		tl{ILLEGAL, "~"},
		tl{RDELIM, "}}"},
		tl{EOF, ""},
	}},
	{"set delims", "{{A}}{{=<% %>=}}<%B%>", []tl{
		tl{LDELIM, "{{"},
		tl{KEY, "A"},
		tl{RDELIM, "}}"},
		tl{LDELIM, "{{"},
		tl{SETDELIM, "="},
		tl{TEXT, "<% %>"},
		tl{SETDELIM, "="},
		tl{RDELIM, "}}"},
		tl{LDELIM, "<%"},
		tl{KEY, "B"},
		tl{RDELIM, "%>"},
		tl{EOF, ""},
	}},
	{"set delims remove padding", "{{= | | =}}|B|", []tl{
		tl{LDELIM, "{{"},
		tl{SETDELIM, "="},
		tl{TEXT, " | | "},
		tl{SETDELIM, "="},
		tl{RDELIM, "}}"},
		tl{LDELIM, "|"},
		tl{KEY, "B"},
		tl{RDELIM, "|"},
		tl{EOF, ""},
	}},
}

func equal(i1, i2 []tl) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].tok != i2[k].tok {
			return false
		}
		if i1[k].lit != i2[k].lit {
			return false
		}
	}
	return true
}

func TestReader_Next(t *testing.T) {
	for _, test := range scanTests {

		l := NewReader("", strings.NewReader(test.src), DefaultLeftDelim, DefaultRightDelim)
		tokens := []tl{}
		for {
			tok, err := l.Next()
			if err != nil {
				t.Error(err)
				continue
			}
			tokens = append(tokens, tl{tok, l.Text()})
			if tok == EOF {
				break
			}
		}

		if !equal(tokens, test.tokens) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}
