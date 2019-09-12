package render

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/eriklott/mustache/internal/parse"
)

const maxPartialDepth = 100000

type renderer struct {
	b          strings.Builder
	indent     string
	indentNext bool
	contexts   []reflect.Value
	trees      map[string]*parse.Tree
	depth      int //the height of the stack of executing partials
}

func Render(tree *parse.Tree, trees map[string]*parse.Tree, contexts []reflect.Value) string {
	r := &renderer{
		indentNext: true,
		contexts:   contexts,
		trees:      trees,
	}
	r.render(tree)
	return r.b.String()
}

func (r *renderer) write(s string) {
	if r.indentNext {
		r.indentNext = false
		r.b.WriteString(r.indent)
	}
	r.b.WriteString(s)
}

func (r *renderer) escapedWrite(s string) {
	r.write(html.EscapeString(s))
}

func (r *renderer) increaseIndent(s string) {
	r.indent += s
}

func (r *renderer) decreaseIndent(i string) {
	r.indent = r.indent[:len(r.indent)-len(i)]
}

func (r *renderer) indentNextWrite() {
	r.indentNext = true
}

func (r *renderer) addContext(context reflect.Value) {
	r.contexts = append(r.contexts, context)
}

func (r *renderer) removeContext() {
	r.contexts = r.contexts[:len(r.contexts)-1]
}

func (r *renderer) render(node interface{}) {
	switch t := node.(type) {
	case *parse.Tree:
		for i := range t.Nodes {
			r.render(t.Nodes[i])
		}

	case *parse.Text:
		r.write(t.Text)
		if t.EndOfLine {
			r.indentNextWrite()
		}

	case *parse.VariableTag:
		v := r.lookup(t.Key)
		s := r.toString(v, parse.DefaultLeftDelim, parse.DefaultRightDelim)
		if t.Unescaped {
			r.write(s)
		} else {
			r.escapedWrite(s)
		}

	case *parse.SectionTag:
		v := indirect(r.lookup(t.Key))
		isTruthy := isSectionTruthy(v)
		if !t.Inverted && isTruthy {
			switch v.Kind() {
			case reflect.Slice, reflect.Array:
				for i := 0; i < v.Len(); i++ {
					r.addContext(v.Index(i))
					for _, node := range t.Nodes {
						r.render(node)
					}
					r.removeContext()
				}
			case reflect.Func:
				s := v.Call([]reflect.Value{reflect.ValueOf(t.ChildrenText)})[0].String()
				tree, err := parse.Parse("lambda", s, t.LDelim, t.RDelim)
				if err != nil {
					return
				}
				r.write(Render(tree, r.trees, r.contexts))

			default:
				r.addContext(v)
				for _, node := range t.Nodes {
					r.render(node)
				}
				r.removeContext()
			}
		} else if t.Inverted && !isTruthy {
			for _, node := range t.Nodes {
				r.render(node)
			}
		}

	case *parse.PartialTag:
		tree, ok := r.trees[t.Key]
		if !ok {
			return
		}
		r.increaseIndent(t.Indent)
		r.indentNextWrite()
		r.depth++
		if r.depth == maxPartialDepth {
			panic(fmt.Sprintf("partial recursion exceeded max depth: %s", t.Key))
		}
		r.render(tree)
		r.depth--
		r.decreaseIndent(t.Indent)
	}
}

func (r *renderer) lookup(key []string) reflect.Value {
	return lookupKeysContexts(key, r.contexts)
}

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

func (r *renderer) toString(v reflect.Value, ldelim, rdelim string) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%v", v.Complex())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Func:
		if v.IsNil() {
			return ""
		}

		t := v.Type()
		isArity0 := t.NumIn() == 0 && t.NumOut() == 1 && t.Out(0).Kind() == reflect.String
		if !isArity0 {
			return ""
		}

		s := v.Call(nil)[0].String()
		tree, err := parse.Parse("lambda", s, ldelim, rdelim)
		if err != nil {
			return ""
		}
		return Render(tree, r.trees, r.contexts)

	case reflect.Ptr, reflect.Interface:
		return r.toString(indirect(v), ldelim, rdelim)
	case reflect.Chan:
		return ""
	case reflect.Invalid:
		return ""
	default:
		return fmt.Sprintf("%v", v.Interface())
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

func isSectionTruthy(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) != 0
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return !(math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0)
	case reflect.String:
		return v.Len() != 0
	case reflect.Array, reflect.Slice:
		return !v.IsNil() && v.Len() > 0
	case reflect.Func:
		if v.IsNil() {
			return false
		}
		t := v.Type()
		isArity1 := t.NumIn() == 1 && t.NumOut() == 1 && t.Out(0).Kind() == reflect.String
		return isArity1
	case reflect.Ptr, reflect.Interface:
		return isSectionTruthy(indirect(v))
	case reflect.Map:
		return !v.IsNil()
	case reflect.Struct:
		return true
	case reflect.Invalid:
		return false
	default:
		return false
	}
}
