// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

import (
	"errors"
	"fmt"

	"github.com/eriklott/mustache/internal/ast"
	"github.com/eriklott/mustache/internal/parser"
	"github.com/eriklott/mustache/internal/render"
)

type Template struct {
	treeMap map[string]*ast.Tree
}

func NewTemplate() *Template {
	return &Template{
		treeMap: make(map[string]*ast.Tree),
	}
}

func (t *Template) Parse(name, text string) error {
	tree, err := parser.Parse(text)
	if err != nil {
		return errors.New(name + ":" + err.Error())
	}
	t.treeMap[name] = tree
	return nil
}

func (t *Template) Render(name string, contexts ...interface{}) (string, error) {
	tree, ok := t.treeMap[name]
	if !ok {
		return "", fmt.Errorf("template not found: %s", name)
	}
	s := render.Render(tree, t.treeMap, contexts)
	return s, nil
}

func Parse(name, text string) (*Template, error) {
	t := NewTemplate()
	err := t.Parse(name, text)
	return t, err
}

func Render(text string, contexts ...interface{}) (string, error) {
	t, err := Parse("main", text)
	if err != nil {
		return "", err
	}
	return t.Render("main", contexts...)
}
