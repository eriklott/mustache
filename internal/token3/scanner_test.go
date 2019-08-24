package token3_test

import (
	"testing"

	"github.com/eriklott/mustache/internal/token3"
	"github.com/google/go-cmp/cmp"
)

type item struct {
	Token  token3.Token
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
				{token3.EOF, "", 0, 0, 1},
			},
		},
		{
			name: "whitespace",
			src:  "     ",
			items: []item{
				{token3.WHITESPACE, "     ", 0, 0, 1},
				{token3.EOF, "", 5, 5, 1},
			},
		},
		{
			name: "text",
			src:  "abcde",
			items: []item{
				{token3.TEXT, "abcde", 0, 0, 1},
				{token3.EOF, "", 5, 5, 1},
			},
		},
		{
			name: "text with leading whitespace",
			src:  "  abcde",
			items: []item{
				{token3.TEXT, "  abcde", 0, 0, 1},
				{token3.EOF, "", 7, 7, 1},
			},
		},
		{
			name: "newline",
			src:  "\n",
			items: []item{
				{token3.NEWLINE, "\n", 0, 0, 1},
				{token3.EOF, "", 1, 0, 2},
			},
		},

		// {
		// 	name: "text & newline",
		// 	src:  "\nabc\r\ndef\n\r\n",
		// 	items: []item{
		// 		{token3.EOL, "\n"},
		// 		{token3.TEXT, "abc"},
		// 		{token3.EOL, "\r\n"},
		// 		{token3.TEXT, "def"},
		// 		{token3.EOL, "\n"},
		// 		{token3.EOL, "\r\n"},
		// 		{token3.EOF, ""},
		// 	},
		// },
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			scanner := token3.NewScanner(tc.src)
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

				if itm.Token == token3.EOF {
					break
				}
			}
			if diff := cmp.Diff(tc.items, items); diff != "" {
				t.Errorf("items mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
