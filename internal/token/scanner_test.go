package token_test

import (
	"testing"

	"github.com/eriklott/mustache/internal/token"
	"github.com/google/go-cmp/cmp"
)

type item struct {
	Token token.Token
	Text  string
}

func TestScanner_Next(t *testing.T) {
	tt := []struct {
		name  string
		src   string
		items []item
	}{
		{
			name: "eof",
			src:  "",
			items: []item{
				{token.EOF, ""},
			},
		},
		{
			name: "whitespace",
			src:  "      \r\n    abc  ",
			items: []item{
				{token.WS, "      "},
				{token.EOL, "\r\n"},
				{token.TEXT, "    abc  "},
				{token.EOF, ""},
			},
		},
		{
			name: "text",
			src:  "abcdefg",
			items: []item{
				{token.TEXT, "abcdefg"},
				{token.EOF, ""},
			},
		},
		{
			name: "text & newline",
			src:  "\nabc\r\ndef\n\r\n",
			items: []item{
				{token.EOL, "\n"},
				{token.TEXT, "abc"},
				{token.EOL, "\r\n"},
				{token.TEXT, "def"},
				{token.EOL, "\n"},
				{token.EOL, "\r\n"},
				{token.EOF, ""},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			scanner := token.NewScanner(tc.src)
			var items []item
			for {
				itm := item{
					Token: scanner.Next(),
					Text:  scanner.Text(),
				}
				items = append(items, itm)

				if itm.Token == token.EOF {
					break
				}
			}
			if diff := cmp.Diff(tc.items, items); diff != "" {
				t.Errorf("items mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
