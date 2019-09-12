package parse

type Node interface {
	node()
}

type Tree struct {
	Nodes []Node
}

func (t *Tree) add(node Node) {
	t.Nodes = append(t.Nodes, node)
}

type Text struct {
	Text      string
	EndOfLine bool
}

func (t *Text) node() {}

type VariableTag struct {
	Key       []string
	Unescaped bool
}

func (v *VariableTag) node() {}

type SectionTag struct {
	Key          []string
	Inverted     bool
	LDelim       string
	RDelim       string
	Nodes        []Node
	ChildrenText string
}

func (s *SectionTag) add(node Node) {
	s.Nodes = append(s.Nodes, node)
}

func (s *SectionTag) node() {}

type PartialTag struct {
	Key    string
	Indent string
}

func (p *PartialTag) node() {}
