// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parse

import (
	"bytes"
	"fmt"

	"github.com/eriklott/mustache/token"
)

// Node represents a node in the parse tree.
type Node interface {
	String() string
}

// Tree serves as the root container node of the parse tree.
type Tree struct {
	Nodes []Node
}

func (t *Tree) String() string {
	var b bytes.Buffer
	for _, node := range t.Nodes {
		b.WriteString(node.String())
	}
	return b.String()
}

// Text represents the text template content that exists around and between
// mustache tags. Text does not contain new lines.
type Text struct {
	Text []byte
	Pos  token.Position
}

func (t *Text) String() string {
	return fmt.Sprintf("%s", t.Text)
}

// Pad represents empty padding before or after a tag or new line.
type Pad struct {
	Text []byte
	Pos  token.Position
}

func (p *Pad) String() string {
	return fmt.Sprintf("%s", p.Text)
}

// NewLine represents a new line char
type NewLine struct {
	Text []byte
	Pos  token.Position
}

func (n *NewLine) String() string {
	return fmt.Sprintf("%s", n.Text)
}

// VariableTag represents an escaped mustache variable tag.
type VariableTag struct {
	Key       string
	LDelim    string
	RDelim    string
	Unescaped bool
	Pos       token.Position
}

func (v *VariableTag) String() string {
	sym := ""
	if v.Unescaped {
		sym = "&"
	}
	return fmt.Sprintf("%s%s%s%s", v.LDelim, sym, v.Key, v.RDelim)
}

// SectionTag represents a mustache section tag.
type SectionTag struct {
	Key      string
	LDelim   string
	RDelim   string
	Inverted bool
	Nodes    []Node
	EndTag   *SectionEndTag
	Pos      token.Position
}

func (s *SectionTag) NodesString() string {
	var str bytes.Buffer
	for _, node := range s.Nodes {
		str.WriteString(node.String())
	}
	return str.String()
}

func (s *SectionTag) String() string {
	var str bytes.Buffer
	str.WriteString(s.LDelim)
	if s.Inverted {
		str.WriteString("^")
	} else {
		str.WriteString("#")
	}
	str.WriteString(s.Key)
	str.WriteString(s.RDelim)
	str.WriteString(s.NodesString())
	if s.EndTag != nil {
		str.WriteString(s.EndTag.String())
	}
	return str.String()
}

// SectionEndTag represents a section end mustache tag.
type SectionEndTag struct {
	Key    string
	LDelim string
	RDelim string
	Pos    token.Position
}

func (s *SectionEndTag) String() string {
	return fmt.Sprintf("%s/%s%s", s.LDelim, s.Key, s.RDelim)
}

// PartialTag represents a mustache partial tag.
type PartialTag struct {
	Key    string
	LDelim string
	RDelim string
	Indent int // indent size applied to each new line of the rendered partial
	Pos    token.Position
}

func (p *PartialTag) String() string {
	return fmt.Sprintf("%s>%s%s", p.LDelim, p.Key, p.RDelim)
}

// CommentTag represents a mustache comment tag.
type CommentTag struct {
	Text   string
	LDelim string
	RDelim string
	Pos    token.Position
}

func (t *CommentTag) String() string {
	return fmt.Sprintf("%s!%s%s", t.LDelim, t.Text, t.RDelim)
}

// SetDelimsTag represents a mustache set delimeters tag.
type SetDelimsTag struct {
	Text   string
	LDelim string
	RDelim string
	Pos    token.Position
}

func (t *SetDelimsTag) String() string {
	return fmt.Sprintf("%s=%s=%s", t.LDelim, t.Text, t.RDelim)
}
