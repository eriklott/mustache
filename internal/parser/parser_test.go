package parser_test

import (
	"io/ioutil"
	"testing"

	"github.com/eriklott/mustache/internal/ast"
	"github.com/eriklott/mustache/internal/parser"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tt := []struct {
		name string
		tmpl string
		err  string
		tree *ast.Tree
	}{
		{"Text", "abc", "", &ast.Tree{Nodes: []ast.Node{&ast.Text{Value: "abc"}}}},
		{"Newline", "\n", "", &ast.Tree{Nodes: []ast.Node{&ast.Text{Value: "\n", EOL: true}}}},
		{"Carriage return", "\r\n", "", &ast.Tree{Nodes: []ast.Node{&ast.Text{Value: "\r\n", EOL: true}}}},
		{"Variable", "{{a}}", "", &ast.Tree{Nodes: []ast.Node{&ast.Variable{Key: "a"}}}},
		{"Variable/InternalSpace", "{{ a }}", "", &ast.Tree{Nodes: []ast.Node{&ast.Variable{Key: "a"}}}},
		{"Variable/Empty", "{{}}", "1:5: empty tag", nil},
		{"Variable/UnescapedSymbol", "{{&a}}", "", &ast.Tree{Nodes: []ast.Node{&ast.Variable{Key: "a", Unescaped: true}}}},
		{"Variable/UnescapedSymbol/Empty", "{{&}}", "1:6: empty tag", nil},
		{"Variable/UnescapedDelims", "{{{a}}}", "", &ast.Tree{Nodes: []ast.Node{&ast.Variable{Key: "a", Unescaped: true}}}},
		{"Variable/UnescapedDelims/Empty", "{{{}}}", "1:7: empty tag", nil},
		{"Section", "{{#a}}{{/a}}", "", &ast.Tree{Nodes: []ast.Node{&ast.Section{Key: "a"}}}},
		{"Section/Inverted", "{{^a}}{{/a}}", "", &ast.Tree{Nodes: []ast.Node{&ast.Section{Key: "a", Inverted: true}}}},
		{"Section/MissingClosingTag", "{{#a}}", "1:7: unclosed section tag: a", nil},
		{"Section/MissingOpeningTag", "{{/a}}", "1:7: unexpected section closing tag: a", nil},
		{"Section/EmptyOpeningTag", "{{#}}", "1:6: empty tag", nil},
		{"Section/EmptyClosingTag", "{{/}}", "1:6: empty tag", nil},
		{"Section/Children", "{{#a}}abc{{/a}}", "", &ast.Tree{Nodes: []ast.Node{&ast.Section{Key: "a", Nodes: []ast.Node{&ast.Text{Value: "abc"}}}}}},
		{"Partial", "{{>a}}", "", &ast.Tree{Nodes: []ast.Node{&ast.Partial{Key: "a"}}}},
		{"Partial/EmptyTag", "{{>}}", "1:6: empty tag", nil},
		{"Partial/StandaloneIndent", "  {{>a}}  ", "", &ast.Tree{Nodes: []ast.Node{&ast.Partial{Key: "a", Indent: "  "}}}},
		{"Comment/Skipped", "{{!abc}}", "", &ast.Tree{}},
		{"SetDelim/Skipped", "{{=| |=}}", "", &ast.Tree{}},
		{"SetDelim/ChangeDelims", "{{=| |=}}|a|", "", &ast.Tree{Nodes: []ast.Node{&ast.Variable{Key: "a"}}}},
		{"Standalone/FirstLine", " {{!a}} \nb", "", &ast.Tree{Nodes: []ast.Node{&ast.Text{Value: "b"}}}},
		{"Standalone/MidLine", "\n {{!a}} \nb", "", &ast.Tree{Nodes: []ast.Node{&ast.Text{Value: "\n", EOL: true}, &ast.Text{Value: "b"}}}},
		{"Standalone/LastLine", "a\n {{!b}} ", "", &ast.Tree{Nodes: []ast.Node{&ast.Text{Value: "a\n", EOL: true}}}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.Parse(tc.tmpl)

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
				t.Errorf("nodes mismatch (-want +got):\n%s", diff)
			}
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
		_, err := parser.Parse(tmpl)
		if err != nil {
			b.Fatal((err))
		}
	}
}
