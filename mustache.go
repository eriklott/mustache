// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

import "text/template/parse"

type Template struct {
	treeMap   map[string]*parse.Tree
	renderErr error
}

func NewTemplate() *Template {
	return &Template{
		treeMap: map[string]*parse.Tree{},
	}
}

func (t *Template) Parse(name string, text string) error {
	return nil
}

func (t *Template) Render(name string, ctx ...interface{}) string {
	return ""
}
