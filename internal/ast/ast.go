package ast

type Type int

const (
	Tree Type = iota
	Text
	Variable
	Section
	Partial
)

type Node struct {
	Type      Type
	Value     string
	Inverted  bool
	Unescaped bool
	EOL       bool
	Indent    string
	Nodes     []Node
}

// ================================================================

// type Tree struct {
// 	Nodes []interface{}
// }

// type Text struct {
// 	Value string
// 	EOL   bool
// }

// type Variable struct {
// 	Key       string
// 	Unescaped bool
// }

// type Section struct {
// 	Key      string
// 	Inverted bool
// 	Nodes    []interface{}
// }

// type Partial struct {
// 	Key    string
// 	Indent string
// }

//=====================================================================================================

// import "strings"

// type Node interface {
// 	String() string
// 	node()
// }

// type Tree struct {
// 	Nodes []Node
// }

// func (t *Tree) String() string {
// 	var b strings.Builder
// 	for _, node := range t.Nodes {
// 		b.WriteString(node.String())
// 	}
// 	return b.String()
// }

// func (t *Tree) Append(node Node) {
// 	t.Nodes = append(t.Nodes, node)
// }

// func (t *Tree) node() {}

// type Text struct {
// 	Value string
// 	EOL   bool // End of line - the last char is \r\n or \n
// }

// func (t *Text) String() string {
// 	return t.Value
// }

// func (t *Text) node() {}

// type Variable struct {
// 	Key       string
// 	Unescaped bool
// }

// func (v *Variable) String() string {
// 	var b strings.Builder
// 	b.WriteString("{{")
// 	if v.Unescaped {
// 		b.WriteString("&")
// 	}
// 	b.WriteString(v.Key)
// 	b.WriteString("}}")
// 	return b.String()
// }

// func (v *Variable) node() {}

// type Section struct {
// 	Key      string
// 	Inverted bool
// 	Nodes    []Node
// }

// func (s *Section) String() string {
// 	var b strings.Builder
// 	b.WriteString("{{")
// 	if s.Inverted {
// 		b.WriteString("^")
// 	} else {
// 		b.WriteString("#")
// 	}
// 	b.WriteString(s.Key)
// 	b.WriteString("}}")
// 	for _, node := range s.Nodes {
// 		b.WriteString(node.String())
// 	}
// 	b.WriteString("{{/")
// 	b.WriteString(s.Key)
// 	b.WriteString("}}")
// 	return b.String()
// }

// func (s *Section) Append(node Node) {
// 	s.Nodes = append(s.Nodes, node)
// }

// func (s *Section) node() {}

// type Partial struct {
// 	Key    string
// 	Indent string
// }

// func (p *Partial) String() string {
// 	var b strings.Builder
// 	b.WriteString("{{>")
// 	b.WriteString(p.Key)
// 	b.WriteString("}}")
// 	return b.String()
// }

// func (p *Partial) node() {}
