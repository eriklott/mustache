package context

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValue_String(t *testing.T) {
	// string
	got := Value{reflect.ValueOf("hello")}.String()
	expect := "hello"

	if got != expect {
		t.Errorf("got %s, expected %s", got, expect)
	}

	// array
	got = Value{reflect.ValueOf([]int{3, 4, 5})}.String()
	expect = "[3 4 5]"

	if got != expect {
		t.Errorf("got %s, expected %s", got, expect)
	}

	// arity 0 lambda
	lambda0 := func() string { return "hello world" }
	got = Value{reflect.ValueOf(lambda0)}.String()
	expect = "hello world"
	if got != expect {
		t.Errorf("got %s, expected %s", got, expect)
	}

	// non arity 0 function
	lambda1 := func(v string) string { return "hello world" }
	got = Value{reflect.ValueOf(lambda1)}.String()
	expect = ""
	if got != expect {
		t.Errorf("got %s, expected %s", got, expect)
	}
}

func TestValue_IsTruthy(t *testing.T) {
	// Zero Value (invalid) - false
	value := Value{reflect.Value{}}
	if value.IsTruthy() {
		t.Error("expected zero to be falsey")
	}
}

func TestValue_IsTruthy_false(t *testing.T) {
	value := Value{reflect.ValueOf(false)}
	if value.IsTruthy() {
		t.Error("expected false to be falsey")
	}
}

func TestValue_IsList(t *testing.T) {
	// Slice - true
	value := Value{reflect.ValueOf([]int{1, 2, 3})}
	if !value.IsList() {
		t.Error("expected slice to be list")
	}

	// Array - true
	value = Value{reflect.ValueOf([3]int{1, 2, 3})}
	if !value.IsList() {
		t.Error("expected array to be list")
	}

	// Zero Value (invalid) - false
	value = Value{reflect.Value{}}
	if value.IsList() {
		t.Error("expected zero valie to not be list")
	}
}

func TestValue_Index(t *testing.T) {
	// Slice
	value := Value{reflect.ValueOf([]int{1, 2, 3})}
	if value.Index(0).String() != "1" {
		t.Error("expected 1")
	}

	// Slice - out of range
	value = Value{reflect.ValueOf([]int{1, 2, 3})}
	if value.Index(23).value.IsValid() {
		t.Error("expected invalid value")
	}

	// Non List returns zero value
	value = Value{reflect.ValueOf(1)}
	if value.Index(0).value.IsValid() {
		t.Error("expected invalid value")
	}
}

func TestValue_IsSectionLambda(t *testing.T) {
	fn := func(text string) string {
		return fmt.Sprintf("hello %s", text)
	}
	value := Value{reflect.ValueOf(fn)}
	if !value.IsSectionLambda() {
		t.Error("expected section lambda")
	}
}

func TestValue_CallSectionLambda(t *testing.T) {
	fn := func(text string) string {
		return fmt.Sprintf("hello %s", text)
	}
	value := Value{reflect.ValueOf(fn)}
	got := value.CallSectionLambda("world")
	expected := "hello world"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
