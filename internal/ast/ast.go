package ast

import "strings"

type Node interface {
	String() string
	node()
}

type Tree struct {
	Nodes []Node
}

func (t *Tree) String() string {
	var b strings.Builder
	for _, node := range t.Nodes {
		b.WriteString(node.String())
	}
	return b.String()
}

func (n *Text) node() {}

func (t *Tree) Append(node Node) {
	t.Nodes = append(t.Nodes, node)
}

type Text struct {
	Value string
}

func (t *Text) String() string {
	return t.Value
}

type Variable struct {
	Key       string
	Unescaped bool
}

func (v *Variable) String() string {
	var b strings.Builder
	b.WriteString("{{")
	if v.Unescaped {
		b.WriteString("&")
	}
	b.WriteString(v.Key)
	b.WriteString("}}")
	return b.String()
}

func (n *Variable) node() {}

type Section struct {
	Key      string
	Inverted bool
	Nodes    []Node
}

func (s *Section) String() string {
	var b strings.Builder
	b.WriteString("{{#")
	b.WriteString(s.Key)
	b.WriteString("}}")
	for _, node := range s.Nodes {
		b.WriteString(node.String())
	}
	b.WriteString("{{/")
	b.WriteString(s.Key)
	b.WriteString("}}")
	return b.String()
}

func (s *Section) Append(node Node) {
	s.Nodes = append(s.Nodes, node)
}

func (n *Section) node() {}

type Partial struct {
	Key    string
	Indent string
}

func (p *Partial) String() string {
	var b strings.Builder
	b.WriteString("{{>")
	b.WriteString(p.Key)
	b.WriteString("}}")
	return b.String()
}

func (n *Partial) node() {}
