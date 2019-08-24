package mustache

import "io"

const (
	defaultLeftDelim  = "{{"
	defaultRightDelim = "}}"
)

// Node represents a node in the parse tree.
type node interface {
	node()
}

// Tree serves as the root container node of the parse tree.
type tree struct {
	nodes []node
}

// Text represents the text template content that exists around and between
// mustache tags. Text does not contain new lines.
type textNode struct {
	value string
}

// VariableTag represents an escaped mustache variable tag.
type variableTag struct {
	key       string
	unescaped bool
}

// SectionTag represents a mustache section tag.
type sectionTag struct {
	key      string
	inverted bool
	nodes    []node
}

// PartialTag represents a mustache partial tag.
type partialTag struct {
	key    string
	indent int // indent size applied to each new line of the rendered partial
}

func (t *textNode) node()    {}
func (t *variableTag) node() {}
func (t *sectionTag) node()  {}
func (t *partialTag) node()  {}

type parser struct {
	src  string
	otag string
	ctag string
	pos  int // the current position
	ln   int // the current line number
}

func Parse(src string) (*tree, error) {
	p := &parser{
		src:  src,
		otag: defaultLeftDelim,
		ctag: defaultRightDelim,
		pos:  0,
		ln:   1,
	}
	return p.parse()
}

func (p *parser) readUntil(pat string) (string, error) {
	newlines := 0
	for i := p.pos; ; i++ {

		// check for end of src
		if i+len(pat) > len(p.src) {
			return p.src[p.pos:], io.EOF
		}

		// increment number of new lines
		if p.src[i] == '\n' {
			newlines++
		}

		if p.src[i] != pat[0] {
			continue
		}

		match := true
		for j := 0; j < len(pat); j++ {
			if p.src[i+j] != pat[j] {
				match = false
				break
			}
		}

		if match {
			end := i + len(pat)
			text := p.src[p.pos:end]
			p.pos = end

			p.ln += newlines
			return text, nil
		}
	}
}

func (p *parser) parse() (*tree, error) {

	t := &tree{}
	for {
		s, err := p.readUntil(p.otag)
		if err == io.EOF {
			t.nodes = append(t.nodes, &textNode{value: s})
			return t, nil
		}

		// store text preceeding opening tag
		s = s[:len(s)-len(p.otag)-1]
		if len(s) > 0 {
			t.nodes = append(t.nodes, &textNode{value: s})
		}

	}

}
