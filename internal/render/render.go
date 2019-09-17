// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package render

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/eriklott/mustache/internal/ast"
	"github.com/eriklott/mustache/internal/parse"
)

const maxPartialDepth = 100000

// renderer represents the state of the rendering of a single template.
type renderer struct {
	b                strings.Builder
	indent           string
	indentNext       bool
	contexts         []reflect.Value
	treeMap          map[string]*ast.Tree
	depth            int //the height of the stack of executing partials
	errOnContextMiss bool
}

// Render transforms a tree of nodes into a rendered string template.
func Render(tree *ast.Tree, treeMap map[string]*ast.Tree, contexts []interface{}) (string, error) {
	var reversedContexts []reflect.Value
	for i := len(contexts) - 1; i >= 0; i-- {
		c := reflect.ValueOf(contexts[i])
		reversedContexts = append(reversedContexts, c)
	}
	return render(tree, treeMap, reversedContexts)
}

// render transforms a template tree into a final rendered string template. The context stack
// provided should be in reverse order.
func render(tree *ast.Tree, treeMap map[string]*ast.Tree, contexts []reflect.Value) (string, error) {
	r := &renderer{
		indentNext: true,
		contexts:   contexts,
		treeMap:    treeMap,
	}
	err := r.renderNode(tree.Name, tree)
	return r.b.String(), err
}

// write a string to the template output.
func (r *renderer) write(s string) {
	if r.indentNext {
		r.indentNext = false
		r.b.WriteString(r.indent)
	}
	r.b.WriteString(s)
}

// write an html escaped string the the template output.
func (r *renderer) escapedWrite(s string) {
	r.write(html.EscapeString(s))
}

// increaseIndent concatenate an string of whitespace to any existing indent.
func (r *renderer) increaseIndent(s string) {
	r.indent += s
}

// decreaseIndent removes a previously appended indent.
func (r *renderer) decreaseIndent(i string) {
	r.indent = r.indent[:len(r.indent)-len(i)]
}

// indentNextWrite instructs the writer to add an indent before writing the next
// string to the template output.
func (r *renderer) indentNextWrite() {
	r.indentNext = true
}

// conceptually shifts a context onto the stack. Since the stack is actually in
// reverse order, the context is pushed.
func (r *renderer) addContext(context reflect.Value) {
	r.contexts = append(r.contexts, context)
}

// conceptually unshifts a context onto the stack. Since the stack is actually in
// reverse order, the context is popped.
func (r *renderer) removeContext() {
	r.contexts = r.contexts[:len(r.contexts)-1]
}

// render recursively walks each node of the tree, incrementally building the template
// string output.
func (r *renderer) renderNode(treeName string, node interface{}) error {
	switch t := node.(type) {
	case *ast.Tree:
		for i := range t.Nodes {
			err := r.renderNode(treeName, t.Nodes[i])
			if err != nil {
				return err
			}
		}

	case *ast.Text:
		r.write(t.Text)
		if t.EndOfLine {
			r.indentNextWrite()
		}

	case *ast.Variable:
		v, err := r.lookup(treeName, t.Line, t.Column, t.Key)
		if err != nil {
			return err
		}
		s, err := r.toString(v, parse.DefaultLeftDelim, parse.DefaultRightDelim)
		if err != nil {
			return err
		}
		if t.Unescaped {
			r.write(s)
		} else {
			r.escapedWrite(s)
		}

	case *ast.Section:
		v, err := r.lookup(treeName, t.Line, t.Column, t.Key)
		if err != nil {
			return err
		}
		v = indirect(v)
		isTruthy, err := r.isSectionTruthy(v)
		if err != nil {
			return err
		}
		if !t.Inverted && isTruthy {
			switch v.Kind() {
			case reflect.Slice, reflect.Array:
				for i := 0; i < v.Len(); i++ {
					r.addContext(v.Index(i))
					for j := range t.Nodes {
						err := r.renderNode(treeName, t.Nodes[j])
						if err != nil {
							return err
						}
					}
					r.removeContext()
				}
			case reflect.Func:
				s := v.Call([]reflect.Value{reflect.ValueOf(t.Text)})[0].String()
				tree, err := parse.Parse("lambda", s, t.LDelim, t.RDelim)
				if err != nil {
					return nil
				}
				err = r.renderNode(treeName, tree)
				if err != nil {
					return err
				}

			default:
				r.addContext(v)
				for i := range t.Nodes {
					err := r.renderNode(treeName, t.Nodes[i])
					if err != nil {
						return err
					}
				}
				r.removeContext()
			}
		} else if t.Inverted && !isTruthy {
			for i := range t.Nodes {
				err := r.renderNode(treeName, t.Nodes[i])
				if err != nil {
					return err
				}
			}
		}

	case *ast.Partial:
		tree, ok := r.treeMap[t.Key]
		if !ok && r.errOnContextMiss {
			return nil
		}
		r.increaseIndent(t.Indent)
		r.indentNextWrite()
		r.depth++
		if r.depth == maxPartialDepth {
			return fmt.Errorf("exceeded maximum partial depth: %d", maxPartialDepth)
		}
		err := r.renderNode(tree.Name, tree)
		if err != nil {
			return err
		}
		r.depth--
		r.decreaseIndent(t.Indent)
	}
	return nil
}

