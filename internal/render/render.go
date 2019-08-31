package render

import (
	"fmt"
	"html"
	"reflect"
	"strconv"
	"strings"

	"github.com/eriklott/mustache/internal/ast"
)

const maxPartialDepth = 100000

type renderer struct {
	b          strings.Builder
	indent     string
	indentNext bool
	contexts   []reflect.Value
	treeMap    map[string]*ast.Tree
	depth      int //the height of the stack of executing partials
}

func Render(tree *ast.Tree, treeMap map[string]*ast.Tree, contexts []interface{}) string {
	var reversedContexts []reflect.Value
	for i := len(contexts) - 1; i >= 0; i-- {
		c := reflect.ValueOf(contexts[i])
		reversedContexts = append(reversedContexts, c)
	}

	r := &renderer{
		b:          strings.Builder{},
		indent:     "",
		indentNext: false,
		contexts:   reversedContexts,
		treeMap:    treeMap,
		depth:      0,
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

func (r *renderer) increaseIndent(i string) {
	r.indent += i
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

func (r *renderer) render(node ast.Node) {
	switch t := node.(type) {
	case *ast.Tree:
		for _, child := range t.Nodes {
			r.render(child)
		}

	case *ast.Text:
		r.write(t.Value)
		if t.EOL {
			r.indentNextWrite()
		}

	case *ast.Variable:
		v := r.lookup(t.Key)

		if t.Unescaped {
			r.write(toString(v))
		} else {
			r.write(html.EscapeString(toString(v)))
		}

	case *ast.Section:
		v := r.lookup(t.Key)
		if isTruthy(v) == t.Inverted {
			return
		}

		if t.Inverted {
			for _, child := range t.Nodes {
				r.render(child)
			}
			return
		}

		if isList(v) {
			for i := 0; i < getLen(v); i++ {
				r.addContext(getIndex(v, i))
				for _, child := range t.Nodes {
					r.render(child)
				}
				r.removeContext()
			}
		} else {
			r.addContext(v)
			for _, child := range t.Nodes {
				r.render(child)
			}
			r.removeContext()
		}

	case *ast.Partial:
		tree, ok := r.treeMap[t.Key]
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

func (r *renderer) lookup(key string) reflect.Value {
	var v reflect.Value
	if key == "." {
		v = r.contexts[len(r.contexts)-1]
	} else if strings.Contains(key, ".") {
		keys := strings.Split(key, ".")
		for i := 0; i < len(keys); i++ {
			if i == 0 {
				v = r.lookup(keys[i])
			} else {
				v = lookup(keys[i], v)
			}
			if !v.IsValid() {
				break
			}
		}
	} else {
		for i := len(r.contexts) - 1; i >= 0; i-- {
			context := r.contexts[i]
			v = lookup(key, context)
			if v.IsValid() {
				break
			}
		}
	}
	return v
}

func lookup(key string, ctx reflect.Value) reflect.Value {
	// check for method on type.
	if ctx.IsValid() {
		method := ctx.MethodByName(key)
		if method.IsValid() {
			if method.Type().NumIn() != 0 {
				return reflect.Value{}
			}
			values := method.Call(nil)
			if len(values) != 1 {
				return reflect.Value{}
			}
			return values[0]
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

func toString(v reflect.Value) string {
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
		t := v.Type()
		if t.NumIn() == 0 && t.NumOut() == 1 {
			return toString(v.Call(nil)[0])
		}
		return ""
	case reflect.Ptr, reflect.Interface:
		return toString(indirect(v))
	case reflect.Invalid:
		return ""
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

func isTruthy(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return v.Len() > 0
	case reflect.Bool:
		return v.Bool()
	case reflect.Chan, reflect.Func:
		return !v.IsNil()
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return false
		}
		return isTruthy(indirect(v))
	case reflect.Struct, reflect.Map:
		return true // Struct values are always true.
	default:
		return v.IsValid()
	}
}

func isList(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return true
	case reflect.Ptr, reflect.Interface:
		return isList(indirect(v))
	default:
		return false
	}
}

func getLen(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return v.Len()
	case reflect.Ptr, reflect.Interface:
		return getLen(indirect(v))
	default:
		return 0
	}
}

func getIndex(v reflect.Value, i int) reflect.Value {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return v.Index(i)
	case reflect.Ptr, reflect.Interface:
		return getIndex(indirect(v), i)
	default:
		return reflect.Value{}
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
