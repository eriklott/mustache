package mustache

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
	template *Template       // the template that initiated the render
	stack    []reflect.Value // the context stack
	depth    int             // the depth of executing partials

	// write fields
	w          strings.Builder // the writer
	indent     string          // the current indent string
	indentNext bool            // when true, apply indent before next write
}

// newRenderer returns a newly initialized renderer.
func (t *Template) newRenderer() *renderer {
	return &renderer{template: t}
}

func (r *renderer) String() string {
	return r.w.String()
}

// renderToString sub-renders a tree into a string. If an error occurs,
// rendering stops and the error is returned.
func (r *renderer) renderToString(tree *ast.Tree) (string, error) {
	subRenderer := &renderer{
		template:   r.template,
		stack:      r.stack,
		depth:      0,
		w:          strings.Builder{},
		indent:     "",
		indentNext: false,
	}
	err := subRenderer.walk(tree.Name, tree)
	s := subRenderer.String()

	// the subRenderer may have pushed and popped enough contexts onto the stack
	// to cause the slice to allocate to a new larger underlaying array. If this
	// has happened, we want to keep the pointer to that larger array to minimize
	// allocations.
	r.stack = subRenderer.stack

	return s, err
}

// write a string to the template output.
func (r *renderer) write(s string, unescaped bool) {
	if r.indentNext {
		r.indentNext = false
		r.w.WriteString(r.indent)
	}
	if !unescaped {
		s = html.EscapeString(s)
	}
	r.w.WriteString(s)
}

// conceptually shifts a context onto the stack. Since the stack is actually in
// reverse order, the context is pushed.
func (r *renderer) push(context reflect.Value) {
	r.stack = append(r.stack, context)
}

// conceptually unshifts a context onto the stack. Since the stack is actually in
// reverse order, the context is popped.
func (r *renderer) pop() reflect.Value {
	if len(r.stack) == 0 {
		return reflect.Value{}
	}
	ctx := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	return ctx
}

