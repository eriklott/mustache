package mustache_test

import (
	"testing"

	// "github.com/cbroglie/mustache"
	"github.com/eriklott/mustache"
	"github.com/google/go-cmp/cmp"
)

type testStruct struct {
	Value string
}

func (t testStruct) AsMethod() string     { return t.Value }
func (t *testStruct) AsPtrMethod() string { return t.Value }

type testStringType string

func (s testStringType) Method() string {
	return string(s)
}

func TestRender(t *testing.T) {
	tt := []struct {
		name   string
		tmpl   string
		data   interface{}
		result string
	}{
		{
			name:   "Pointer Struct Data",
			tmpl:   "{{Value}}-{{AsMethod}}-{{AsPtrMethod}}",
			data:   &testStruct{"hello"},
			result: "hello-hello-hello",
		},
		{
			name:   "Struct Data",
			tmpl:   "{{Value}}-{{AsMethod}}-{{AsPtrMethod}}",
			data:   testStruct{"hello"},
			result: "hello-hello-",
		},
		{
			name:   "Custom Type Data",
			tmpl:   "{{.}}-{{Method}}",
			data:   testStringType("hello"),
			result: "hello-hello",
		},
		{
			name:   "Accepts arity 0 func as variable",
			tmpl:   "{{.}}",
			data:   func() string { return "hello world!" },
			result: "hello world!",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s, err := mustache.Render(tc.tmpl, tc.data)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.result, s); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
