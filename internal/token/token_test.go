package token_test

import (
	"io"
	"io/ioutil"
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
		{"text", "abc", []item{{token.TEXT, "abc"}}},
		{"leading newline", "\n", []item{{token.NEWLINE, "\n"}}},
		{"leading carriage return", "\r\n", []item{{token.NEWLINE, "\r\n"}}},
		{"text & newline", "abc\n", []item{{token.TEXT, "abc"}, {token.NEWLINE, "\n"}}},
		{"text & carriage return", "abc\r\n", []item{{token.TEXT, "abc"}, {token.NEWLINE, "\r\n"}}},
		{"newline & text", "\nabc", []item{{token.NEWLINE, "\n"}, {token.TEXT, "abc"}}},
		{"tag", "{{a}}", []item{{token.TAG, "{{a}}"}}},
		{"set delims tag", "{{= | | =}}", []item{{token.TAG, "{{= | | =}}"}}},
		{"unescaped tag", "{{{a}}}", []item{{token.TAG, "{{{a}}}"}}},
		{"unescaped tag", "{{{a}}}", []item{{token.TAG, "{{{a}}}"}}},
		{"tags", "{{a}}{{b}}", []item{{token.TAG, "{{a}}"}, {token.TAG, "{{b}}"}}},
		{"change delimes", "{{a}}{{=| |=}}|b|", []item{{token.TAG, "{{a}}"}, {token.TAG, "{{=| |=}}"}, {token.TAG, "|b|"}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			scanner := token.NewScanner("main", tc.src, "{{", "}}")
			items := []item{}
			for {
				scannedItem, err := scanner.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}

				itm := item{
					Token: scannedItem.Token,
					Text:  scannedItem.Text,
				}
				items = append(items, itm)
			}

			if diff := cmp.Diff(tc.items, items); diff != "" {
				t.Errorf("Next() mismatch (-want +got):\n%s", diff)
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
	// scanner := token.NewScanner("main", src, "{{", "}}")

	for n := 0; n < b.N; n++ {
		scanner := token.NewScanner("main", src, "{{", "}}")
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
