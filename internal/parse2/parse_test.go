package parse2_test

import (
	"io/ioutil"
	"testing"

	parse "github.com/eriklott/mustache/internal/parse2"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tt := []struct {
		name string
		tmpl string
		err  string
		tree *parse.Tree
	}{
		// {
		// 	name: "Text",
		// 	tmpl: "abc",
		// 	nodes: []node.Node{
		// 		node.NewText("abc", false, 3),
		// 	},
		// },
		// {
		// 	name: "Newline",
		// 	tmpl: "\n",
		// 	nodes: []node.Node{
		// 		node.NewText("\n", true, 1),
		// 	},
		// },
		// {
		// 	name: "Carriage return",
		// 	tmpl: "\r\n",
		// 	nodes: []node.Node{
		// 		node.NewText("\r\n", true, 2),
		// 	},
		// },
		// {
		// 	name: "Variable",
		// 	tmpl: "{{a}}",
		// 	nodes: []node.Node{
		// 		node.NewVariable("a", false, 5),
		// 	},
		// },
		// {
		// 	name: "Variable/Whitespace",
		// 	tmpl: "{{ a }}",
		// 	nodes: []node.Node{
		// 		node.NewVariable("a", false, 7),
		// 	},
		// },
		// {
		// 	name: "Variable/Empty",
		// 	tmpl: "{{}}",
		// 	err:  "main:1:1: missing key",
		// },
		// {
		// 	name: "Variable/Key/Dotted",
		// 	tmpl: "{{a.b.c}}",
		// 	nodes: []node.Node{
		// 		node.NewVariable("a.b.c", false, 9),
		// 	},
		// },
		// {
		// 	name: "Variable/Key/LeadingDot",
		// 	tmpl: "{{.a.b}}",
		// 	err:  "main:1:1: invalid key: .a.b",
		// },
		// {
		// 	name: "Variable/Key/TrailingDot",
		// 	tmpl: "{{a.b.}}",
		// 	err:  "main:1:1: invalid key: a.b.",
		// },
		// {
		// 	name: "Variable/Key/Whitespace",
		// 	tmpl: "{{a . b}}",
		// 	err:  "main:1:1: invalid key: a . b",
		// },
		// {
		// 	name: "Variable/Key/SingleDot",
		// 	tmpl: "{{.}}",
		// 	nodes: []node.Node{
		// 		node.NewVariable(".", false, 5),
		// 	},
		// },
		// {
		// 	name: "Variable/UnescapedSymbol",
		// 	tmpl: "{{&a}}",
		// 	nodes: []node.Node{
		// 		node.NewVariable("a", true, 6),
		// 	},
		// },
		// {
		// 	name: "Variable/Unescaped",
		// 	tmpl: "{{{a}}}",
		// 	nodes: []node.Node{
		// 		node.NewVariable("a", true, 7),
		// 	},
		// },
		// {
		// 	name: "Variable/Unescaped/Empty",
		// 	tmpl: "{{{}}}",
		// 	err:  "main:1:1: missing key",
		// },
		// {
		// 	name: "Variable/Unescaped/MismatchDelims",
		// 	tmpl: "{{{a}}",
		// 	err:  "main:1:1: unclosed tag",
		// },
		// {
		// 	name: "Section",
		// 	tmpl: "{{#a}}{{/a}}",
		// 	nodes: []node.Node{
		// 		node.NewSection("a", false, "{{ }}", 6),
		// 	},
		// },
		// {
		// 	name: "Section/Inverted",
		// 	tmpl: "{{^a}}{{/a}}",
		// 	nodes: []node.Node{
		// 		node.NewSection("a", true, "{{ }}", 6),
		// 	},
		// },
		// {
		// 	name: "Section/MissingClosingTag",
		// 	tmpl: "{{#a}}",
		// 	err:  "main:1:1: unclosed section tag: a",
		// },
		// {
		// 	name: "Section/MissingOpeningTag",
		// 	tmpl: "{{/a}}",
		// 	err:  "main:1:1: unexpected section closing tag: a",
		// },
		// {
		// 	name: "Section/EmptyOpeningTag",
		// 	tmpl: "{{#}}",
		// 	err:  "main:1:1: missing key",
		// },
		// {
		// 	name: "Section/EmptyClosingTag",
		// 	tmpl: "{{/}}",
		// 	err:  "main:1:1: missing key",
		// },
		{
			name: "Section/Children",
			tmpl: "{{#a}}abc{{/a}}",
			tree: &parse.Tree{
				Nodes: []parse.Node{
					&parse.SectionTag{
						Key:          []string{"a"},
						Inverted:     false,
						LDelim:       "{{",
						RDelim:       "}}",
						ChildrenText: "abc",
						Nodes: []parse.Node{
							&parse.Text{Text: "abc"},
						},
					},
				},
			},
		},

		// {
		// 	name: "Partial",
		// 	tmpl: "{{>a}}",
		// 	nodes: []node.Node{
		// 		node.NewPartial("a", "", 6),
		// 	},
		// },
		// {
		// 	name: "Partial/Empty",
		// 	tmpl: "{{>}}",
		// 	err:  "main:1:1: missing key",
		// },
		// {
		// 	name:  "Comment/Skipped",
		// 	tmpl:  "{{! This is a comment }}",
		// 	nodes: []node.Node{},
		// },
		// {
		// 	name: "SetDelim",
		// 	tmpl: "{{=| |=}}",
		// 	nodes: []node.Node{
		// 		node.NewSetDelimeter("{{=| |=}}", 9),
		// 	},
		// },
		// {
		// 	name: "SetDelim/ChangesDelimeters",
		// 	tmpl: "{{=| |=}}|a|",
		// 	nodes: []node.Node{
		// 		node.NewSetDelimeter("{{=| |=}}", 9),
		// 		node.NewVariable("a", false, 12),
		// 	},
		// },
		// {
		// 	name: "Standalone/FirstLine",
		// 	tmpl: " {{!a}} \nb",
		// 	nodes: []node.Node{
		// 		node.NewText("b", false, 10),
		// 	},
		// },
		// {
		// 	name: "Standalone/MidLine",
		// 	tmpl: "\n {{!a}} \nb",
		// 	nodes: []node.Node{
		// 		node.NewText("\n", true, 1),
		// 		node.NewText("b", false, 11),
		// 	},
		// },
		// {
		// 	name: "Standalone/LastLine",
		// 	tmpl: "a\n {{!b}} ",
		// 	nodes: []node.Node{
		// 		node.NewText("a\n", true, 2),
		// 	},
		// },
		// {
		// 	name: "Partial/StandaloneIndent",
		// 	tmpl: "  {{>a}}  ",
		// 	nodes: []node.Node{
		// 		node.NewPartial("a", "  ", 8),
		// 	},
		// },
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parse.Parse("main", tc.tmpl, parse.DefaultLeftDelim, parse.DefaultRightDelim)

			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tc.err {
				t.Errorf("unexpected error, got: %s, want: %s", errStr, tc.err)
			}
			if err != nil || tc.err != "" {
				return
			}

			if diff := cmp.Diff(tc.tree, tree); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}

			// if !reflect.DeepEqual(tree, tc.tree) {
			// 	t.Errorf("unexpected tree, got: %+v, want: %+v", tree, tc.tree)
			// }
		})
	}
}

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
