// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/eriklott/mustache"
)

func TestRender_Spec(t *testing.T) {
	specs := []string{
		"comments.json",
		"delimiters.json",
		"interpolation.json",
		"inverted.json",
		"partials.json",
		"sections.json",
	}

	type test struct {
		Name        string            `json:"name"`
		Data        interface{}       `json:"data"`
		Expected    string            `json:"expected"`
		Template    string            `json:"template"`
		Description string            `json:"desc"`
		Partials    map[string]string `json:"partials"`
	}

	type testSuite struct {
		Tests []test `json:"tests"`
	}

	for _, spec := range specs {
		t.Run(spec, func(t *testing.T) {
			path := filepath.Join("spec", "specs", spec)
			b, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read spec file %s: %v", path, err)
			}
			suite := testSuite{}
			err = json.Unmarshal(b, &suite)
			if err != nil {
				t.Fatalf("failed to unmarshal json spec %s, %v", path, err)
			}

			for _, tc := range suite.Tests {
				t.Run(tc.Name, func(t *testing.T) {

					tmpl := mustache.NewTemplate()

					for k, v := range tc.Partials {
						err := tmpl.Parse(k, v)
						if err != nil {
							t.Fatalf("failed to parse partial %s: %v", k, err)
						}
					}

					err = tmpl.Parse("main", tc.Template)
					if err != nil {
						t.Fatalf("failed to parse template: %v", err)
					}

					got, err := tmpl.Render("main", tc.Data)
					if err != nil {
						t.Fatalf("failed to render template: %v", err)
					}

					// specs test against escaped char &quot; rather than go's &#34;
					got = strings.Replace(got, "&#34;", "&quot;", -1)

					if !reflect.DeepEqual(tc.Expected, got) {
						t.Errorf("unexpected response, got:%s, want:%s", got, tc.Expected)
					}
				})
			}
		})
	}
}

func TestRender_SpecLambda(t *testing.T) {

	var testLambdaNum = 0

	tt := []struct {
		name        string
		description string
		data        interface{}
		template    string
		expected    string
	}{
		{
			"Interpolation",
			"A lambda's return value should be interpolated.",
			map[string]interface{}{
				"lambda": func() string { return "world" },
			},
			"Hello, {{lambda}}!",
			"Hello, world!",
		},
		{
			"Interpolation - Expansion",
			"A lambda's return value should be parsed.",
			map[string]interface{}{
				"planet": "world",
				"lambda": func() string { return "{{planet}}" },
			},
			"Hello, {{lambda}}!",
			"Hello, world!",
		},
		{
			"Interpolation - Alternate Delimiters",
			"A lambda's return value should parse with the default delimiters.",
			map[string]interface{}{
				"planet": "world",
				"lambda": func() string { return "|planet| => {{planet}}" },
			},
			"{{= | | =}}\nHello, (|&lambda|)!",
			"Hello, (|planet| => world)!",
		},
		{
			"Interpolation - Multiple Calls",
			"Interpolated lambdas should not be cached.",
			map[string]interface{}{
				"lambda": func() string { testLambdaNum++; return fmt.Sprintf("%d", testLambdaNum) },
			},
			"{{lambda}} == {{{lambda}}} == {{lambda}}",
			"1 == 2 == 3",
		},
		{
			"Escaping",
			"Lambda results should be appropriately escaped.",
			map[string]interface{}{
				"lambda": func() string { return ">" },
			},
			"<{{lambda}}{{{lambda}}}",
			"<&gt;>",
		},
		{
			"Section",
			"Lambdas used for sections should receive the raw section string.",
			map[string]interface{}{
				"x": "Error!",
				"lambda": func(text string) string {
					if text == "{{x}}" {
						return "yes"
					} else {
						return "no"
					}
				},
			},
			"<{{#lambda}}{{x}}{{/lambda}}>",
			"<yes>",
		},
		{
			"Section - Expansion",
			"Lambdas used for sections should have their results parsed.",
			map[string]interface{}{
				"planet": "Earth",
				"lambda": func(text string) string { return fmt.Sprintf("%s{{planet}}%s", text, text) },
			},
			"<{{#lambda}}-{{/lambda}}>",
			"<-Earth->",
		},
		{
			"Section - Alternate Delimiters",
			"Lambdas used for sections should parse with the current delimiters.",
			map[string]interface{}{
				"planet": "Earth",
				"lambda": func(text string) string { return fmt.Sprintf("%s{{planet}} => |planet|%s", text, text) },
			},
			"{{= | | =}}<|#lambda|-|/lambda|>",
			"<-{{planet}} => Earth->",
		},
		{
			"Section - Multiple Calls",
			"Lambdas used for sections should not be cached.",
			map[string]interface{}{
				"planet": "Earth",
				"lambda": func(text string) string { return "__" + text + "__" },
			},
			"{{#lambda}}FILE{{/lambda}} != {{#lambda}}LINE{{/lambda}}",
			"__FILE__ != __LINE__",
		},
		{
			"Inverted Section",
			"Lambdas used for inverted sections should be considered truthy.",
			map[string]interface{}{
				"static": "static",
				"lambda": func(text string) string { return "" },
			},
			"<{{^lambda}}{{static}}{{/lambda}}>",
			"<>",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := mustache.NewTemplate()
			err := tmpl.Parse("main", tc.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			got, err := tmpl.Render("main", tc.data)
			if err != nil {
				t.Fatalf("failed to render template: %v", err)
			}

			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("unexpected response, got:%s, want:%s", got, tc.expected)
			}
		})
	}
}

