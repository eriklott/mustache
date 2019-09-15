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

func TestSpec(t *testing.T) {
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

func TestLambda(t *testing.T) {

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
				t.Fatal(err)
			}

			got, err := tmpl.Render("main", tc.data)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("unexpected response, got:%s, want:%s", got, tc.expected)
			}
		})
	}
}

type structContext struct {
	v  string
	v2 int
}

func (s structContext) StringVal() string {
	return s.v
}

func (s structContext) IntVal() int {
	return s.v2
}

func (s *structContext) PtrStringVal() string {
	return s.v
}

func (s *structContext) PtrIntVal() int {
	return s.v2
}

func TestRender_Methods(t *testing.T) {
	got, err := mustache.Render("{{StringVal}},{{IntVal}}", structContext{"1", 2})
	if err != nil {
		t.Fatal(err)
	}
	want := "1,2"
	if got != want {
		t.Errorf("unexpected response, got: '%s', want: '%s'", got, want)
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
		_, err := mustache.Render(tmpl, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
