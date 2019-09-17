// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ast

// Node represents a node in the ast tree. Only constructs implementing
// the Node inteface in this package can be added as child nodes.
type Node interface {
	node()
}

// Tree is the representation of a single parsed template.
type Tree struct {
	Name  string
	Nodes []Node
}

// Add appends a child node to the Tree.
func (t *Tree) Add(node Node) {
	t.Nodes = append(t.Nodes, node)
}

// Text node represents text exising between mustache tags.
// When EndOfLine is true, the text string is guaranteed to end with
// with \n or \r\n.
type Text struct {
	Text      string
	EndOfLine bool
}

func (t *Text) node() {}

// Variable represents a mustache variable tag.
type Variable struct {
	Key       []string
	Unescaped bool
	Line      int
	Column    int
}

func (v *Variable) node() {}

// Section a mustache section tag.
type Section struct {
	Key      []string
	Inverted bool
	LDelim   string
	RDelim   string
	Text     string
	Nodes    []Node
	Line     int
	Column   int
}

// Add appends a child node to the Section.
func (s *Section) Add(node Node) {
	s.Nodes = append(s.Nodes, node)
}

func (s *Section) node() {}

// Partial represents a mustache partial tag.
type Partial struct {
	Key    string
	Indent string
	Line   int
	Column int
}

func (p *Partial) node() {}
