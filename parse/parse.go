package parse

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/mreriklott/mustache/token"
)

func Parse(r token.Reader) (tree *Tree, err error) {
	// catch parse errors
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			if _, ok = err.(runtime.Error); ok {
				panic(err)
			}
		}
	}()
	nodes := []Node{}
	for {
		node, eof := parseNode(r)
		if eof {
			break
		}
		nodes = append(nodes, node)
	}
	nodes = parseLines(nodes)
	nodes = parseSections(nodes)
	return &Tree{Nodes: nodes}, nil
}

// parseSections organizes a flat stream of nodes into a nested structure
func parseSections(nodes []Node) []Node {
	out := []Node{}
	openSections := []*SectionTag{}
	for _, node := range nodes {
		if len(openSections) > 0 {
			section := openSections[len(openSections)-1]
			if sectionEnd, ok := node.(*SectionEndTag); ok {
				if section.Key != sectionEnd.Key {
					parseError(fmt.Errorf("%s unexpected unopenned section tag '%s' in input", sectionEnd.Pos, sectionEnd.Key))
				}
				section.EndTag = sectionEnd
				openSections = openSections[:len(openSections)-1]
				continue
			}
			section.Nodes = append(section.Nodes, node)
		} else {
			if sectionEnd, ok := node.(*SectionEndTag); ok {
				parseError(fmt.Errorf("%s unexpected unopenned section tag '%s' in input", sectionEnd.Pos, sectionEnd.Key))
			}
			out = append(out, node)
		}
		if section, ok := node.(*SectionTag); ok {
			openSections = append(openSections, section)
		}
	}
	if len(openSections) > 0 {
		section := openSections[len(openSections)-1]
		parseError(fmt.Errorf("%s unexpected unclosed section tag '%s' in input", section.Pos, section.Key))
	}
	return out
}

// parseLines detects standalone lines and removes neseccary nodes according
// to mustache spec
func parseLines(nodes []Node) []Node {
	line := []Node{}
	out := []Node{}
	for _, node := range nodes {
		line = append(line, node)
		if isEOLNode(node) {
			out = append(out, parseLine(line)...)
			line = []Node{}
		}
	}
	if len(line) > 0 {
		out = append(out, parseLine(line)...)
	}
	return out
}

// parseLine determines if the nodes argument represents a standalone line,
// and removes the neseccary nodes according to mustache spec
func parseLine(nodes []Node) []Node {
	hasContent := false
	standaloneTags := []int{}
	for i, node := range nodes {
		if isContentNode(node) {
			hasContent = true
		}
		if isStandaloneTag(node) {
			standaloneTags = append(standaloneTags, i)
		}
	}
	if !hasContent && len(standaloneTags) == 1 {
		tagIndex := standaloneTags[0]
		hasPreceeding := (tagIndex > 0)
		tag := nodes[tagIndex]
		if hasPreceeding {
			for j := 0; j < tagIndex; j++ {
				pad := getNodePadding(nodes[j])
				indentNode(tag, pad)
			}
		}
		nodes = []Node{tag}
	}
	return nodes
}

// isContentNode returns true is this node is a content node. A content node
// is any node that would not be present on a standalone line.
func isContentNode(node Node) bool {
	switch node.(type) {
	case *Text, *VariableTag:
		return true
	default:
		return false
	}
}

// isStandaloneTag returns true if the node is a potential standalone tag
func isStandaloneTag(node Node) bool {
	switch node.(type) {
	case *CommentTag, *SetDelimsTag, *SectionTag, *SectionEndTag, *PartialTag:
		return true
	default:
		return false
	}
}

// isEOLNode returns true if this node represents an end of line
func isEOLNode(node Node) bool {
	_, ok := node.(*NewLine)
	return ok
}

// getNodePadding returns padding a an int, if this node has, or is, padding
func getNodePadding(node Node) int {
	switch node := node.(type) {
	case *Pad:
		return len(node.Text)
	default:
		return 0
	}
}

// indentNode add indent to this node, if the node can receive an indent
func indentNode(node Node, n int) {
	if tag, ok := node.(*PartialTag); ok {
		tag.Indent = n
	}
}

