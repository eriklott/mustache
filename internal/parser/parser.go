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

func (p *parser) parse(parent appendable) error {
	for {
		// read text that preceeds tag
		text, err := p.readTo(p.ldelim)
		if err == io.EOF {
			if section, ok := parent.(*ast.Section); ok {
				return p.errorf("unclosed section tag: %s", section.Key)
			}
			parent.Append(&ast.Text{Value: text})
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
			return errors.New("unclosed tag")
		}

		// trim right delim from tag
		tag = tag[:len(tag)-len(p.rdelim)]
		tag = strings.TrimSpace(tag)
		if len(tag) == 0 {
			return p.error("empty tag")
		}

		// check if tag is of standalone type
		var padding string
		standalone := false
		if strings.Contains(standaloneTagSymbols, tag[0:1]) {
			// Starting at the tag's left delim, scan backwards to check
			// if the line of text preceeding the tag consists only of whitespace.
			// If is does, seperate it from the text as padding.
			var i int
			for i = len(text); i > 0; i-- {
				if !isSpace(text[i-1]) {
					break
				}
			}

			maybeStandalone := (i == 0 || text[i-1] == '\n')

			if maybeStandalone {

				// Starting at the tag's right delim, scan forwards to check
				// if the text following the tag consists only of whitespace.
				// If is does, set the cursor position after the whitespace, and
				// set the standalone = true.
				eow := p.pos
				for j := p.pos; j < len(p.src); j++ {
					if !isSpace(p.src[j]) {
						eow = j
						break
					}
				}

				if eow == len(p.src) {
					standalone = true
					p.col += eow - p.pos
					p.pos = eow
				} else if eow < len(p.src) && p.src[eow] == '\n' {
					standalone = true
					p.col += eow - p.pos + 1
					p.pos = eow + 1
					p.ln++
				} else if eow+1 < len(p.src) && p.src[eow] == '\r' && p.src[eow+1] == '\n' {
					standalone = true
					p.col += eow - p.pos + 2
					p.pos = eow + 2
					p.ln++
				}

				if standalone {
					padding = text[i:]
					text = text[:i]
				}
			}
		}

		// add text node to parent
		parent.Append(&ast.Text{
			Value: text,
		})

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
			parent.Append(&ast.Partial{
				Key:    key,
				Indent: padding,
			})

		case '!':
			key := strings.TrimSpace(tag[1:])
			if len(key) == 0 {
				return p.error("empty tag")
			}

		case '=':
			delims := strings.TrimSpace(tag[1 : len(tag)-1])
			parts := strings.Split(delims, " ")
			if len(parts) != 2 {
				return p.errorf("invalid set delims tag: %s", delims)
			}
			p.ldelim = parts[0]
			p.rdelim = parts[1]

		default:
			key := tag
			parent.Append(&ast.Variable{
				Key: key,
			})
		}
	}
}

func (p *parser) errorf(format string, a ...interface{}) error {
	return p.error(fmt.Sprintf(format, a...))
}

func (p *parser) error(msg string) error {
	return fmt.Errorf("%d:%d: %s", p.ln, p.col, msg)
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}
