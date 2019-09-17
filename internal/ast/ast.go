// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ast

import "strings"

type NodeType int

const (
	Tree NodeType = iota
	Text
	TextEOF
	Variable
	UnescapedVariable
	Section
	InvertedSection
	Partial
)

type Node struct {
	Type   NodeType
	V1     string
	V2     string
	V3     string
	Line   int
	Column int
	Nodes  []Node
}

func (n Node) IsTree() bool {
	return n.Type == Tree
}

func (n Node) IsText() bool {
	return n.Type == Text || n.Type == TextEOF
}

func (n Node) IsEndOfLine() bool {
	return n.Type == TextEOF
}

func (n Node) IsVariable() bool {
	return n.Type == Variable || n.Type == UnescapedVariable
}

func (n Node) IsUnescaped() bool {
	return n.Type == UnescapedVariable
}

func (n Node) IsSection() bool {
	return n.Type == Section || n.Type == InvertedSection
}

func (n Node) IsInverted() bool {
	return n.Type == InvertedSection
}

func (n Node) IsPartial() bool {
	return n.Type == Partial
}

// Text

func (n Node) Text() string {
	return n.V1
}

func (n Node) Key() string {
	return n.V1
}

func (n Node) Name() string {
	return n.V1
}

func (n Node) Delims() (string, string) {
	parts := strings.Split(n.V2, " ")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func (n Node) SectionText() string {
	return n.V3
}

func (n Node) Indent() string {
	return n.V2
}
