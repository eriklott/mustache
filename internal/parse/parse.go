// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parse

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/eriklott/mustache/internal/ast"
	"github.com/eriklott/mustache/internal/token"
)

// Default mustache delimeters
const (
	DefaultLeftDelim  = "{{"
	DefaultRightDelim = "}}"
)

// parser contains the state for the parsing process.
type parser struct {
	name string
	src  string
	s    *token.Scanner
}

// Parse transforms a template string into a tree of nodes. If an error is
// encountered, parsing stops and the error is returned.
func Parse(name, src, leftDelim, rightDelim string) (ast.Node, error) {
	p := &parser{
		name: name,
		src:  src,
		s:    token.NewScanner(name, src, leftDelim, rightDelim),
	}
	tree := ast.Node{
		Type: ast.Tree,
		V1:   name,
	}
	return p.parse(tree, 0)
}

// parentNode represents an element that can add an ast.Node (mainly a ast.Tree or ast.Section)
type parentNode interface {
	Add(ast.Node)
}

// parse recursively parses the template string, constructing nodes and adding them to
// the tree. If an error is encountered, parse stops and the error is returned.
func (p *parser) parse(parent ast.Node, start int) (ast.Node, error) {
	for {
		t, err := p.s.Next()
		if err == io.EOF {
			// If the eof has been reached while parsing the inside of a section, return
			// the eof error to the calling function so the error can be handled there.
			if parent.IsSection() {
				return parent, err
			}

			// eof reached normally. parsing is complete.
			return parent, nil
		}
		if err != nil {
			return parent, err
		}

		switch t.Type {
		case token.TEXT:
			parent.Nodes = append(parent.Nodes, ast.Node{
				Type:   ast.Text,
				V1:     t.Text,
				Line:   t.Line,
				Column: t.Column,
			})

		case token.TEXT_EOL:
			parent.Nodes = append(parent.Nodes, ast.Node{
				Type:   ast.TextEOF,
				V1:     t.Text,
				Line:   t.Line,
				Column: t.Column,
			})

		case token.VARIABLE:
			parent.Nodes = append(parent.Nodes, ast.Node{
				Type:   ast.Variable,
				V1:     t.Text,
				Line:   t.Line,
				Column: t.Column,
			})

		case token.UNESCAPED_VARIABLE, token.UNESCAPED_VARIABLE_SYM:
			parent.Nodes = append(parent.Nodes, ast.Node{
				Type:   ast.UnescapedVariable,
				V1:     t.Text,
				Line:   t.Line,
				Column: t.Column,
			})

		case token.SECTION, token.INVERTED_SECTION:
			typ := ast.Section
			if t.Type == token.INVERTED_SECTION {
				typ = ast.InvertedSection
			}
			node := ast.Node{
				Type:   typ,
				V1:     t.Text,
				V2:     p.s.LeftDelim() + " " + p.s.RightDelim(),
				Line:   t.Line,
				Column: t.Column,
			}
			node, err = p.parse(node, t.EndOffset)
			if err == io.EOF {
				return parent, p.error(t.Line, t.Column, "unclosed section tag: "+t.Text)
			}
			if err != nil {
				return parent, err
			}
			parent.Nodes = append(parent.Nodes, node)

		case token.SECTION_END:
			if !parent.IsSection() || parent.Key() != t.Text {
				return parent, p.error(t.Line, t.Column, "unexpected section closing tag: "+t.Text)
			}
			parent.V3 = p.src[start:t.Offset]
			return parent, nil

		case token.PARTIAL:
			parent.Nodes = append(parent.Nodes, ast.Node{
				Type:   ast.Partial,
				V1:     t.Text,
				V2:     t.Indent,
				Line:   t.Line,
				Column: t.Column,
			})
		}
	}
}

// error returns an error message prefixed with the line and column number of
// where in the template the error occured.
func (p *parser) error(ln, col int, msg string) error {
	var b strings.Builder
	b.WriteString(p.name)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(ln))
	b.WriteString(":")
	b.WriteString(strconv.Itoa(col))
	b.WriteString(":")
	b.WriteString(" ")
	b.WriteString(msg)
	return errors.New(b.String())
}
