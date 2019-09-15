// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ast

type Node interface {
	node()
}

type Tree struct {
	Nodes []Node
}

func (t *Tree) Add(node Node) {
	t.Nodes = append(t.Nodes, node)
}

type Text struct {
	Text      string
	EndOfLine bool
}

func (t *Text) node() {}

type Variable struct {
	Key       []string
	Unescaped bool
}

func (v *Variable) node() {}

type Section struct {
	Key      []string
	Inverted bool
	LDelim   string
	RDelim   string
	Text     string
	Nodes    []Node
}

func (s *Section) Add(node Node) {
	s.Nodes = append(s.Nodes, node)
}

func (s *Section) node() {}

type Partial struct {
	Key    string
	Indent string
}

func (p *Partial) node() {}
