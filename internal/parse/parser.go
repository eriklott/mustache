package parse

import (
	"errors"
	"fmt"
	"strings"
)

// Default Delimeters
const (
	DefaultLeftDelim  = "{{"
	DefaultRightDelim = "}}"
)

var (
	errEOL = errors.New("end of line")
	errEOF = errors.New("end of file")
)

type parser struct {
	name      string
	src       string
	pos       int
	col       int
	ln        int
	ldelim    string
	rdelim    string
	isNewLine bool
}

type parentNode interface {
	add(Node)
}

// Parse transforms a template string into a tree
func Parse(name, src, ldelim, rdelim string) (*Tree, error) {
	p := &parser{
		name:      name,
		src:       src,
		pos:       0,
		col:       1,
		ln:        1,
		ldelim:    ldelim,
		rdelim:    rdelim,
		isNewLine: true,
	}
	tree := &Tree{}
	err := p.parse(tree, 0)
	return tree, err
}

func (p *parser) error(ln, col int, msg string) error {
	return fmt.Errorf("%s:%d:%d: %s", p.name, ln, col, msg)
}

func (p *parser) readTo(pattern string, haltEOL bool) (string, error) {
	start := p.pos
	matchIdx := 0
	TextLen := len(p.src)
	patLen := len(pattern)
	for {
		if p.pos >= TextLen {
			return p.src[start:p.pos], errEOF
		}
		b := p.src[p.pos]
		p.pos++
		p.col++
		if b == '\n' {
			p.col = 1
			p.ln++
			if haltEOL {
				return p.src[start:p.pos], errEOL
			}
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

type textResult struct {
	text     string
	startPos int
	startCol int
	startLn  int
	endPos   int
}

func (p *parser) readText() (textResult, error) {
	startPos := p.pos
	startCol := p.col
	startLn := p.ln

	text, err := p.readTo(p.ldelim, true)
	if err != nil {
		result := textResult{
			text:     text,
			startPos: startPos,
			startCol: startCol,
			startLn:  startLn,
			endPos:   p.pos,
		}
		return result, err
	}

	p.pos -= len(p.ldelim)
	p.col -= len(p.ldelim)

	result := textResult{
		text:     text[:len(text)-len(p.ldelim)],
		startPos: startPos,
		startCol: startCol,
		startLn:  startLn,
		endPos:   p.pos,
	}
	return result, nil
}

type tagResult struct {
	text     string
	symbol   byte
	startPos int
	startCol int
	startLn  int
	endPos   int
}

func (p *parser) readTag() (tagResult, error) {
	startPos := p.pos
	startCol := p.col
	startLn := p.ln

	p.pos += len(p.ldelim)
	p.col += len(p.ldelim)

	var symbol byte
	if p.pos < len(p.src) {
		symbol = p.src[p.pos]
	}

	var err error
	switch symbol {
	case '{':
		_, err = p.readTo("}"+p.rdelim, false)
	case '=':
		_, err = p.readTo("="+p.rdelim, false)
	default:
		_, err = p.readTo(p.rdelim, false)
	}
	if err != nil {
		return tagResult{}, p.error(startLn, startCol, "unclosed tag")
	}

	result := tagResult{
		text:     p.src[startPos:p.pos],
		symbol:   symbol,
		startPos: startPos,
		startCol: startCol,
		startLn:  startLn,
		endPos:   p.pos,
	}
	return result, nil
}

func (p *parser) parse(parent parentNode, sectionStart int) error {
	for {
		// read text
		text, err := p.readText()
		if err == errEOL {
			p.isNewLine = true
			parent.add(&Text{
				Text:      text.text,
				EndOfLine: true,
			})
			continue
		}
		if err == errEOF {
			// If the eof has been reached while parsing the inside of a section, return
			// the eof error to the calling function so the parent section tag can report
			// this error using it's col & line number.
			if _, ok := parent.(*SectionTag); ok {
				return err
			}

			if len(text.text) > 0 {
				parent.add(&Text{
					Text: text.text,
				})
			}
			return nil
		}

		// read tag
		tag, err := p.readTag()
		if err != nil {
			return err
		}

		isStandaloneTag := false
		if p.isNewLine {
			p.isNewLine = false
			if isStandaloneTagSymbol(tag.symbol) {

				// Starting at the tag's left delim, scan backwards to check
				// if the line of text preceeding the tag consists only of whitespace
				// since the last newline character (or the beginning of the template string).
				// If is does, seperate it from the text as padding.
				hasLeftWhitespace := isSpaces(text.text)
				if hasLeftWhitespace {
					// Starting at the tag's right delim, scan forwards to check
					// if the text following the tag consists only of whitespace until the
					// next newline char, or the end of template string.
					// If is does, set the cursor position after the whitespace (and newline chars), and
					// set the standalone = true.
					i := tag.endPos
					for {
						// reached eof
						if i >= len(p.src) {
							isStandaloneTag = true
							p.pos = i
							p.col += i - tag.endPos
							break
						}

						b := p.src[i]
						i++

						// reached eol
						if b == '\n' {
							isStandaloneTag = true
							p.pos = i
							p.col = 1
							p.ln++
							p.isNewLine = true
							break
						}

						if !isSpace(b) && b != '\r' {
							break
						}
					}
				}
			}
		}
		// add text node to parent
		if !isStandaloneTag && len(text.text) > 0 {
			parent.add(&Text{
				Text: text.text,
			})
		}

		switch tag.symbol {
		case '{':
			rawKey := tag.text[len(p.ldelim)+1 : len(tag.text)-len(p.rdelim)-1]
			rawKey = trimSpace(rawKey)
			key, err := p.parseDottedKey(tag.startLn, tag.startCol, rawKey)
			if err != nil {
				return err
			}

			parent.add(&VariableTag{
				Key:       key,
				Unescaped: true,
			})

		case '&':
			rawKey := tag.text[len(p.ldelim)+1 : len(tag.text)-len(p.rdelim)]
			rawKey = trimSpace(rawKey)
			key, err := p.parseDottedKey(tag.startLn, tag.startCol, rawKey)
			if err != nil {
				return err
			}

			parent.add(&VariableTag{
				Key:       key,
				Unescaped: true,
			})

		case '#':
			rawKey := tag.text[len(p.ldelim)+1 : len(tag.text)-len(p.rdelim)]
			rawKey = trimSpace(rawKey)
			key, err := p.parseDottedKey(tag.startLn, tag.startCol, rawKey)
			if err != nil {
				return err
			}

			node := &SectionTag{
				Key:      key,
				Inverted: false,
				LDelim:   p.ldelim,
				RDelim:   p.rdelim,
			}
			err = p.parse(node, tag.endPos)
			if err == errEOF {
				return p.error(tag.startLn, tag.startCol, "unclosed section tag: "+rawKey)
			}
			if err != nil {
				return err
			}
			parent.add(node)

		case '^':
			rawKey := tag.text[len(p.ldelim)+1 : len(tag.text)-len(p.rdelim)]
			rawKey = trimSpace(rawKey)
			key, err := p.parseDottedKey(tag.startLn, tag.startCol, rawKey)
			if err != nil {
				return err
			}

			node := &SectionTag{
				Key:      key,
				Inverted: true,

				LDelim: p.ldelim,
				RDelim: p.rdelim,
			}
			err = p.parse(node, tag.endPos)
			if err == errEOF {
				return p.error(tag.startLn, tag.startCol, "unclosed section tag: "+rawKey)
			}
			if err != nil {
				return err
			}
			parent.add(node)

		case '/':
			rawKey := tag.text[len(p.ldelim)+1 : len(tag.text)-len(p.rdelim)]
			rawKey = trimSpace(rawKey)
			err := p.validateDottedKey(tag.startLn, tag.startCol, rawKey)
			if err != nil {
				return err
			}

			section, ok := parent.(*SectionTag)
			if !ok || strings.Join(section.Key, ".") != rawKey {
				return p.error(tag.startLn, tag.startCol, "unexpected section closing tag: "+rawKey)
			}
			section.ChildrenText = p.src[sectionStart:tag.startPos]

			return nil

		case '>':
			key := tag.text[len(p.ldelim)+1 : len(tag.text)-len(p.rdelim)]
			key = trimSpace(key)
			err = p.validateKey(tag.startLn, tag.startCol, key)
			if err != nil {
				return err
			}

			if isStandaloneTag {
				parent.add(&PartialTag{
					Key:    key,
					Indent: text.text,
				})
			} else {
				parent.add(&PartialTag{
					Key:    key,
					Indent: "",
				})
			}

		case '!':
			// Skip comments

		case '=':
			delims := tag.text[len(p.ldelim)+1 : len(tag.text)-len(p.rdelim)-1]
			delims = trimSpace(delims)
			parts := strings.Split(delims, " ")
			if len(parts) == 2 {
				p.ldelim = parts[0]
				p.rdelim = parts[1]
			}

		default:
			rawKey := tag.text[len(p.ldelim) : len(tag.text)-len(p.rdelim)]
			rawKey = trimSpace(rawKey)
			key, err := p.parseDottedKey(p.ln, p.col, rawKey)
			if err != nil {
				return err
			}

			parent.add(&VariableTag{
				Key: key,
			})
		}
	}
}

func (p *parser) parseDottedKey(ln, col int, key string) ([]string, error) {
	err := p.validateDottedKey(ln, col, key)
	if err != nil {
		return nil, err
	}
	if key == "." {
		return []string{"."}, nil
	}
	return strings.Split(key, "."), nil
}

func (p *parser) validateDottedKey(ln, col int, key string) error {
	if len(key) == 0 {
		return p.error(ln, col, "missing key")
	}

	if key == "." {
		return nil
	}

	isValid := false
Loop:
	for i := range key {
		switch b := key[i]; {
		case isKeyChar(b):
			isValid = true
		case b == '.':
			isValid = false
		default:
			isValid = false
			break Loop
		}
	}
	if !isValid {
		return p.error(ln, col, "invalid key: "+key)
	}
	return nil
}

func (p *parser) validateKey(ln, col int, key string) error {
	if len(key) == 0 {
		return p.error(ln, col, "missing key")
	}

	for i := range key {
		if !isKeyChar(key[i]) {
			return p.error(ln, col, "invalid key: "+key)
		}
	}
	return nil
}

func isKeyChar(b byte) bool {
	switch b {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

func isSpace(b byte) bool {
	switch b {
	case ' ', '\t':
		return true
	default:
		return false
	}
}

func isSpaces(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isSpace(s[i]) {
			return false
		}
	}
	return true
}

func isStandaloneTagSymbol(b byte) bool {
	switch b {
	case '#', '^', '/', '>', '=', '!':
		return true
	default:
		return false
	}
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
