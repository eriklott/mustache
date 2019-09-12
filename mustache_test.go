package mustache_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/eriklott/mustache"
	m2 "github.com/cbroglie/mustache"
)

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"testing"

// 	m2 "github.com/cbroglie/mustache"
// 	"github.com/eriklott/mustache"
// 	"github.com/google/go-cmp/cmp"
// )

// func TestRender(t *testing.T) {
// 	tt := []struct {
// 		name     string
// 		desc     string
// 		data     interface{}
// 		template string
// 		partials map[string]string
// 		expected string
// 	}{
// 		{
// 			name : "Lambda variable",
// 			desc : "Allows an arity 1 lambda to be used as a variable",
// 			data :

// 		}

// 	}
// 	for _, tc := range tt {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tmpl := mustache.NewTemplate()
// 			err := tmpl.Parse("main", tc.template)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			for k, v := range tc.partials {
// 				err := tmpl.Parse(k, v)
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 			}
// 			got, err := tmpl.Render("main", tc.data)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			if diff := cmp.Diff(tc.expected, got); diff != "" {
// 				t.Errorf("template mismatch (-want +got):\n%s", diff)
// 			}
// 		})
// 	}
// }

func BenchmarkRender(b *testing.B) {
	tmplBytes, err := ioutil.ReadFile("testdata/template.mustache")
	if err != nil {
		b.Fatal(err)
	}
	tmpl := string(tmplBytes)

	dataBytes, err := ioutil.ReadFile("testdata/data.json")
	if err != nil {
		b.Fatal(err)
	}

	var data interface{}
	json.Unmarshal(dataBytes, data)

	for n := 0; n < b.N; n++ {
		_, err := mustache.Render(tmpl, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMustacheRender(b *testing.B) {
	tmplBytes, err := ioutil.ReadFile("testdata/template.mustache")
	if err != nil {
		b.Fatal(err)
	}
	tmpl := string(tmplBytes)

	dataBytes, err := ioutil.ReadFile("testdata/data.json")
	if err != nil {
		b.Fatal(err)
	}

	var data interface{}
	json.Unmarshal(dataBytes, data)

	for n := 0; n < b.N; n++ {
		_, err := m2.Render(tmpl, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
