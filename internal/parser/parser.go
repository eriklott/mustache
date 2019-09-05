package parser

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/eriklott/mustache/internal/ast"
)

const (
	defaultLeftDelim  = "{{"
	defaultRightDelim = "}}"
)

type parentNode interface {
	Append(ast.Node)
}

var errEOL = errors.New("end of line")

// parser holds the state for the parsing process
type parser struct {
	src    string
	pos    int
	col    int
	ln     int
	ldelim string
	rdelim string
}

// Parse converts a template string into a list of AST nodes.
func Parse(src string) (*ast.Tree, error) {
	p := &parser{
		src:    src,
		pos:    0,
		col:    1,
		ln:     1,
		ldelim: defaultLeftDelim,
		rdelim: defaultRightDelim,
	}
	tree := &ast.Tree{}
	err := p.parse(tree)
	return tree, err
}

// readTo advances through the string until reaching a target pattern. Returns io.EOF
// at the end of the string, as well as any text that has been scanned up to that point.
func (p *parser) readTo(pattern string) (string, error) {
	start := p.pos
	matchIdx := 0
	srcLen := len(p.src)
	patLen := len(pattern)
	for {
		if p.pos >= srcLen {
			return p.src[start:p.pos], io.EOF
		}
		b := p.src[p.pos]
		p.pos++
		p.col++
		if b == '\n' {
			p.col = 1
			p.ln++
		}

		if b != pattern[matchIdx] {
			matchIdx = 0
			continue
		}

		matchIdx++
		if matchIdx == patLen {
			return p.src[start:p.pos], nil
		}
	}
}

// readLineTo has the same scanning behaviour as readTo, except that it stops
// scanning at the end of a line, returning errEOL
func (p *parser) readLineTo(pattern string) (string, error) {
	start := p.pos
	matchIdx := 0
	srcLen := len(p.src)
	patLen := len(pattern)
	for {
		if p.pos >= srcLen {
			return p.src[start:p.pos], io.EOF
		}
		b := p.src[p.pos]
		p.pos++
		p.col++
		if b == '\n' {
			p.col = 1
			p.ln++
			return p.src[start:p.pos], errEOL
		}

		if b != pattern[matchIdx] {
			matchIdx = 0
			continue
		}

		matchIdx++
		if matchIdx == patLen {
			return p.src[start:p.pos], nil
		}
	}
}

// parse returns a list of ast nodes.
func (p *parser) parse(parent parentNode) error {
	for {
		textStart := p.pos

		// read line of text that preceeds tag
		text, err := p.readLineTo(p.ldelim)

		// reached end of line
		if err == errEOL {
			parent.Append(&ast.Text{
				Value: text,
				EOL:   true,
			})
			continue
		}

		if err == io.EOF {
			if section, ok := parent.(*ast.Section); ok {
				return p.error("unclosed section tag: " + section.Key)
			}
			if len(text) > 0 {
				parent.Append(&ast.Text{Value: text})
			}
			return nil
		}

		// trim left delim from text
		text = text[:len(text)-len(p.ldelim)]

		// read tag
		var tag string
		if p.pos < len(p.src) && p.src[p.pos] == '{' {
			tag, err = p.readTo("}" + p.rdelim)
		} else {
			tag, err = p.readTo(p.rdelim)
		}
		if err == io.EOF {
			return p.error("unclosed tag")
		}

		// trim right delim from tag
		tag = tag[:len(tag)-len(p.rdelim)]
		if len(tag) == 0 {
			return p.error("empty tag")
		}

		// determine if tag is of standalone type
		isStandaloneTag := false

		if isStandaloneTagSym(tag[0]) {
			// Starting at the tag's left delim, scan backwards to check
			// if the line of text preceeding the tag consists only of whitespace
			// since the last newline character (or the beginning of the template string).
			// If is does, seperate it from the text as padding.
			var i int
			for i = textStart + len(text); i > 0; i-- {
				if !isSpace(p.src[i-1]) {
					break
				}
			}

			if i == 0 || p.src[i-1] == '\n' {
				// Starting at the tag's right delim, scan forwards to check
				// if the text following the tag consists only of whitespace until the
				// next newline char, or the end of template string.
				// If is does, set the cursor position after the whitespace (and newline chars), and
				// set the standalone = true.
				var j int
				for j = p.pos; j < len(p.src); j++ {
					if !isSpace(p.src[j]) {
						break
					}
				}

				if j >= len(p.src) {
					isStandaloneTag = true
					p.col += (j - p.pos)
					p.pos = j
				} else if p.src[j] == '\n' {
					isStandaloneTag = true
					p.col += (j - p.pos + 1)
					p.pos = j + 1
					p.ln++
				} else if j+1 < len(p.src) && p.src[j] == '\r' && p.src[j+1] == '\n' {
					isStandaloneTag = true
					p.col += (j - p.pos + 2)
					p.pos = j + 2
					p.ln++
				}
			}
		}

		// add text node to parent
		if !isStandaloneTag && len(text) > 0 {
			parent.Append(&ast.Text{Value: text})
		}

		switch tag[0] {
		case '{':
			key := trimSpace(tag[1 : len(tag)-1])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			parent.Append(&ast.Variable{
				Key:       key,
				Unescaped: true,
			})

		case '&':
			key := trimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			parent.Append(&ast.Variable{
				Key:       key,
				Unescaped: true,
			})

		case '#':
			key := trimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			node := &ast.Section{Key: key}
			err = p.parse(node)
			if err != nil {
				return err
			}
			parent.Append(node)

		case '^':
			key := trimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			node := &ast.Section{
				Key:      key,
				Inverted: true,
			}
			err := p.parse(node)
			if err != nil {
				return err
			}
			parent.Append(node)

		case '/':
			key := trimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			if section, ok := parent.(*ast.Section); !ok || section.Key != key {
				return p.error("unexpected section closing tag: " + key)
			}
			return nil

		case '>':
			key := trimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			node := &ast.Partial{Key: key}
			if isStandaloneTag && len(text) > 0 {
				node.Indent = text
			}
			parent.Append(node)

		case '!':
			key := trimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}

		case '=':
			delims := trimSpace(tag[1 : len(tag)-1])
			parts := strings.SplitN(delims, " ", 2)
			if len(parts) == 2 {
				p.ldelim = trimSpace(parts[0])
				p.rdelim = trimSpace(parts[1])
			}

		default:
			key := trimSpace(tag)
			parent.Append(&ast.Variable{Key: key})
		}
	}
}

type parseError struct {
	col int
	ln  int
	msg string
}

func (e parseError) Error() string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(e.ln))
	b.WriteString(":")
	b.WriteString(strconv.Itoa(e.col))
	b.WriteString(": ")
	b.WriteString(e.msg)
	s := b.String()
	return s
}

// error returns an error with line/col prefix
func (p *parser) error(msg string) error {
	return parseError{
		col: p.col,
		ln:  p.ln,
		msg: msg,
	}
}

// returns true when the byte represents a space
func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}

func isStandaloneTagSym(b byte) bool {
	return b == '#' || b == '^' || b == '/' || b == '>' || b == '=' || b == '!'
}

func trimLeftSpace(s string) string {
	var i int
	for i = 0; i < len(s); i++ {
		if !isSpace(s[i]) {
			return s[i:]
		}
	}
	return s
}

func trimRightSpace(s string) string {
	var i int
	for i = len(s); i > 0; i-- {
		if !isSpace(s[i-1]) {
			return s[:i]
		}
	}
	return s
}

func trimSpace(s string) string {
	return trimLeftSpace(trimRightSpace(s))
}
