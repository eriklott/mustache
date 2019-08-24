package token2_test

import (
	"testing"

	"github.com/eriklott/mustache/internal/token2"
	"github.com/google/go-cmp/cmp"
)

type item struct {
	Token token2.Token
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
				{token2.EOF, ""},
			},
		},
		{
			name: "whitespace",
			src:  "      \r\n    abc  ",
			items: []item{
				{token2.WS, "      "},
				{token2.EOL, "\r\n"},
				{token2.TEXT, "    abc  "},
				{token2.EOF, ""},
			},
		},
		{
			name: "text",
			src:  "abcdefg",
			items: []item{
				{token2.TEXT, "abcdefg"},
				{token2.EOF, ""},
			},
		},
		{
			name: "text & newline",
			src:  "\nabc\r\ndef\n\r\n",
			items: []item{
				{token2.EOL, "\n"},
				{token2.TEXT, "abc"},
				{token2.EOL, "\r\n"},
				{token2.TEXT, "def"},
				{token2.EOL, "\n"},
				{token2.EOL, "\r\n"},
				{token2.EOF, ""},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			scanner := token2.NewScanner(tc.src)
			var items []item
			for {
				itm := item{
					Token: scanner.Next(),
					Text:  scanner.Text(),
				}
				items = append(items, itm)

				if itm.Token == token2.EOF {
					break
				}
			}
			if diff := cmp.Diff(tc.items, items); diff != "" {
				t.Errorf("items mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