// render recursively walks each node of the tree, incrementally building the template
// string output.
func (r *renderer) walk(treeName string, node interface{}) error {
	switch t := node.(type) {
	case *ast.Tree:
		for i := range t.Nodes {
			err := r.walk(treeName, t.Nodes[i])
			if err != nil {
				return err
			}
		}

	case *ast.Text:
		r.write(t.Text, true)
		if t.EndOfLine {
			r.indentNext = true
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
		r.write(s, t.Unescaped)

	case *ast.Section:
		v, err := r.lookup(treeName, t.Line, t.Column, t.Key)
		if err != nil {
			return err
		}
		v, err = r.toTruthyValue(v)
		if err != nil {
			return err
		}
		isTruthy := v.IsValid()
		if !t.Inverted && isTruthy {
			switch v.Kind() {
			case reflect.Slice, reflect.Array:
				for i := 0; i < v.Len(); i++ {
					r.push(v.Index(i))
					for j := range t.Nodes {
						err := r.walk(treeName, t.Nodes[j])
						if err != nil {
							return err
						}
					}
					r.pop()
				}
			case reflect.Func:
				s := v.Call([]reflect.Value{reflect.ValueOf(t.Text)})[0].String()
				tree, err := parse.Parse("lambda", s, t.LDelim, t.RDelim)
				if err != nil {
					return nil
				}
				err = r.walk(treeName, tree)
				if err != nil {
					return err
				}

			default:
				r.push(v)
				for i := range t.Nodes {
					err := r.walk(treeName, t.Nodes[i])
					if err != nil {
						return err
					}
				}
				r.pop()
			}
		} else if t.Inverted && !isTruthy {
			for i := range t.Nodes {
				err := r.walk(treeName, t.Nodes[i])
				if err != nil {
					return err
				}
			}
		}

	case *ast.Partial:
		tree, ok := r.template.treeMap[t.Key]
		if !ok {
			if r.template.ContextErrorsEnabled {
				return fmt.Errorf("%s:%d:%d: partial not found: %s", treeName, t.Line, t.Column, t.Key)
			}
			return nil
		}

		origIndent := r.indent
		r.indent += t.Indent

		r.indentNext = true

		r.depth++
		if r.depth >= maxPartialDepth {
			return fmt.Errorf("exceeded maximum partial depth: %d", maxPartialDepth)
		}

		err := r.walk(tree.Name, tree)
		if err != nil {
			return err
		}

		r.depth--

		r.indent = origIndent
	}
	return nil
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
		s, err := r.renderToString(tree)
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

// toTruthyValue returns a value when it is "truthy". If the value is
// falsey, the reflect zero value is returned.
func (r *renderer) toTruthyValue(v reflect.Value) (reflect.Value, error) {
	switch v.Kind() {
	case reflect.Bool:
		if !v.Bool() {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() == 0 {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if v.Uint() == 0 {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.Float32, reflect.Float64:
		if math.Float64bits(v.Float()) == 0 {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		if math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0 {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.String:
		if v.Len() == 0 {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.Array, reflect.Slice:
		if v.IsNil() || v.Len() == 0 {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.Func:
		if v.IsNil() {
			return reflect.Value{}, nil
		}
		t := v.Type()
		isArity0 := t.NumIn() == 0 && t.NumOut() == 1
		if isArity0 {
			v = v.Call(nil)[0]
			if v.Kind() != reflect.String {
				return r.toTruthyValue(v)
			}
			tree, err := parse.Parse("lambda", v.String(), parse.DefaultLeftDelim, parse.DefaultRightDelim)
			if err != nil {
				return reflect.Value{}, nil
			}
			s, err := r.renderToString(tree)
			if err != nil {
				return reflect.Value{}, nil
			}
			return r.toTruthyValue(reflect.ValueOf(s))
		}
		isArity1 := t.NumIn() == 1 && t.In(0).Kind() == reflect.String && t.NumOut() == 1 && t.Out(0).Kind() == reflect.String
		if isArity1 {
			return v, nil
		}
		return reflect.Value{}, nil
	case reflect.Ptr, reflect.Interface:
		return r.toTruthyValue(indirect(v))
	case reflect.Map:
		if v.IsNil() {
			return reflect.Value{}, nil
		}
		return v, nil
	case reflect.Struct:
		return v, nil
	case reflect.Invalid:
		return reflect.Value{}, nil
	default:
		return reflect.Value{}, nil
	}
}

// indirect returns the value that v points to, or concrete
// element underlying an interface.
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

// lookup a key in the context stack. If a value was not found, the reflect.Value zero
// type is returned.
func (r *renderer) lookup(name string, ln, col int, key []string) (reflect.Value, error) {
	v := lookupKeysStack(key, r.stack)
	if !v.IsValid() && r.template.ContextErrorsEnabled {
		return v, fmt.Errorf("%s:%d:%d: cannot find value %s in context", name, ln, col, strings.Join(key, "."))
	}
	return v, nil
}

// lookupKeysStack obtains a value for a dotted key - eg: a.b.c . If a value
// was not found, the reflect.Value zero type is returned.
func lookupKeysStack(key []string, contexts []reflect.Value) reflect.Value {
	var v reflect.Value

	if len(key) == 0 {
		return v
	}

	for i := range key {
		if i == 0 {
			v = lookupKeyStack(key[i], contexts)
			continue
		}
		v = lookupKeyContext(key[i], v)
		if !v.IsValid() {
			break
		}
	}
	return v
}

// lookupKeyStack returns a value from the first context in the stack that
// contains a value for that key. If a value was not found, the reflect.Value zero
// type is returned.
func lookupKeyStack(key string, contexts []reflect.Value) reflect.Value {
	var v reflect.Value
	for i := len(contexts) - 1; i >= 0; i-- {
		ctx := contexts[i]
		v = lookupKeyContext(key, ctx)
		if v.IsValid() {
			break
		}
	}
	return v
}

// lookup returns a value by key from the context. If a value
// was not found, the reflect.Value zero type is returned.
func lookupKeyContext(key string, ctx reflect.Value) reflect.Value {
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
		return lookupKeyContext(key, indirect(ctx))
	case reflect.Map:
		return ctx.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		return ctx.FieldByName(key)
	default:
		return reflect.Value{}
	}
}
