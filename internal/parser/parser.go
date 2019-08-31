package parser

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/eriklott/mustache/internal/ast"
)

const (
	defaultLeftDelim     = "{{"
	defaultRightDelim    = "}}"
	standaloneTagSymbols = "#^/>=!"
)

type appendable interface {
	Append(node ast.Node)
}

type parser struct {
	src    string
	pos    int
	col    int
	ln     int
	ldelim string
	rdelim string
}

// Parse converts a template string into an AST representing the template's structure.
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

// read advances and returns the next byte from the source string. Returns io.EOF
// at the end of string.
func (p *parser) read() (byte, error) {
	if p.pos >= len(p.src) {
		return 0, io.EOF
	}
	b := p.src[p.pos]
	p.pos++
	p.col++
	if b == '\n' {
		p.col = 1
		p.ln++
	}
	return b, nil
}

// readTo advances through the string until reaching a target pattern. Returns io.EOF
// at the end of the string, as well as any text that has been scanned up to that point.
func (p *parser) readTo(pattern string) (string, error) {
	start := p.pos
	matchIdx := 0
	for {
		b, err := p.read()
		if err == io.EOF {
			return p.src[start:p.pos], io.EOF
		}

		if b != pattern[matchIdx] {
			matchIdx = 0
			continue
		}

		matchIdx++
		if matchIdx == len(pattern) {
			return p.src[start:p.pos], nil
		}
	}
}

var eol = errors.New("end of line")

// readLineTo has the same scanning behaviour as readTo, except that it stops
// scanning at the end of a line, returning EOL error
func (p *parser) readLineTo(pattern string) (string, error) {
	start := p.pos
	matchIdx := 0
	for {
		if p.pos >= len(p.src) {
			return p.src[start:p.pos], io.EOF
		}
		b := p.src[p.pos]
		p.pos++
		p.col++
		if b == '\n' {
			p.col = 1
			p.ln++
			return p.src[start:p.pos], eol
		}

		if b != pattern[matchIdx] {
			matchIdx = 0
			continue
		}

		matchIdx++
		if matchIdx == len(pattern) {
			return p.src[start:p.pos], nil
		}
	}
}

// parse returns an ast tree representation of the source string.
func (p *parser) parse(parent appendable) error {
	for {
		textStart := p.pos

		// read line of text that preceeds tag
		text, err := p.readLineTo(p.ldelim)

		// reached end of line
		if err == eol {
			parent.Append(&ast.Text{Value: text, EOL: true})
			continue
		}

		if err == io.EOF {
			if section, ok := parent.(*ast.Section); ok {
				return p.errorf("unclosed section tag: %s", section.Key)
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
		tag = strings.TrimSpace(tag)
		if len(tag) == 0 {
			return p.error("empty tag")
		}

		// determine if tag is of standalone type
		isStandaloneTag := false
		if strings.Contains(standaloneTagSymbols, tag[0:1]) {
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
			key := strings.TrimSpace(tag[1 : len(tag)-1])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			parent.Append(&ast.Variable{
				Key:       key,
				Unescaped: true,
			})

		case '&':
			key := strings.TrimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			parent.Append(&ast.Variable{
				Key:       key,
				Unescaped: true,
			})

		case '#':
			key := strings.TrimSpace(tag[1:])
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
			key := strings.TrimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			node := &ast.Section{
				Key:      key,
				Inverted: true,
			}
			err = p.parse(node)
			if err != nil {
				return err
			}
			parent.Append(node)

		case '/':
			key := strings.TrimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			if section, ok := parent.(*ast.Section); !ok || key != section.Key {
				return p.error("unexpected section closing tag")
			}
			return nil

		case '>':
			key := strings.TrimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}
			node := &ast.Partial{Key: key}
			if isStandaloneTag {
				node.Indent = text
			}
			parent.Append(node)

		case '!':
			key := strings.TrimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}

		case '=':
			delims := strings.TrimSpace(tag[1 : len(tag)-1])
			parts := strings.SplitN(delims, " ", 2)
			if len(parts) == 2 {
				p.ldelim = strings.TrimSpace(parts[0])
				p.rdelim = strings.TrimSpace(parts[1])
			}

		default:
			key := tag
			parent.Append(&ast.Variable{Key: key})
		}
	}
}

// errorf returns a formatted error with line/column prefix
func (p *parser) errorf(format string, a ...interface{}) error {
	return p.error(fmt.Sprintf(format, a...))
}

// error returns an error with line/col prefix
func (p *parser) error(msg string) error {
	return fmt.Errorf("%d:%d: %s", p.ln, p.col, msg)
}

// returns true when the byte represents a space
func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}
