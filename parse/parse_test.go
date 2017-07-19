// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parse

import (
	"fmt"
	"testing"

	"github.com/eriklott/mustache/token"
)

type parseTest struct {
	name     string
	input    string
	expected string
	err      bool
}

var parseTests = []parseTest{
	// passing tests
	{"empty", "", "", false},
	{"text", "Hello World!", "Hello World!", false},
	{"variable tag", "{{hello}}", "{{hello}}", false},
	{"unescaped variable tag with delims", "{{{hello}}}", "{{&hello}}", false},
	{"unescaped variable tag with symbol", "{{& hello}}", "{{&hello}}", false},
	{"partial tag", "{{>var}}", "{{>var}}", false},
	{"comment tag", "{{!var}}", "{{!var}}", false},
	{"section tag", "A{{#key}}B{{/key}}C", "A{{#key}}B{{/key}}C", false},
	{"inverted section tag", "A{{^key}}B{{/key}}C", "A{{^key}}B{{/key}}C", false},
	{"nested section tag", "A{{#key}}B{{#key2}}C{{/key2}}D{{/key}}E", "A{{#key}}B{{#key2}}C{{/key2}}D{{/key}}E", false},
	{"set delims", "{{A}}{{=<% %>=}}<%B%><%={{ }}=%>{{C}}", "{{A}}{{=<% %>=}}<%B%><%={{ }}=%>{{C}}", false},
	{"identity", "{{var}}", "{{var}}", false},
	{"identity with periods", "{{var.var}}", "{{var.var}}", false},

	// standalone line tests
	{"empty", "", "", false},
	{"text", "Hello World!", "Hello World!", false},
	{"standalone line", "abc\n{{!var}}\ndef", "abc\n{{!var}}def", false},
	{"standalone line padding", "abc\n    {{!var}}   \ndef", "abc\n{{!var}}def", false},
	{"standalone line end", "abc\n    {{!var}}   ", "abc\n{{!var}}", false},
	{"standalone line start", "    {{!var}}   \ndef", "{{!var}}def", false},
	{"standalone line carriage return", "abc\r\n{{!var}}\r\ndef", "abc\r\n{{!var}}def", false},

	// error tests
	{"empty tag", "{{}}", "", true},
	{"unclosed tag", "{{", "", true},
	{"unclosed escape delim", "{{{}}", "", true},
	{"unexpected escape delim", "{{}}}", "", true},
	{"illegal value", "{{*}}", "", true},
	{"leading period", "{{.var}}", "", true},
	{"trailing period", "{{var.}}", "", true},
	{"spaces period", "{{var . var}}", "", true},
	{"dubble ident", "{{var var}}", "", true},
	{"unclosed section", "{{#var}}", "", true},
	{"unopened section", "{{/var}}", "", true},
	{"unclosed comment", "{{!var", "", true},
	{"unclosed set delims", "{{=<% %>=", "", true},
	{"invalid set delims", "{{=<%%>=}}", "", true},
}

func TestParser_Parse(t *testing.T) {
	for _, test := range parseTests {
		reader := token.NewScanner("", test.input, token.DefaultLeftDelim, token.DefaultRightDelim)
		tree, err := Parse(reader)

		if test.err && err == nil {
			t.Errorf("%s: expected error, got none", test.name)
		}
		if test.err {
			continue
		}

		if err != nil {
			t.Errorf("%s: %s", test.name, err)
			continue
		}

		got := fmt.Sprintf("%s", tree)
		if got != test.expected {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, got, test.expected)
		}
	}
}
