// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package token_test

import (
	"io"
	"io/ioutil"
	"reflect"
	"testing"

	x "github.com/eriklott/mustache/internal/token"
)

type token struct {
	Type x.Type
	Text string
}

func TestScanner_Next(t *testing.T) {
	tt := []struct {
		name   string
		src    string
		tokens []token
		isErr  bool
	}{
		{"text", "abc", []token{{x.TEXT, "abc"}}, false},
		{"leading newline", "\n", []token{{x.TEXT_EOL, "\n"}}, false},
		{"leading carriage return", "\r\n", []token{{x.TEXT_EOL, "\r\n"}}, false},
		{"text & newline", "abc\n", []token{{x.TEXT_EOL, "abc\n"}}, false},
		{"text & carriage return", "abc\r\n", []token{{x.TEXT_EOL, "abc\r\n"}}, false},
		{"newline & text", "\nabc", []token{{x.TEXT_EOL, "\n"}, {x.TEXT, "abc"}}, false},
		{"variable tag", "{{ a }}", []token{{x.VARIABLE, "a"}}, false},
		{"unescaped variable tag", "{{{ a }}}", []token{{x.UNESCAPED_VARIABLE, "a"}}, false},
		{"unescaped variable symbole tag", "{{& a }}", []token{{x.UNESCAPED_VARIABLE_SYM, "a"}}, false},
		{"dotted variable tag", "{{ . }}", []token{{x.VARIABLE, "."}}, false},
		{"section tag", "{{# a }}", []token{{x.SECTION, "a"}}, false},
		{"inverted section tag", "{{^ a }}", []token{{x.INVERTED_SECTION, "a"}}, false},
		{"section end tag", "{{/ a }}", []token{{x.SECTION_END, "a"}}, false},
		{"partial tag", "{{> a }}", []token{{x.PARTIAL, "a"}}, false},
		{"comment tag", "{{! abc  }}", []token{{x.COMMENT, "abc"}}, false},
		{"set delims tag", "{{= | | =}}", []token{{x.SET_DELIMETERS, "| |"}}, false},
		{"tags", "{{a}}{{b}}", []token{{x.VARIABLE, "a"}, {x.VARIABLE, "b"}}, false},
		{"text & tag", "abc{{a}}", []token{{x.TEXT, "abc"}, {x.VARIABLE, "a"}}, false},
		{"change delimes", "{{a}}{{=| |=}}|b|", []token{{x.VARIABLE, "a"}, {x.SET_DELIMETERS, "| |"}, {x.VARIABLE, "b"}}, false},
		{"leading standalone", " {{#a}} \nabc", []token{{x.SECTION, "a"}, {x.TEXT, "abc"}}, false},
		{"mid standalone", "abc\n {{#a}} \ndef", []token{{x.TEXT_EOL, "abc\n"}, {x.SECTION, "a"}, {x.TEXT, "def"}}, false},
		{"trailing standalone", "abc\n {{#a}} ", []token{{x.TEXT_EOL, "abc\n"}, {x.SECTION, "a"}}, false},
		{"consecutive standalone", "{{#a}}\n{{#b}}\n{{#c}}\n", []token{{x.SECTION, "a"}, {x.SECTION, "b"}, {x.SECTION, "c"}}, false},

		// errors
		{"empty tag", "{{}}", nil, true},
		{"leading dot", "{{.a}}", nil, true},
		{"trailing dot", "{{a.}}", nil, true},
		{"dot whitespace", "{{a . b}}", nil, true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			scanner := x.NewScanner("main", tc.src, "{{", "}}")
			tokens := []token{}
			for {
				scannedToken, err := scanner.Next()
				if err == io.EOF {
					break
				}
				isErr := (err != nil)
				if isErr != tc.isErr {
					t.Fatalf("error mismatch, got %v, want: %v", isErr, tc.isErr)
				}
				if isErr || tc.isErr {
					return
				}

				tokn := token{
					Type: scannedToken.Type,
					Text: scannedToken.Text,
				}
				tokens = append(tokens, tokn)
			}

			if !reflect.DeepEqual(tc.tokens, tokens) {
				t.Errorf("unexpected tokens, got:%v, want:%v", tokens, tc.tokens)
			}
		})
	}
}

func BenchmarkScanner_Next(b *testing.B) {
	srcBytes, err := ioutil.ReadFile("../../testdata/template.mustache")
	if err != nil {
		b.Fatal(err)
	}
	src := string(srcBytes)
	scanner := x.NewScanner("main", src, "{{", "}}")

	for n := 0; n < b.N; n++ {
		for {
			_, err := scanner.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
