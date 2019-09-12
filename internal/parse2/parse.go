package parse2

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/eriklott/mustache/internal/token"
)

const (
	bufN = 4
)

const (
	DefaultLeftDelim  = "{{"
	DefaultRightDelim = "}}"
)

type parentNode interface {
	add(Node)
}

type parser struct {
	scanner *token.Scanner
	buf     [bufN]token.Item
	bufLen  int
	pos     int
	isBuf   bool
}

func Parse(name, src, ldelim, rdelim string) (*Tree, error) {
	p := &parser{
		scanner: token.NewScanner(name, src, ldelim, rdelim),
		isBuf:   true,
	}
	tree := &Tree{}
	err := p.parse(tree, 0)
	return tree, err
}

func (p *parser) next() (token.Item, error) {
	// if there are items in the buffer, return one.
	if p.pos < p.bufLen {
		item := p.buf[p.pos]
		p.pos++
		return item, nil
	}

	// if we are not currently buffering, return items directly
	// from the scanner. Reactivate the buffer after a
	// newline token has been received.
	if !p.isBuf {
		item, err := p.scanner.Next()
		if err == nil && item.Token == token.NEWLINE {
			p.isBuf = true
		}
		return item, err
	}

	// Fill the buffer with the next 4 tokens, or until an EOF or
	// newline token has been received.

	p.isBuf = false
	p.pos = 0
	p.bufLen = 0

	i := 0
	for {
		item, err := p.scanner.Next()
		if err == io.EOF {
			return p.next()
		}
		if err != nil && err != io.EOF {
			return item, err
		}

		p.buf[i] = item
		p.bufLen = i + 1

		i++
		if item.Token == token.NEWLINE || i == bufN {
			break
		}
	}

	checkLen := p.bufLen

	// if the last token in the items is a newline, remove it
	if p.buf[p.bufLen-1].Token == token.NEWLINE {
		checkLen--
	}

	// Check for TXT, TAG, TXT
	if checkLen == 3 &&
		p.buf[0].Token == token.TEXT &&
		p.buf[1].Token == token.TAG &&
		p.buf[2].Token == token.TEXT &&
		hasStandaloneTagSymbol(p.buf[1]) &&
		isWhitespace(p.buf[0].Text) &&
		isWhitespace(p.buf[2].Text) {
		p.buf[0] = p.buf[1]
		p.bufLen = 1

		// Check for TXT, TAG
	} else if checkLen == 2 &&
		p.buf[0].Token == token.TEXT &&
		p.buf[1].Token == token.TAG &&
		hasStandaloneTagSymbol(p.buf[1]) &&
		isWhitespace(p.buf[0].Text) {
		p.buf[0] = p.buf[1]
		p.bufLen = 1

		// Check for TAG, TXT, NL
	} else if checkLen == 2 &&
		p.buf[0].Token == token.TAG &&
		p.buf[1].Token == token.TEXT &&
		hasStandaloneTagSymbol(p.buf[0]) &&
		isWhitespace(p.buf[1].Text) {
		p.bufLen = 1
	}
	return p.next()
}

func (p *parser) parse(parent parentNode, startPos int) error {
	for {
		item, err := p.scanner.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch item.Token {
		case token.TEXT:
			parent.add(&Text{Text: item.Text})
		case token.NEWLINE:
			parent.add(&Newline{Text: item.Text})
		case token.TAG:
			symbol := item.Text[len(item.LDelim)]
			switch symbol {
			case '{':
				key, err := p.parseDottedKey(item.Text[len(item.LDelim)+1 : len(item.Text)-len(item.RDelim)-1])
				if err != nil {
					return err
				}
				parent.add(&VariableTag{
					Key:       key,
					Unescaped: true,
				})

			case '&':
				key, err := p.parseDottedKey(item.Text[len(item.LDelim)+1 : len(item.Text)-len(item.RDelim)])
				if err != nil {
					return err
				}
				parent.add(&VariableTag{
					Key:       key,
					Unescaped: true,
				})

			case '#':
				rawKey := item.Text[len(item.LDelim)+1 : len(item.Text)-len(item.RDelim)]
				key, err := p.parseDottedKey(rawKey)
				if err != nil {
					return err
				}
				node := &SectionTag{
					Key:      key,
					Inverted: false,
					LDelim:   item.LDelim,
					RDelim:   item.RDelim,
				}
				err = p.parse(node, item.EndPos)
				if err == io.EOF {
					return p.error(item.StartLn, item.StartCol, "unclosed section tag: "+rawKey)
				}
				if err != nil {
					return err
				}
				parent.add(node)

			case '^':
				rawKey := item.Text[len(item.LDelim)+1 : len(item.Text)-len(item.RDelim)]
				key, err := p.parseDottedKey(rawKey)
				if err != nil {
					return err
				}
				node := &SectionTag{
					Key:      key,
					Inverted: true,
					LDelim:   item.LDelim,
					RDelim:   item.RDelim,
				}
				err = p.parse(node, item.EndPos)
				if err == io.EOF {
					return p.error(item.StartLn, item.StartCol, "unclosed section tag: "+rawKey)
				}
				if err != nil {
					return err
				}
				parent.add(node)

			case '>':
				key, err := p.parseKey(item.Text[len(item.LDelim)+1 : len(item.Text)-len(item.RDelim)])
				if err != nil {
					return err
				}
				parent.add(&PartialTag{
					Key:    key,
					Indent: item.Indent,
				})

			case '!':
				// skip comments
			case '=':
				// skip set delim tag
			default:
				key, err := p.parseDottedKey(item.Text[len(item.LDelim)+1 : len(item.Text)-len(item.RDelim)])
				if err != nil {
					return err
				}
				parent.add(&VariableTag{
					Key:       key,
					Unescaped: false,
				})
			}
		}
	}
}

func (p *parser) parseDottedKey(raw string) ([]string, error) {
	return strings.Split(strings.TrimSpace(raw), "."), nil
}

func (p *parser) parseKey(raw string) (string, error) {
	return strings.TrimSpace(raw), nil
}

func (p *parser) error(ln, col int, msg string) error {
	var b strings.Builder
	b.WriteString(p.scanner.Name())
	b.WriteString(":")
	b.WriteString(strconv.Itoa(ln))
	b.WriteString(":")
	b.WriteString(strconv.Itoa(col))
	b.WriteString(":")
	b.WriteString(" ")
	b.WriteString(msg)
	return errors.New(b.String())
}

func hasStandaloneTagSymbol(item token.Item) bool {
	if item.Token != token.TAG {
		return false
	}
	switch item.Text[len(item.LDelim)] {
	case '#', '^', '/', '>', '=', '!':
		return true
	default:
		return false
	}
}

func isWhitespace(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return false
		}
	}
	return true
}
