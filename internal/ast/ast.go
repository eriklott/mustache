package ast

import "strings"

type Node interface {
	String(strings.Builder)
	node()
}

type Tree struct {
	Nodes []Node
}

func (t *Tree) String(b strings.Builder) {
	for _, node := range t.Nodes {
		node.String(b)
	}
}

func (t *Tree) Append(node Node) {
	t.Nodes = append(t.Nodes, node)
}

type Text struct {
	Value string
	EOL   bool // End of line - the last char is \r\n or \n
}

func (t *Text) String(b strings.Builder) {
	b.WriteString(t.Value)
}

func (t *Text) node() {}

type Variable struct {
	Key       string
	Unescaped bool
}

func (v *Variable) String(b strings.Builder) {
	b.WriteString("{{")
	if v.Unescaped {
		b.WriteString("&")
	}
	b.WriteString(v.Key)
	b.WriteString("}}")
}

func (v *Variable) node() {}

type Section struct {
	Key      string
	Inverted bool
	Nodes    []Node
}

func (s *Section) String(b strings.Builder) {
	b.WriteString("{{")
	if s.Inverted {
		b.WriteString("^")
	} else {
		b.WriteString("#")
	}
	b.WriteString(s.Key)
	b.WriteString("}}")
	for _, node := range s.Nodes {
		node.String(b)
	}
	b.WriteString("{{/")
	b.WriteString(s.Key)
	b.WriteString("}}")
}

func (s *Section) Append(node Node) {
	s.Nodes = append(s.Nodes, node)
}

func (s *Section) node() {}

type Partial struct {
	Key    string
	Indent string
}

func (p *Partial) String(b strings.Builder) {
	b.WriteString("{{>")
	b.WriteString(p.Key)
	b.WriteString("}}")
}

func (p *Partial) node() {}
