// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Stack struct {
	stack []reflect.Value
}

func NewStack(contexts ...interface{}) *Stack {
	stack := []reflect.Value{}
	for _, ctx := range contexts {
		stack = append(stack, reflect.ValueOf(ctx))
	}
	return &Stack{stack}
}

func (s *Stack) Add(v Value) {
	s.stack = append(s.stack, v.value)
}

func (s *Stack) Remove() {
	if len(s.stack) > 0 {
		s.stack = s.stack[:len(s.stack)-1]
	}
	return
}

func (s *Stack) Lookup(ident string) (Value, error) {
	var v reflect.Value
	var err error
	if len(s.stack) == 0 {
		err = errors.New("no context data")
	} else if ident == "." {
		v = s.stack[len(s.stack)-1]
	} else {
		stack := s.stack
		keys := strings.Split(ident, ".")
		for _, key := range keys {
			for i := len(stack) - 1; i >= 0; i-- {
				v, err = lookup(key, stack[i])
				if err == nil {
					break
				}
			}
			if err != nil {
				break
			}
			stack = []reflect.Value{v}
		}
	}
	return newValue(v), err
}

func lookup(key string, ctx reflect.Value) (reflect.Value, error) {
	ctx, _ = indirect(ctx)
	var v reflect.Value
	switch ctx.Kind() {
	case reflect.Map:
		if v = ctx.MapIndex(reflect.ValueOf(key)); v.IsValid() {
			return v, nil
		}
	case reflect.Struct:
		// struct field
		if v = ctx.FieldByName(key); v.IsValid() {
			return v, nil
		}

		// struct method
		// Unless it's an interface, need to get to a value of type *T to guarantee
		// we see all methods of T and *T.
		ptr := ctx
		if ptr.Kind() != reflect.Interface && ptr.Kind() != reflect.Ptr && ptr.CanAddr() {
			ptr = ptr.Addr()
		}
		// find method on *T
		if method := ptr.MethodByName(key); method.IsValid() {
			if method.Type().NumIn() != 0 {
				return v, fmt.Errorf("method %q requires arguments", key)
			}
			values := method.Call([]reflect.Value{})
			if len(values) != 1 {
				return v, fmt.Errorf("method %q returns multiple values", key)
			}
			v = values[0]
			return v, nil
		}
	}
	return v, fmt.Errorf("key %q not found", key)
}

// indirect returns the item at the end of indirection, and a bool to indicate if it's nil.
func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
	}
	return v, false
}