type testStruct struct {
}

func (s testStruct) Arity0() string {
	return "Hello World!"
}

func (s testStruct) Arity1(text string) string {
	return "--" + text + "--"
}

func TestRender_Misc(t *testing.T) {
	tt := []struct {
		name     string
		desc     string
		text     string
		data     interface{}
		partials map[string]string
		want     string
		err      string
	}{
		{
			name: "Method",
			desc: "Struct methods are treated as lambdas",
			text: "{{Arity0}},{{#Arity1}}Hello Again!{{/Arity1}}",
			data: testStruct{},
			want: "Hello World!,--Hello Again!--",
		},
		{
			name: "Section - Lambda",
			desc: "An arity 0 function can be used as section data",
			text: "{{#A}}Hello World!{{/A}}",
			data: map[string]interface{}{
				"A": func() bool { return true },
			},
			want: "Hello World!",
		},
		{
			name:     "Recursive Partial",
			desc:     "Infinitely recursive partials will return an error",
			text:     "{{>partial}}",
			partials: map[string]string{"partial": "{{>partial}}"},
			data:     nil,
			err:      "exceeded maximum partial depth: 100000",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := mustache.NewTemplate()
			err := tmpl.Parse("main", tc.text)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			for key, partial := range tc.partials {
				err := tmpl.Parse(key, partial)
				if err != nil {
					t.Fatalf("failed to parse partial: %v", err)
				}
			}

			got, err := tmpl.Render("main", tc.data)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tc.err {
				t.Errorf("unexpected error, got:%s, want:%s", errStr, tc.err)
			}
			if err != nil || tc.err != "" {
				return
			}

			if got != tc.want {
				t.Errorf("unexpected response, got:%s, want:%s", got, tc.want)
			}
		})
	}
}

func TestRender_ContextMiss(t *testing.T) {
	tt := []struct {
		name       string
		text       string
		data       map[string]interface{}
		partials   map[string]string
		errEnabled bool
		want       string
		err        string
	}{
		{
			name: "context miss when errors disabled",
			text: "{{a}} {{b}}",
			data: map[string]interface{}{
				"a": "Hello World!",
			},
			errEnabled: false,
			want:       "Hello World! ",
		},
		{
			name: "context miss when errors enabled",
			text: "{{a}} {{b}}",
			data: map[string]interface{}{
				"a": "Hello World!",
			},
			errEnabled: true,
			err:        "main:1:7: cannot find value b in context",
		},
		{
			name: "partial miss when errors disabled",
			text: "{{>a}} {{>b}}",
			data: map[string]interface{}{},
			partials: map[string]string{
				"a": "A",
			},
			errEnabled: false,
			want:       "A ",
		},
		{
			name: "partial miss when errors enabled",
			text: "{{>a}} {{>b}}",
			data: map[string]interface{}{},
			partials: map[string]string{
				"a": "A",
			},
			errEnabled: true,
			err:        "main:1:8: partial not found: b",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := mustache.NewTemplate()
			tmpl.ContextErrorsEnabled = tc.errEnabled
			err := tmpl.Parse("main", tc.text)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			for key, partial := range tc.partials {
				err := tmpl.Parse(key, partial)
				if err != nil {
					t.Fatalf("failed to parse partial: %v", err)
				}
			}

			got, err := tmpl.Render("main", tc.data)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tc.err {
				t.Errorf("unexpected error, got:%s, want:%s", errStr, tc.err)
			}
			if err != nil || tc.err != "" {
				return
			}

			if got != tc.want {
				t.Errorf("unexpected response, got:%s, want:%s", got, tc.want)
			}
		})
	}
}

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
		t := mustache.NewTemplate()
		err := t.Parse("main", tmpl)
		if err != nil {
			b.Fatalf("failed to parse template: %v", err)
		}
		_, err = t.Render("main", data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
