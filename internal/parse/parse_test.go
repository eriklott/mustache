// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parse_test

import (
	"io/ioutil"
	"testing"

	"github.com/eriklott/mustache/internal/parse"
)

// func TestParse(t *testing.T) {
// 	tt := []struct {
// 		name  string
// 		tmpl  string
// 		err   string
// 		nodes []ast.Node
// 	}{
// 		{
// 			name: "Text",
// 			tmpl: "abc",
// 			nodes: []ast.Node{
// 				&ast.Text{Text: "abc"},
// 			},
// 		},
// 		{
// 			name: "Newline",
// 			tmpl: "\n",
// 			nodes: []ast.Node{
// 				&ast.Text{Text: "\n", EndOfLine: true},
// 			},
// 		},
// 		{
// 			name: "Carriage return",
// 			tmpl: "\r\n",
// 			nodes: []ast.Node{
// 				&ast.Text{Text: "\r\n", EndOfLine: true},
// 			},
// 		},
// 		{
// 			name: "Variable",
// 			tmpl: "{{a}}",
// 			nodes: []ast.Node{
// 				&ast.Variable{
// 					Key:       []string{"a"},
// 					Unescaped: false,
// 					Line:      1,
// 					Column:    1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Variable/Whitespace",
// 			tmpl: "{{ a }}",
// 			nodes: []ast.Node{
// 				&ast.Variable{
// 					Key:       []string{"a"},
// 					Unescaped: false,
// 					Line:      1,
// 					Column:    1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Variable/Empty",
// 			tmpl: "{{}}",
// 			err:  "main:1:1: missing key",
// 		},
// 		{
// 			name: "Variable/Key/Dotted",
// 			tmpl: "{{a.b.c}}",
// 			nodes: []ast.Node{
// 				&ast.Variable{
// 					Key:       []string{"a", "b", "c"},
// 					Unescaped: false,
// 					Line:      1,
// 					Column:    1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Variable/Key/LeadingDot",
// 			tmpl: "{{.a.b}}",
// 			err:  "main:1:1: invalid key: .a.b",
// 		},
// 		{
// 			name: "Variable/Key/TrailingDot",
// 			tmpl: "{{a.b.}}",
// 			err:  "main:1:1: invalid key: a.b.",
// 		},
// 		{
// 			name: "Variable/Key/Whitespace",
// 			tmpl: "{{a . b}}",
// 			err:  "main:1:1: invalid key: a . b",
// 		},
// 		{
// 			name: "Variable/Key/SingleDot",
// 			tmpl: "{{.}}",
// 			nodes: []ast.Node{
// 				&ast.Variable{
// 					Key:       []string{"."},
// 					Unescaped: false,
// 					Line:      1,
// 					Column:    1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Variable/UnescapedSymbol",
// 			tmpl: "{{&a}}",
// 			nodes: []ast.Node{
// 				&ast.Variable{
// 					Key:       []string{"a"},
// 					Unescaped: true,
// 					Line:      1,
// 					Column:    1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Variable/Unescaped",
// 			tmpl: "{{{a}}}",
// 			nodes: []ast.Node{
// 				&ast.Variable{
// 					Key:       []string{"a"},
// 					Unescaped: true,
// 					Line:      1,
// 					Column:    1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Variable/Unescaped/Empty",
// 			tmpl: "{{{}}}",
// 			err:  "main:1:1: missing key",
// 		},
// 		{
// 			name: "Variable/Unescaped/MismatchDelims",
// 			tmpl: "{{{a}}",
// 			err:  "main:1:1: unclosed tag",
// 		},
// 		{
// 			name: "Section",
// 			tmpl: "{{#a}}{{/a}}",
// 			nodes: []ast.Node{
// 				&ast.Section{
// 					Key:    []string{"a"},
// 					LDelim: "{{",
// 					RDelim: "}}",
// 					Line:   1,
// 					Column: 1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Section/Inverted",
// 			tmpl: "{{^a}}{{/a}}",
// 			nodes: []ast.Node{
// 				&ast.Section{
// 					Key:      []string{"a"},
// 					Inverted: true,
// 					LDelim:   "{{",
// 					RDelim:   "}}",
// 					Line:     1,
// 					Column:   1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Section/MissingClosingTag",
// 			tmpl: "{{#a}}",
// 			err:  "main:1:1: unclosed section tag: a",
// 		},
// 		{
// 			name: "Section/MissingOpeningTag",
// 			tmpl: "{{/a}}",
// 			err:  "main:1:1: unexpected section closing tag: a",
// 		},
// 		{
// 			name: "Section/EmptyOpeningTag",
// 			tmpl: "{{#}}",
// 			err:  "main:1:1: missing key",
// 		},
// 		{
// 			name: "Section/EmptyClosingTag",
// 			tmpl: "{{/}}",
// 			err:  "main:1:1: missing key",
// 		},
// 		{
// 			name: "Section/Children",
// 			tmpl: "{{#a}}abc{{/a}}",
// 			nodes: []ast.Node{
// 				&ast.Section{
// 					Key:      []string{"a"},
// 					Inverted: false,
// 					LDelim:   "{{",
// 					RDelim:   "}}",
// 					Text:     "abc",
// 					Line:     1,
// 					Column:   1,
// 					Nodes: []ast.Node{
// 						&ast.Text{Text: "abc"},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "Partial",
// 			tmpl: "{{>a}}",
// 			nodes: []ast.Node{
// 				&ast.Partial{
// 					Key:    "a",
// 					Line:   1,
// 					Column: 1,
// 				},
// 			},
// 		},
// 		{
// 			name: "Partial/Empty",
// 			tmpl: "{{>}}",
// 			err:  "main:1:1: missing key",
// 		},
// 		{
// 			name:  "Comment/Skipped",
// 			tmpl:  "{{! This is a comment }}",
// 			nodes: nil,
// 		},
// 		{
// 			name:  "SetDelim",
// 			tmpl:  "{{=| |=}}",
// 			nodes: nil,
// 		},
// 		{
// 			name: "SetDelim/ChangesDelimeters",
// 			tmpl: "{{=| |=}}|a|",
// 			nodes: []ast.Node{
// 				&ast.Variable{
// 					Key:    []string{"a"},
// 					Line:   1,
// 					Column: 10,
// 				},
// 			},
// 		},
// 		{
// 			name: "Standalone/FirstLine",
// 			tmpl: " {{!a}} \nb",
// 			nodes: []ast.Node{
// 				&ast.Text{Text: "b"},
// 			},
// 		},
// 		{
// 			name: "Standalone/MidLine",
// 			tmpl: "\n {{!a}} \nb",
// 			nodes: []ast.Node{
// 				&ast.Text{Text: "\n", EndOfLine: true},
// 				&ast.Text{Text: "b", EndOfLine: false},
// 			},
// 		},
// 		{
// 			name: "Standalone/LastLine",
// 			tmpl: "a\n {{!b}} ",
// 			nodes: []ast.Node{
// 				&ast.Text{Text: "a\n", EndOfLine: true},
// 			},
// 		},
// 		{
// 			name: "Partial/StandaloneIndent",
// 			tmpl: "  {{>a}}  ",
// 			nodes: []ast.Node{
// 				&ast.Partial{
// 					Key:    "a",
// 					Indent: "  ",
// 					Line:   1,
// 					Column: 3,
// 				},
// 			},
// 		},
// 	}

// 	for _, tc := range tt {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tree, err := parse.Parse("main", tc.tmpl, parse.DefaultLeftDelim, parse.DefaultRightDelim)

// 			var errStr string
// 			if err != nil {
// 				errStr = err.Error()
// 			}
// 			if errStr != tc.err {
// 				t.Errorf("unexpected error, got: %s, want: %s", errStr, tc.err)
// 			}
// 			if err != nil || tc.err != "" {
// 				return
// 			}

// 			if !reflect.DeepEqual(tc.nodes, tree.Nodes) {
// 				t.Errorf("Parse() mismatch, got:%v, want:%v", tc.nodes, tree.Nodes)
// 			}
// 		})
// 	}
// }

func BenchmarkParse(b *testing.B) {
	tmplBytes, err := ioutil.ReadFile("../../testdata/template.mustache")
	if err != nil {
		b.Fatal(err)
	}
	tmpl := string(tmplBytes)

	for n := 0; n < b.N; n++ {
		_, err := parse.Parse("main", tmpl, parse.DefaultLeftDelim, parse.DefaultRightDelim)
		if err != nil {
			b.Fatal((err))
		}
	}
}
