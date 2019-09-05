package mustache_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	m2 "github.com/cbroglie/mustache"
	m1 "github.com/eriklott/mustache"
)

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
		_, err := m1.Render(tmpl, data)
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
