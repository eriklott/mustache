package scanner_test

import (
	"testing"

	"github.com/eriklott/mustache/internal/scanner"
	"github.com/eriklott/mustache/internal/token"
	"github.com/google/go-cmp/cmp"
)

type item struct {
	Token  token.Token
	Text   string
	Offset int
	Column int
	Line   int
}

func TestScanner_Scan(t *testing.T) {
	tt := []struct {
		name  string
		src   string
		items []item
	}{
		{
			name: "eof",
			src:  "",
			items: []item{
				{token.EOF, "", 0, 0, 1},
			},
		},
		{
			name: "whitespace",
			src:  "     ",
			items: []item{
				{token.WS, "     ", 0, 0, 1},
				{token.EOF, "", 5, 5, 1},
			},
		},
		{
			name: "text",
			src:  "abcde",
			items: []item{
				{token.TEXT, "abcde", 0, 0, 1},
				{token.EOF, "", 5, 5, 1},
			},
		},
		{
			name: "padded text",
			src:  "  abcde",
			items: []item{
				{token.TEXT, "  abcde", 0, 0, 1},
				{token.EOF, "", 7, 7, 1},
			},
		},
		{
			name: "newline",
			src:  "\n",
			items: []item{
				{token.NEWLINE, "\n", 0, 0, 1},
				{token.EOF, "", 1, 0, 2},
			},
		},
		{
			name: "carriage return",
			src:  "\r\n",
			items: []item{
				{token.NEWLINE, "\r\n", 0, 0, 1},
				{token.EOF, "", 2, 0, 2},
			},
		},
		{
			name: "mixed text",
			src:  "  \n abcd \r\n  ",
			items: []item{
				{token.WS, "  ", 0, 0, 1},
				{token.NEWLINE, "\n", 2, 2, 1},
				{token.TEXT, " abcd ", 3, 0, 2},
				{token.NEWLINE, "\r\n", 9, 6, 2},
				{token.WS, "  ", 11, 0, 3},
				{token.EOF, "", 13, 2, 3},
			},
		},
		{
			name: "tag/empty",
			src:  "{{}}",
			items: []item{
				{token.LDELIM, "{{", 0, 0, 1},
				{token.RDELIM, "}}", 2, 2, 1},
				{token.EOF, "", 4, 4, 1},
			},
		},
		{
			name: "tag/ident",
			src:  "{{key}}",
			items: []item{
				{token.LDELIM, "{{", 0, 0, 1},
				{token.IDENT, "key", 2, 2, 1},
				{token.RDELIM, "}}", 5, 5, 1},
				{token.EOF, "", 7, 7, 1},
			},
		},
		{
			name: "tag/whitespace",
			src:  "{{ key }}",
			items: []item{
				{token.LDELIM, "{{", 0, 0, 1},
				{token.IDENT, "key", 3, 3, 1},
				{token.RDELIM, "}}", 7, 7, 1},
				{token.EOF, "", 9, 9, 1},
			},
		},
		{
			name: "tag/dot",
			src:  "{{ key1.key2 }}",
			items: []item{
				{token.LDELIM, "{{", 0, 0, 1},
				{token.IDENT, "key1", 3, 3, 1},
				{token.DOT, ".", 7, 7, 1},
				{token.IDENT, "key2", 8, 8, 1},
				{token.RDELIM, "}}", 13, 13, 1},
				{token.EOF, "", 15, 15, 1},
			},
		},
		{
			name: "tag/withSurroundingText",
			src:  "a{{b}}c",
			items: []item{
				{token.TEXT, "a", 0, 0, 1},
				{token.LDELIM, "{{", 1, 1, 1},
				{token.IDENT, "b", 3, 3, 1},
				{token.RDELIM, "}}", 4, 4, 1},
				{token.TEXT, "c", 6, 6, 1},
				{token.EOF, "", 7, 7, 1},
			},
		},
		{
			name: "tag/unescaped",
			src:  "{{{a}}}",
			items: []item{
				{token.LDELIM_UNESCAPED, "{{{", 0, 0, 1},
				{token.IDENT, "a", 3, 3, 1},
				{token.RDELIM_UNESCAPED, "}}}", 4, 4, 1},
				{token.EOF, "", 7, 7, 1},
			},
		},

		{
			name: "tag/unescapeSymbol",
			src:  "{{&a}}",
			items: []item{
				{token.LDELIM_UNESCAPED_SYM, "{{&", 0, 0, 1},
				{token.IDENT, "a", 3, 3, 1},
				{token.RDELIM, "}}", 4, 4, 1},
				{token.EOF, "", 6, 6, 1},
			},
		},
		{
			name: "tag/section",
			src:  "{{#a}}",
			items: []item{
				{token.LDELIM_SECTION, "{{#", 0, 0, 1},
				{token.IDENT, "a", 3, 3, 1},
				{token.RDELIM, "}}", 4, 4, 1},
				{token.EOF, "", 6, 6, 1},
			},
		},
		{
			name: "tag/invertedSection",
			src:  "{{^a}}",
			items: []item{
				{token.LDELIM_INVERSE_SECTION, "{{^", 0, 0, 1},
				{token.IDENT, "a", 3, 3, 1},
				{token.RDELIM, "}}", 4, 4, 1},
				{token.EOF, "", 6, 6, 1},
			},
		},
		{
			name: "tag/sectionEnd",
			src:  "{{/a}}",
			items: []item{
				{token.LDELIM_SECTION_END, "{{/", 0, 0, 1},
				{token.IDENT, "a", 3, 3, 1},
				{token.RDELIM, "}}", 4, 4, 1},
				{token.EOF, "", 6, 6, 1},
			},
		},
		{
			name: "tag/partial",
			src:  "{{>a}}",
			items: []item{
				{token.LDELIM_PARTIAL, "{{>", 0, 0, 1},
				{token.IDENT, "a", 3, 3, 1},
				{token.RDELIM, "}}", 4, 4, 1},
				{token.EOF, "", 6, 6, 1},
			},
		},
		{
			name: "tag/comment",
			src:  "{{! comment }}",
			items: []item{
				{token.LDELIM_COMMENT, "{{!", 0, 0, 1},
				{token.COMMENT, " comment ", 3, 3, 1},
				{token.RDELIM, "}}", 12, 12, 1},
				{token.EOF, "", 14, 14, 1},
			},
		},
		{
			name: "tag/setDelims",
			src:  "{{=<< >>=}}<<a>>",
			items: []item{
				{token.LDELIM_SETDELIM, "{{=", 0, 0, 1},
				{token.DELIMS, "<< >>", 3, 3, 1},
				{token.RDELIM_SETDELIM, "=}}", 8, 8, 1},
				{token.LDELIM, "<<", 11, 11, 1},
				{token.IDENT, "a", 13, 13, 1},
				{token.RDELIM, ">>", 14, 14, 1},
				{token.EOF, "", 16, 16, 1},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			scanner := scanner.New(tc.src)
			var items []item
			for {
				itm := item{
					Token:  scanner.Scan(),
					Text:   scanner.Text(),
					Offset: scanner.Position().Offset,
					Column: scanner.Position().Column,
					Line:   scanner.Position().Line,
				}
				items = append(items, itm)

				if itm.Token == token.EOF || itm.Token == token.ILLEGAL {
					break
				}
			}
			if diff := cmp.Diff(tc.items, items); diff != "" {
				t.Errorf("items mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkScanner_Scan(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		scanner := scanner.New(`
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
			t := scanner.Scan()
			if t == token.EOF {
				break
			}
			if t == token.ILLEGAL {
				b.Fatal(scanner.Text())
			}
		}
	}
}
