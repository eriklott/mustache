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

const (
	DefaultLeftDelim  = "{{"
	DefaultRightDelim = "}}"
)

type parser struct {
	name string
	src  string
	s    *token.Scanner
}

func Parse(name, src, leftDelim, rightDelim string) (*ast.Tree, error) {
	p := &parser{
		name: name,
		src:  src,
		s:    token.NewScanner(name, src, leftDelim, rightDelim),
	}
	tree := &ast.Tree{}
	err := p.parse(tree, 0)
	return tree, err
}

type parentNode interface {
	Add(ast.Node)
}

func (p *parser) parse(parent parentNode, start int) error {
	for {
		t, err := p.s.Next()
		if err == io.EOF {
			// If the eof has been reached while parsing the inside of a section, return
			// the eof error to the calling function so the error can be handled there.
			if _, ok := parent.(*ast.Section); ok {
				return err
			}

			// eof reached normally. parsing is complete.
			return nil
		}
		if err != nil {
			return err
		}

		switch t.Type {
		case token.TEXT:
			parent.Add(&ast.Text{
				Text:      t.Text,
				EndOfLine: false,
			})

		case token.TEXT_EOL:
			parent.Add(&ast.Text{
				Text:      t.Text,
				EndOfLine: true,
			})

		case token.VARIABLE:
			parent.Add(&ast.Variable{
				Key:       splitKey(t.Text),
				Unescaped: false,
			})

		case token.UNESCAPED_VARIABLE, token.UNESCAPED_VARIABLE_SYM:
			parent.Add(&ast.Variable{
				Key:       splitKey(t.Text),
				Unescaped: true,
			})

		case token.SECTION, token.INVERTED_SECTION:
			node := &ast.Section{
				Key:      splitKey(t.Text),
				Inverted: t.Type == token.INVERTED_SECTION,
				LDelim:   p.s.LeftDelim(),
				RDelim:   p.s.RightDelim(),
			}
			err := p.parse(node, t.EndOffset)
			if err == io.EOF {
				return p.error(t.Line, t.Column, "unclosed section tag: "+t.Text)
			}
			if err != nil {
				return err
			}
			parent.Add(node)

		case token.SECTION_END:
			section, ok := parent.(*ast.Section)
			if !ok || strings.Join(section.Key, ".") != t.Text {
				return p.error(t.Line, t.Column, "unexpected section closing tag: "+t.Text)
			}
			section.Text = p.src[start:t.Offset]
			return nil

		case token.PARTIAL:
			parent.Add(&ast.Partial{
				Key:    t.Text,
				Indent: t.Indent,
			})
		}
	}
}

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

func splitKey(key string) []string {
	if key == "." {
		return []string{"."}
	}
	return strings.Split(key, ".")
}
