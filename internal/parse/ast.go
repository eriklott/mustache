package parse

// Node represents a node in the parse tree.
type Node interface {
	node()
}

// Tree serves as the root container node of the parse tree.
type Tree struct {
	Nodes []Node
}

// Text represents the text template content that exists around and between
// mustache tags. Text does not contain new lines.
type Text struct {
	Value string
}

// VariableTag represents an escaped mustache variable tag.
type VariableTag struct {
	Key       string
	Unescaped bool
}

// SectionTag represents a mustache section tag.
type SectionTag struct {
	Key      string
	Inverted bool
	Nodes    []Node
}

// PartialTag represents a mustache partial tag.
type PartialTag struct {
	Key    string
	Indent int // indent size applied to each new line of the rendered partial
}

func (t *Text) node()        {}
func (t *VariableTag) node() {}
func (t *SectionTag) node()  {}
func (t *PartialTag) node()  {}
