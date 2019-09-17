// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

import (
	"fmt"
	"reflect"

	"github.com/eriklott/mustache/internal/ast"
	"github.com/eriklott/mustache/internal/parse"
)

// Template is the representation of a parsed template.
type Template struct {
	treeMap              map[string]*ast.Tree
	ContextErrorsEnabled bool
}

// NewTemplate allocates a new template.
func NewTemplate() *Template {
	t := &Template{
		treeMap: make(map[string]*ast.Tree),
	}
	return t
}

// Parse adds a mustache string to the template, making it available to render by name via
// the Render method, or using a partial tag. If an error occurs during parsing, the parsing
// process stops, and the error is returned.
func (t *Template) Parse(name, text string) error {
	tree, err := parse.Parse(name, text, parse.DefaultLeftDelim, parse.DefaultRightDelim)
	if err != nil {
		return err
	}
	t.treeMap[name] = tree
	return nil
}

// Render applies a data context to a parsed template and returns the output as a string.
// If an error occurs, the rendering process stops and the error is returned.
func (t *Template) Render(name string, contexts ...interface{}) (string, error) {
	tree, ok := t.treeMap[name]
	if !ok {
		return "", fmt.Errorf("template not found: %s", name)
	}

	// init new renderer
	r := t.newRenderer()

	// push contexts onto stack
	for i := range contexts {
		context := reflect.ValueOf(contexts[i])
		r.push(context)
	}

	err := r.walk(tree.Name, tree)
	s := r.String()
	return s, err
}
