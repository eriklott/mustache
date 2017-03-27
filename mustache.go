// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

import (
	"io"
	"strings"

	"github.com/eriklott/mustache/parse"
	"github.com/eriklott/mustache/token"
)

// Template is the representation of a parsed mustache template.
type Template struct {
	treeMap   map[string]*parse.Tree
	renderErr error
}

// NewTemplate returns a new template instance
func NewTemplate() *Template {
	return &Template{
		treeMap: map[string]*parse.Tree{},
	}
}

// Parse parses the mustache input as a named partial. Base (or root) templates,
// as well as partials, are all considered 'partials' by this library, and must
// each be added via the Parse method. Partial names must be alphanumeric. If an
// error is returned, the mustache source has not been added to the template.
func (t *Template) Parse(name string, r io.Reader) error {
	tokenReader := token.NewReader("", r, token.DefaultLeftDelim, token.DefaultRightDelim)
	tree, err := parse.Parse(tokenReader)
	if err != nil {
		return err
	}
	t.treeMap[name] = tree
	return nil
}

// ParseString functions similarly to Parse, but accepts
// a string argument instead of an io.Reader.
func (t *Template) ParseString(name, s string) error {
	return t.Parse(name, strings.NewReader(s))
}
