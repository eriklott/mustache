// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/eriklott/mustache/context"
	"github.com/eriklott/mustache/parse"
	"github.com/eriklott/mustache/token"
	"github.com/eriklott/mustache/writer"
)

// Render fetches a template/partial by name, applies it to the specified
// data object(s), and writes the resulting output to w. If multiple data
// objects are provided, the group will be considered a data stack, with lookups
// occuring on the 0'th item first, moving sequenentially to the N'th item, until
// a match is found.
//
// If an error is returned, it will describe the first error that occured
// during the rendering process. Errors consist of context lookup misses, unknown
// partials, or parsing errors in the case of lambda functions.
//
// Render always produces a fully rendered template, regardless of if an error
// has been returned or not. As per the mustache spec, context misses, unknown
// partials, and unparsable lambdas are all considered falsey values.
func (t *Template) Render(w io.Writer, name string, ctx ...interface{}) error {
	// reverse data stack args
	rctx := []interface{}{}
	for i := len(ctx) - 1; i >= 0; i-- {
		rctx = append(rctx, ctx[i])
	}
	stack := context.NewStack(rctx...)
	// init
	t.renderErr = nil
	wr := writer.New(w)
	// render
	t.renderPartial(wr, stack, name)
	return t.renderErr
}

func (t *Template) renderPartial(w writer.Writer, stack *context.Stack, name string) {
	tree, ok := t.treeMap[name]
	if ok {
		t.walk(w, stack, tree)
	} else {
		t.handleErr(fmt.Errorf("template/partial %s not found", name))
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

func (t *Template) handleErr(err error) {
	if t.renderErr == nil && err != nil {
		t.renderErr = err
	}
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
	t.handleErr(err)
	if err == nil && val.IsTruthy() {
		var str string
		if val.IsLambda() {
			var b bytes.Buffer
			err = t.renderString(writer.New(&b), stack, val.CallLambda(), token.DefaultLeftDelim, token.DefaultRightDelim)
			t.handleErr(err)
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
	t.handleErr(err)
	if err == nil && val.IsTruthy() {
		switch {
		case val.IsSectionLambda():
			err = t.renderString(w, stack, val.CallSectionLambda(node.NodesString()), node.LDelim, node.RDelim)
			t.handleErr(err)
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
	t.handleErr(err)
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