// parseNode parses a stream of tokens into a node.
func parseNode(r token.Reader) (Node, bool) {
	var node Node
	var eof bool
	switch tok := nextToken(r); tok {
	case token.EOF:
		eof = true
	case token.EOL:
		node = &NewLine{
			Text: []byte(r.Text()),
			Pos:  r.Pos(),
		}
	case token.TEXT:
		node = &Text{
			Text: []byte(r.Text()),
			Pos:  r.Pos(),
		}
	case token.PAD:
		node = &Pad{
			Text: []byte(r.Text()),
			Pos:  r.Pos(),
		}
	case token.LDELIM:
		lDelim := r.Text()
		pos := r.Pos()

		// advance next token
		switch tok = nextToken(r); tok {
		case token.LUNESC:
			var key string
			tok = nextToken(r)
			tok = skipWhitespaceToken(r, tok)
			key, tok = parseTagKey(r, tok)
			tok = skipWhitespaceToken(r, tok)
			acceptToken(r, tok, token.RUNESC, "tag")
			tok = nextToken(r)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &VariableTag{
				Key:       key,
				LDelim:    lDelim,
				RDelim:    rDelim,
				Unescaped: true,
				Pos:       pos,
			}
		case token.SECTION:
			var key string
			tok = nextToken(r)
			tok = skipWhitespaceToken(r, tok)
			key, tok = parseTagKey(r, tok)
			tok = skipWhitespaceToken(r, tok)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &SectionTag{
				Key:      key,
				LDelim:   lDelim,
				RDelim:   rDelim,
				Inverted: false,
				Nodes:    []Node{},
				Pos:      pos,
			}
		case token.ISECTION:
			var key string
			tok = nextToken(r)
			tok = skipWhitespaceToken(r, tok)
			key, tok = parseTagKey(r, tok)
			tok = skipWhitespaceToken(r, tok)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &SectionTag{
				Key:      key,
				LDelim:   lDelim,
				RDelim:   rDelim,
				Inverted: true,
				Nodes:    []Node{},
				Pos:      pos,
			}
		case token.SECTIONEND:
			var key string
			tok = nextToken(r)
			tok = skipWhitespaceToken(r, tok)
			key, tok = parseTagKey(r, tok)
			tok = skipWhitespaceToken(r, tok)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &SectionEndTag{
				Key:    key,
				LDelim: lDelim,
				RDelim: rDelim,
				Pos:    pos,
			}
		case token.PARTIAL:
			var key string
			tok = nextToken(r)
			tok = skipWhitespaceToken(r, tok)
			key, tok = parseTagKey(r, tok)
			tok = skipWhitespaceToken(r, tok)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &PartialTag{
				Key:    key,
				LDelim: lDelim,
				RDelim: rDelim,
				Pos:    pos,
			}
		case token.SETDELIM:
			var text string
			tok = nextToken(r)
			acceptToken(r, tok, token.TEXT, "tag")
			text = r.Text()
			tok = nextToken(r)
			acceptToken(r, tok, token.SETDELIM, "tag")
			tok = nextToken(r)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &SetDelimsTag{
				Text:   text,
				LDelim: lDelim,
				RDelim: rDelim,
				Pos:    pos,
			}
		case token.UNESC:
			var key string
			tok = nextToken(r)
			tok = skipWhitespaceToken(r, tok)
			key, tok = parseTagKey(r, tok)
			tok = skipWhitespaceToken(r, tok)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &VariableTag{
				Key:       key,
				LDelim:    lDelim,
				RDelim:    rDelim,
				Unescaped: true,
				Pos:       pos,
			}
		case token.COMMENT:
			var text string
			tok = nextToken(r)
			acceptToken(r, tok, token.TEXT, "tag")
			text = r.Text()
			tok = nextToken(r)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &CommentTag{
				Text:   text,
				LDelim: lDelim,
				RDelim: rDelim,
				Pos:    pos,
			}
		default:
			var key string
			tok = skipWhitespaceToken(r, tok)
			key, tok = parseTagKey(r, tok)
			tok = skipWhitespaceToken(r, tok)
			acceptToken(r, tok, token.RDELIM, "tag")
			rDelim := r.Text()
			node = &VariableTag{
				Key:       key,
				LDelim:    lDelim,
				RDelim:    rDelim,
				Unescaped: false,
				Pos:       pos,
			}
		}
	default:
		unexpectedToken(r, "input")
	}
	return node, eof
}

// parseTagKey parses a sequence of nodes that represent a tag identifier
func parseTagKey(r token.Reader, tok token.Token) (string, token.Token) {
	var accum bytes.Buffer
	switch tok {
	case token.DOT:
		accum.WriteString(r.Text())
		next := nextToken(r)
		return accum.String(), next
	case token.KEY:
		accum.WriteString(r.Text())
		for {
			tok = nextToken(r)
			if tok != token.DOT {
				return accum.String(), tok
			}
			accum.WriteString(r.Text())
			tok = nextToken(r)
			acceptToken(r, tok, token.KEY, "tag")
			accum.WriteString(r.Text())
		}
	default:
		unexpectedToken(r, "tag")
	}
	// impossible to reach this point
	return "", token.ILLEGAL
}

// nextToken returns the next token from the token reader
func nextToken(r token.Reader) token.Token {
	tok, err := r.Next()
	if err != nil {
		parseError(err)
	}
	return tok
}

// acceptToken generates a parse error (panic) if the actual token doesn't match the acceptable token
func acceptToken(r token.Reader, actual token.Token, accept token.Token, context string) {
	if actual != accept {
		unexpectedToken(r, context)
	}
}

// skipWhitespaceToken returns the next non-whitespace token
func skipWhitespaceToken(r token.Reader, tok token.Token) token.Token {
	for {
		if tok != token.WS {
			return tok
		}
		tok = nextToken(r)
	}
}

// unexpectedToken generates an error
func unexpectedToken(r token.Reader, context string) {
	parseError(fmt.Errorf("%s  unexpected %s in %s", r.Pos(), r.Text(), context))
}

// generates an error which is panicced
func parseError(err error) {
	panic(err)
}
