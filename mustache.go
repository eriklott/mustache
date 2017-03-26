package mustache

import (
	"io"

	"github.com/mreriklott/mustache/parse"
	"github.com/mreriklott/mustache/token"
)

type Template struct {
	treeMap map[string]*parse.Tree
}

func NewTemplate() *Template {
	return &Template{
		treeMap: map[string]*parse.Tree{},
	}
}

func (t *Template) Parse(name string, r io.Reader) error {
	tokenReader := token.NewReader("", r, token.DefaultLeftDelim, token.DefaultRightDelim)
	tree, err := parse.Parse(tokenReader)
	if err != nil {
		return err
	}
	t.treeMap[name] = tree
	return nil
}
