package scanner2_test

import (
	"io"
	"testing"

	"github.com/eriklott/mustache/internal/scanner2"
	"github.com/google/go-cmp/cmp"
)

type token struct {
	Type   scanner2.TokenType
	Value  string
	Offset int
	Column int
	Line   int
}

func TestScanner_Scan(t *testing.T) {
	tt := []struct {
		name   string
		src    string
		tokens []token
		err    string
	}{
		{
			name:   "eof",
			src:    "",
			tokens: []token{},
		},
		{
			name: "whitespace",
			src:  "     ",
			tokens: []token{
				{scanner2.WS, "     ", 0, 1, 1},
			},
		},
		{
			name: "text",
			src:  "abcde",
			tokens: []token{
				{scanner2.TEXT, "abcde", 0, 1, 1},
			},
		},
		{
			name: "padded text",
			src:  "  abcde",
			tokens: []token{
				{scanner2.TEXT, "  abcde", 0, 1, 1},
			},
		},
		{
			name: "newline",
			src:  "\n",
			tokens: []token{
				{scanner2.NEWLINE, "\n", 0, 1, 1},
			},
		},
		{
			name: "carriage return",
			src:  "\r\n",
			tokens: []token{
				{scanner2.NEWLINE, "\r\n", 0, 1, 1},
			},
		},
		{
			name: "mixed text",
			src:  "  \n abcd \r\n  ",
			tokens: []token{
				{scanner2.WS, "  ", 0, 1, 1},
				{scanner2.NEWLINE, "\n", 2, 3, 1},
				{scanner2.TEXT, " abcd ", 3, 1, 2},
				{scanner2.NEWLINE, "\r\n", 9, 7, 2},
				{scanner2.WS, "  ", 11, 1, 3},
			},
		},
		{
			name: "tag/empty",
			src:  "{{}}",
			err:  "main:1:1: unclosed tag",
		},
		{
			name: "tag/ident",
			src:  "{{key}}",
			tokens: []token{
				{scanner2.VARIABLE, "key", 0, 1, 1},
			},
		},
		{
			name: "tag/whitespace",
			src:  "{{ key }}",
			tokens: []token{
				{scanner2.VARIABLE, "key", 0, 1, 1},
			},
		},
		{
			name: "tag/dot",
			src:  "{{ key1.key2 }}",
			tokens: []token{
				{scanner2.VARIABLE, "key1.key2", 0, 1, 1},
			},
		},
		{
			name: "tag/withSurroundingText",
			src:  "a{{b}}c",
			tokens: []token{
				{scanner2.TEXT, "a", 0, 1, 1},
				{scanner2.VARIABLE, "b", 1, 2, 1},
				{scanner2.TEXT, "c", 6, 7, 1},
			},
		},
		{
			name: "tag/unescaped",
			src:  "{{{a}}}",
			tokens: []token{
				{scanner2.UNESCAPED_VARIABLE, "a", 0, 1, 1},
			},
		},

		{
			name: "tag/unescapeSymbol",
			src:  "{{&a}}",
			tokens: []token{
				{scanner2.UNESCAPED_VARIABLE, "a", 0, 1, 1},
			},
		},
		{
			name: "tag/section",
			src:  "{{#a}}",
			tokens: []token{
				{scanner2.SECTION, "a", 0, 1, 1},
			},
		},
		{
			name: "tag/invertedSection",
			src:  "{{^a}}",
			tokens: []token{
				{scanner2.INVERTED_SECTION, "a", 0, 1, 1},
			},
		},
		{
			name: "tag/sectionEnd",
			src:  "{{/a}}",
			tokens: []token{
				{scanner2.SECTION_END, "a", 0, 1, 1},
			},
		},
		{
			name:   "tag/comment",
			src:    "{{! comment }}",
			tokens: []token{},
		},
		{
			name: "tag/setDelims",
			src:  "{{=<< >>=}}<<a>>",
			tokens: []token{
				{scanner2.VARIABLE, "a", 11, 12, 1},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			scanner := scanner2.New("main", tc.src)
			tokens := []token{}
			for {
				tok, err := scanner.Scan()
				if err == io.EOF {
					break
				}

				errMsg := ""
				if err != nil {
					errMsg = err.Error()
				}

				if errMsg != tc.err {
					t.Fatalf("unexpected error, got:%v, want:%v", err, tc.err)
				}
				if err != nil || tc.err != "" {
					return
				}

				tokens = append(tokens, token{
					Type:   tok.Type,
					Value:  tok.Value,
					Offset: tok.Position.Offset,
					Column: tok.Position.Column,
					Line:   tok.Position.Line,
				})

			}
			if diff := cmp.Diff(tc.tokens, tokens); diff != "" {
				t.Errorf("tokens mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkScanner_Scan(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		scanner := scanner2.New("main", `
			abcdefg
			{{key}}
			{{{key}}}
			{{#key}}
			{{! 
				This is a multline comment 
			}}
			{{=<< >>=}}
			<<^key>>
			<</key>>
			<<>key>>
			<<&key>>
		`)
		for {
			_, err := scanner.Scan()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
