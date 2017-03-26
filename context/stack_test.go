package context

import (
	"reflect"
	"testing"
)

type LookupStruct struct {
	Name  string
	age   int
	Child *LookupStruct
}

func (l LookupStruct) GetName() string {
	return l.Name
}

func (l *LookupStruct) PtrName() string {
	return l.Name
}

type LookupInterface interface {
	PtrName() string
}

func TestStack_Lookup_mapHit(t *testing.T) {
	stack := NewStack(map[string]string{"name": "john doe"})
	v, err := stack.Lookup("name")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_mapInterfaceHit(t *testing.T) {
	stack := NewStack(interface{}(map[string]string{"name": "john doe"}))
	v, err := stack.Lookup("name")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_mapNestedHit(t *testing.T) {
	person := map[string]string{"name": "john doe"}
	people := map[string]map[string]string{"person1": person}
	stack := NewStack(people)
	v, err := stack.Lookup("person1.name")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_mapNestedInterfaceHit(t *testing.T) {
	fullname := interface{}(map[string]interface{}{"first": interface{}("john doe")})
	person := interface{}(map[string]interface{}{"name": interface{}(fullname)})
	people := interface{}(map[string]interface{}{"person1": person})
	stack := NewStack(people)
	v, err := stack.Lookup("person1.name.first")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_mapMiss(t *testing.T) {
	stack := NewStack(map[string]string{"name": "john doe"})
	_, err := stack.Lookup("age")
	if err == nil {
		t.Error("expected lookup miss")
	}
}

func TestStack_Lookup_mapNestedMiss(t *testing.T) {
	person := map[string]string{"name": "john doe"}
	people := map[string]map[string]string{"person1": person}
	stack := NewStack(people)
	_, err := stack.Lookup("person1.blahlbah")
	if err == nil {
		t.Error("expected context miss")
	}
}

func TestStack_Lookup_structFieldHit(t *testing.T) {
	stack := NewStack(LookupStruct{Name: "john doe"})
	v, err := stack.Lookup("Name")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_ptrStructFieldHit(t *testing.T) {
	stack := NewStack(&LookupStruct{Name: "john doe"})
	v, err := stack.Lookup("Name")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_structMethodHit(t *testing.T) {
	stack := NewStack(LookupStruct{Name: "john doe"})
	v, err := stack.Lookup("GetName")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_structPtrMethodHit(t *testing.T) {
	stack := NewStack(&LookupStruct{Name: "john doe"})
	v, err := stack.Lookup("PtrName")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_interfacePtrMethodHit(t *testing.T) {
	var s LookupInterface
	s = &LookupStruct{Name: "john doe"}
	stack := NewStack(s)
	v, err := stack.Lookup("PtrName")
	if err != nil {
		t.Error(err)
	}
	got := v.String()
	expected := "john doe"
	if got != expected {
		t.Errorf("got %s, expected %s", got, expected)
	}
}

func TestStack_Lookup_interfacePtrMethodMiss(t *testing.T) {
	var s LookupInterface
	stack := NewStack(s)
	_, err := stack.Lookup("PtrName")
	if err == nil {
		t.Error("expected miss")
	}
}

func TestStack_Lookup_returnsNonInterface(t *testing.T) {
	stack := NewStack(map[string]interface{}{"name": interface{}(false)})
	v, err := stack.Lookup("name")
	if err != nil {
		t.Error(err)
	}

	if v.value.Kind() == reflect.Interface {
		t.Error("expected non interface internal value")
	}
}
