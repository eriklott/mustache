package parser_test

import (
	"testing"

	p "github.com/eriklott/mustache/internal/parser"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tt := []struct {
		name string
		tmpl string
		tree string
	}{
		{
			name: "text",
			tmpl: "a\n  {{#z}}  \nb\n  {{/z}}  \nc",
			tree: "a\n{{#z}}b\n{{/z}}c",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := p.Parse(tc.tmpl)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.tree, tree.String()); diff != "" {
				t.Errorf("tree mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for n := 0; n < b.N; n++ {

		_, err := p.Parse(
			`
		abcdefg asdfjs;aldfjsad asd;lfkjsadlf
		a;lsdkfj;alsdf  as;dlkfjls al;skdfjla;sdf
		asdf
		{{key}}
		{{{key}}}
		{{#key}}
			asldkfjals;dkjf asdlfjkasdlf
		{{! 
			This is a multline comment 
		}}
		{{=<< >>=}}
		<<^key>>
		<</key>>
		<</key>>
		<<>key>>
		<<&key>>
	`)
		if err != nil {
			b.Fatal(err)
		}

	}
}