// lookup a key in the context stack. If a value was not found, the reflect zero
// type is returned.
func (r *renderer) lookup(name string, ln, col int, key []string) (reflect.Value, error) {
	v := lookupKeysContexts(key, r.contexts)
	if !v.IsValid() && r.errOnContextMiss {
		return v, fmt.Errorf("%s:%d:%d: cannot find value %s in context", name, ln, col, strings.Join(key, "."))
	}
	return v, nil
}

// lookupKeysContexts obtains a value for a dotted key - eg: a.b.c . If a value
// was not found, the reflect zero type is returned.
func lookupKeysContexts(key []string, contexts []reflect.Value) reflect.Value {
	var v reflect.Value

	if len(key) == 0 {
		return v
	}

	for i := range key {
		if i == 0 {
			v = lookupContexts(key[i], contexts)
			continue
		}
		v = lookup(key[i], v)
		if !v.IsValid() {
			break
		}
	}
	return v
}

// lookupContexts returns a value from the first context in the stack that
// contains a value for that key. If a value was not found, the reflect zero
// type is returned.
func lookupContexts(key string, contexts []reflect.Value) reflect.Value {
	var v reflect.Value
	for i := len(contexts) - 1; i >= 0; i-- {
		ctx := contexts[i]
		v = lookup(key, ctx)
		if v.IsValid() {
			break
		}
	}
	return v
}

// lookup returns a value by key from the context. If a value
// was not found, the reflect zero type is returned.
func lookup(key string, ctx reflect.Value) reflect.Value {
	if key == "." {
		return ctx
	}

	// check context for method by name
	if ctx.IsValid() {
		method := ctx.MethodByName(key)
		if method.IsValid() {
			return method
		}
	}

	// check for fields and keys on concrete types.
	switch ctx.Kind() {
	case reflect.Ptr, reflect.Interface:
		return lookup(key, indirect(ctx))
	case reflect.Map:
		return ctx.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		return ctx.FieldByName(key)
	default:
		return reflect.Value{}
	}
}

// toString transforms a reflect.Value into a string.
func (r *renderer) toString(v reflect.Value, ldelim, rdelim string) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", v.Complex()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Func:
		if v.IsNil() {
			return "", nil
		}

		t := v.Type()
		isArity0 := t.NumIn() == 0 && t.NumOut() == 1
		if !isArity0 {
			return "", nil
		}

		v = v.Call(nil)[0]
		if v.Kind() != reflect.String {
			return r.toString(v, ldelim, rdelim)
		}
		tree, err := parse.Parse("lambda", v.String(), ldelim, rdelim)
		if err != nil {
			return "", err
		}
		s, err := render(tree, r.treeMap, r.contexts)
		if err != nil {
			return "", err
		}
		return s, nil

	case reflect.Ptr, reflect.Interface:
		return r.toString(indirect(v), ldelim, rdelim)
	case reflect.Chan:
		return "", nil
	case reflect.Invalid:
		return "", nil
	default:
		return fmt.Sprintf("%v", v.Interface()), nil
	}
}

func indirect(v reflect.Value) reflect.Value {
loop:
	for v.IsValid() {
		switch av := v; av.Kind() {
		case reflect.Ptr:
			v = av.Elem()
		case reflect.Interface:
			v = av.Elem()
		default:
			break loop
		}
	}
	return v
}

// isSectionTruthy returns a value when the section is truthy. Returns the
// reflect zero value when the value is falsey.
func (r *renderer) isSectionTruthy(v reflect.Value) (bool, error) {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b := v.Int() != 0
		return b, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b := v.Uint() != 0
		return b, nil
	case reflect.Float32, reflect.Float64:
		b := math.Float64bits(v.Float()) != 0
		return b, nil
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		b := !(math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0)
		return b, nil
	case reflect.String:
		b := v.Len() != 0
		return b, nil
	case reflect.Array, reflect.Slice:
		b := !v.IsNil() && v.Len() > 0
		return b, nil
	case reflect.Func:
		if v.IsNil() {
			return false, nil
		}
		t := v.Type()
		isArity0 := t.NumIn() == 0 && t.NumOut() == 1
		if isArity0 {
			v = v.Call(nil)[0]
			if v.Kind() != reflect.String {
				return r.isSectionTruthy(v)
			}
			tree, err := parse.Parse("lambda", v.String(), parse.DefaultLeftDelim, parse.DefaultRightDelim)
			if err != nil {
				return false, nil
			}
			s, err := render(tree, r.treeMap, r.contexts)
			if err != nil {
				return false, err
			}
			return r.isSectionTruthy(reflect.ValueOf(s))
		}
		isArity1 := t.NumIn() == 1 && t.In(0).Kind() == reflect.String && t.NumOut() == 1 && t.Out(0).Kind() == reflect.String
		return isArity1, nil
	case reflect.Ptr, reflect.Interface:
		return r.isSectionTruthy(indirect(v))
	case reflect.Map:
		b := !v.IsNil()
		return b, nil
	case reflect.Struct:
		return true, nil
	case reflect.Invalid:
		return false, nil
	default:
		return false, nil
	}
}
