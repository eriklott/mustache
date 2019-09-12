// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

import (
	"fmt"
	"reflect"

	"github.com/eriklott/mustache/internal/parse"
	"github.com/eriklott/mustache/internal/render"
)

type Template struct {
	treeMap map[string]*parse.Tree
}

func NewTemplate() *Template {
	return &Template{
		treeMap: make(map[string]*parse.Tree),
	}
}

func (t *Template) Parse(name, text string) error {
	tree, err := parse.Parse(name, text, parse.DefaultLeftDelim, parse.DefaultRightDelim)
	if err != nil {
		return err
	}
	t.treeMap[name] = tree
	return nil
}

func (t *Template) Render(name string, contexts ...interface{}) (string, error) {
	var reversedContexts []reflect.Value
	for i := len(contexts) - 1; i >= 0; i-- {
		c := reflect.ValueOf(contexts[i])
		reversedContexts = append(reversedContexts, c)
	}

	tree, ok := t.treeMap[name]
	if !ok {
		return "", fmt.Errorf("template not found: %s", name)
	}
	s := render.Render(tree, t.treeMap, reversedContexts)
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
