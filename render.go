package mustache

import (
	"bytes"
	"io"
	"strings"

	"github.com/mreriklott/mustache/context"
	"github.com/mreriklott/mustache/parse"
	"github.com/mreriklott/mustache/token"
	"github.com/mreriklott/mustache/writer"
)

func (t *Template) Render(wr io.Writer, name string, ctx ...interface{}) {
	w := writer.New(wr)
	stack := context.NewStack(ctx...)
	t.renderPartial(w, stack, name)
}

func (t *Template) renderPartial(w writer.Writer, stack *context.Stack, name string) {
	if tree, ok := t.treeMap[name]; ok {
		t.walk(w, stack, tree)
	}
}

func (t *Template) renderString(w writer.Writer, stack *context.Stack, src, ldelim, rdelim string) error {
	tokenReader := token.NewReader("", strings.NewReader(src), ldelim, rdelim)
	tree, err := parse.Parse(tokenReader)
	if err != nil {
		return err
	}
	t.walk(w, stack, tree)
	return nil
}

func (t *Template) walk(w writer.Writer, stack *context.Stack, node parse.Node) {
	switch node := node.(type) {
	case *parse.Tree:
		t.walkTree(w, stack, node)
	case *parse.Text:
		t.walkText(w, stack, node)
	case *parse.Pad:
		t.walkPad(w, stack, node)
	case *parse.NewLine:
		t.walkNewLine(w, stack, node)
	case *parse.VariableTag:
		t.walkVariableTag(w, stack, node)
	case *parse.SectionTag:
		if node.Inverted {
			t.walkInvertedSectionTag(w, stack, node)
		} else {
			t.walkSectionTag(w, stack, node)
		}
	case *parse.PartialTag:
		t.walkPartialTag(w, stack, node)
	}
}

func (t *Template) walkTree(w writer.Writer, stack *context.Stack, node *parse.Tree) {
	for _, child := range node.Nodes {
		t.walk(w, stack, child)
	}
}

func (t *Template) walkText(w writer.Writer, stack *context.Stack, node *parse.Text) {
	w.Write(node.Text)
}

func (t *Template) walkPad(w writer.Writer, stack *context.Stack, node *parse.Pad) {
	w.Write(node.Text)
}

func (t *Template) walkNewLine(w writer.Writer, stack *context.Stack, node *parse.NewLine) {
	w.Write(node.Text)
	w.IndentNext()
}

func (t *Template) walkVariableTag(w writer.Writer, stack *context.Stack, node *parse.VariableTag) {
	val, err := stack.Lookup(node.Key)
	if err == nil && val.IsTruthy() {
		var str string
		if val.IsLambda() {
			var b bytes.Buffer
			t.renderString(writer.New(&b), stack, val.CallLambda(), token.DefaultLeftDelim, token.DefaultRightDelim)
			str = b.String()
		} else {
			str = val.String()
		}
		if node.Unescaped {
			w.Write(str)
		} else {
			w.WriteHTMLEscaped(str)
		}
	}
}

func (t *Template) walkSectionTag(w writer.Writer, stack *context.Stack, node *parse.SectionTag) {
	val, err := stack.Lookup(node.Key)
	if err == nil && val.IsTruthy() {
		switch {
		case val.IsSectionLambda():
			t.renderString(w, stack, val.CallSectionLambda(node.NodesString()), node.LDelim, node.RDelim)
		case val.IsList():
			len := val.Len()
			for i := 0; i < len; i++ {
				stack.Add(val.Index(i))
				for _, child := range node.Nodes {
					t.walk(w, stack, child)
				}
				stack.Remove()
			}
		default:
			stack.Add(val)
			for _, child := range node.Nodes {
				t.walk(w, stack, child)
			}
			stack.Remove()
		}
	}
}

func (t *Template) walkInvertedSectionTag(w writer.Writer, stack *context.Stack, node *parse.SectionTag) {
	val, err := stack.Lookup(node.Key)
	if err != nil || !val.IsTruthy() {
		for _, child := range node.Nodes {
			t.walk(w, stack, child)
		}
	}
}

func (t *Template) walkPartialTag(w writer.Writer, stack *context.Stack, node *parse.PartialTag) {
	w.IncreaseIndent(node.Indent)
	w.IndentNext()
	t.renderPartial(w, stack, node.Key)
	w.DecreaseIndent(node.Indent)
}
