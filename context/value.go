// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"reflect"
)

type Value struct {
	value reflect.Value
}

func newValue(v reflect.Value) Value {
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(v.Interface())
	}
	return Value{v}
}

func (v Value) String() string {
	var s string
	switch v.value.Kind() {
	case reflect.Func:
		s = ""
	default:
		s = fmt.Sprint(v.value)
	}
	return s
}

// isLambda returns true if the value is a arity 0 function that returns
// 1 value
func (v Value) IsLambda() bool {
	val := v.value
	if !val.IsValid() {
		return false
	}
	if val.Kind() != reflect.Func {
		return false
	}
	typ := val.Type()
	if typ.NumIn() != 0 {
		return false
	}
	if typ.NumOut() != 1 {
		return false
	}
	if typ.Out(0).Kind() != reflect.String {
		return false
	}
	return true
}

func (v Value) CallLambda() string {
	var s string
	if v.IsLambda() {
		results := v.value.Call([]reflect.Value{})
		if len(results) > 0 {
			s = results[0].String()
		}
	}
	return s
}

func (v Value) IsTruthy() bool {
	val := v.value
	truth := false
	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		truth = val.Len() > 0
	case reflect.Bool:
		truth = val.Bool()
	case reflect.Complex64, reflect.Complex128:
		truth = val.Complex() != 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		truth = val.Int() != 0
	case reflect.Float32, reflect.Float64:
		truth = val.Float() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		truth = val.Uint() != 0
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Interface:
		truth = !val.IsNil()
	case reflect.Struct:
		truth = true // Struct values are always true.
	}
	return truth
}

func (v Value) IsList() bool {
	var isList bool
	switch v.value.Kind() {
	case reflect.Array, reflect.Slice:
		isList = true
	default:
		isList = false
	}
	return isList
}

func (v Value) Len() int {
	var len int
	if v.IsList() {
		len = v.value.Len()
	}
	return len
}

func (v Value) Index(i int) (val Value) {
	defer func() {
		if r := recover(); r != nil {
			val = Value{reflect.Value{}}
		}
	}()
	if v.IsList() {
		val.value = v.value.Index(i)
	}
	return val
}

// IsSectionLambda returns true if the value is a arity 1 function that accepts
// a string argument, and returns a string
func (v Value) IsSectionLambda() bool {
	val := v.value
	if !val.IsValid() {
		return false
	}
	if val.Kind() != reflect.Func {
		return false
	}
	typ := val.Type()
	if typ.NumIn() != 1 {
		return false
	}
	if typ.In(0).Kind() != reflect.String {
		return false
	}
	if typ.NumOut() != 1 {
		return false
	}
	if typ.Out(0).Kind() != reflect.String {
		return false
	}
	return true
}

func (v Value) CallSectionLambda(text string) string {
	var s string
	if v.IsSectionLambda() {
		results := v.value.Call([]reflect.Value{reflect.ValueOf(text)})
		if len(results) > 0 {
			s = results[0].String()
		}
	}
	return s
}
