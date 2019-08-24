// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var enabledTests = map[string]map[string]bool{
	"comments.json": map[string]bool{
		"Inline":                           true,
		"Multiline":                        false,
		"Standalone":                       false,
		"Indented Standalone":              false,
		"Standalone Line Endings":          false,
		"Standalone Without Previous Line": false,
		"Standalone Without Newline":       false,
		"Multiline Standalone":             false,
		"Indented Multiline Standalone":    false,
		"Indented Inline":                  false,
		"Surrounding Whitespace":           false,
	},
	"delimiters.json": map[string]bool{
		"Pair Behavior":                    false,
		"Special Characters":               false,
		"Sections":                         false,
		"Inverted Sections":                false,
		"Partial Inheritence":              false,
		"Post-Partial Behavior":            false,
		"Outlying Whitespace (Inline)":     false,
		"Standalone Tag":                   false,
		"Indented Standalone Tag":          false,
		"Pair with Padding":                false,
		"Surrounding Whitespace":           false,
		"Standalone Line Endings":          false,
		"Standalone Without Previous Line": false,
		"Standalone Without Newline":       false,
	},
	"interpolation.json": map[string]bool{
		"No Interpolation":                             false,
		"Basic Interpolation":                          false,
		"HTML Escaping":                                false,
		"Triple Mustache":                              false,
		"Ampersand":                                    false,
		"Basic Integer Interpolation":                  false,
		"Triple Mustache Integer Interpolation":        false,
		"Ampersand Integer Interpolation":              false,
		"Basic Decimal Interpolation":                  false,
		"Triple Mustache Decimal Interpolation":        false,
		"Ampersand Decimal Interpolation":              false,
		"Basic Context Miss Interpolation":             false,
		"Triple Mustache Context Miss Interpolation":   false,
		"Ampersand Context Miss Interpolation":         false,
		"Dotted Names - Basic Interpolation":           false,
		"Dotted Names - Triple Mustache Interpolation": false,
		"Dotted Names - Ampersand Interpolation":       false,
		"Dotted Names - Arbitrary Depth":               false,
		"Dotted Names - Broken Chains":                 false,
		"Dotted Names - Broken Chain Resolution":       false,
		"Dotted Names - Initial Resolution":            false,
		"Interpolation - Surrounding Whitespace":       false,
		"Triple Mustache - Surrounding Whitespace":     false,
		"Ampersand - Surrounding Whitespace":           false,
		"Interpolation - Standalone":                   false,
		"Triple Mustache - Standalone":                 false,
		"Ampersand - Standalone":                       false,
		"Interpolation With Padding":                   false,
		"Triple Mustache With Padding":                 false,
		"Ampersand With Padding":                       false,
	},
	"inverted.json": map[string]bool{
		"Falsey":                           false,
		"Truthy":                           false,
		"Context":                          false,
		"List":                             false,
		"Empty List":                       false,
		"Doubled":                          false,
		"Nested (Falsey)":                  false,
		"Nested (Truthy)":                  false,
		"Context Misses":                   false,
		"Dotted Names - Truthy":            false,
		"Dotted Names - Falsey":            false,
		"Internal Whitespace":              false,
		"Indented Inline Sections":         false,
		"Standalone Lines":                 false,
		"Standalone Indented Lines":        false,
		"Padding":                          false,
		"Dotted Names - Broken Chains":     false,
		"Surrounding Whitespace":           false,
		"Standalone Line Endings":          false,
		"Standalone Without Previous Line": false,
		"Standalone Without Newline":       false,
	},
	"partials.json": map[string]bool{
		"Basic Behavior":                   false,
		"Failed Lookup":                    false,
		"Context":                          false,
		"Recursion":                        false,
		"Surrounding Whitespace":           false,
		"Inline Indentation":               false,
		"Standalone Line Endings":          false,
		"Standalone Without Previous Line": false,
		"Standalone Without Newline":       false,
		"Standalone Indentation":           false,
		"Padding Whitespace":               false,
	},
	"sections.json": map[string]bool{
		"Truthy":                           false,
		"Falsey":                           false,
		"Context":                          false,
		"Deeply Nested Contexts":           false,
		"List":                             false,
		"Empty List":                       false,
		"Doubled":                          false,
		"Nested (Truthy)":                  false,
		"Nested (Falsey)":                  false,
		"Context Misses":                   false,
		"Implicit Iterator - String":       false,
		"Implicit Iterator - Integer":      false,
		"Implicit Iterator - Decimal":      false,
		"Implicit Iterator - Array":        false,
		"Dotted Names - Truthy":            false,
		"Dotted Names - Falsey":            false,
		"Dotted Names - Broken Chains":     false,
		"Surrounding Whitespace":           false,
		"Internal Whitespace":              false,
		"Indented Inline Sections":         false,
		"Standalone Lines":                 false,
		"Indented Standalone Lines":        false,
		"Standalone Line Endings":          false,
		"Standalone Without Previous Line": false,
		"Standalone Without Newline":       false,
		"Padding":                          false,
	},
	"~lambdas.json": nil, // not implemented
}

type specTest struct {
	Name        string            `json:"name"`
	Data        interface{}       `json:"data"`
	Expected    string            `json:"expected"`
	Template    string            `json:"template"`
	Description string            `json:"desc"`
	Partials    map[string]string `json:"partials"`
}

type specTestSuite struct {
	Tests []specTest `json:"tests"`
}

func TestSpec(t *testing.T) {
	root := filepath.Join(os.Getenv("PWD"), "spec", "specs")
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Could not find the specs folder at %s, ensure the submodule exists by running 'git submodule update --init'", root)
		}
		t.Fatal(err)
	}

	paths, err := filepath.Glob(root + "/*.json")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)

	for _, path := range paths {
		_, file := filepath.Split(path)
		enabled, ok := enabledTests[file]
		if !ok {
			t.Errorf("Unexpected file %s, consider adding to enabledFiles", file)
			continue
		}
		if enabled == nil {
			continue
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var suite specTestSuite
		err = json.Unmarshal(b, &suite)
		if err != nil {
			t.Fatal(err)
		}
		for _, test := range suite.Tests {
			runTest(t, file, &test)
		}
	}
}

func runTest(t *testing.T, file string, test *specTest) {
	enabled, ok := enabledTests[file][test.Name]
	if !ok {
		t.Errorf("[%s %s]: Unexpected test, add to enabledTests", file, test.Name)
	}
	if !enabled {
		t.Logf("[%s %s]: Skipped", file, test.Name)
		return
	}

	// init template
	tmpl := NewTemplate()
	err := tmpl.Parse("main", test.Template)
	if err != nil {
		t.Error(err)
	}
	for k, v := range test.Partials {
		err := tmpl.Parse(k, v)
		if err != nil {
			t.Error(err)
		}
	}

	// render template
	out := tmpl.Render("main", test.Data)

	// specs test against escaped char &quot; rather than go's &#34;
	out = strings.Replace(out, "&#34;", "&quot;", -1)

	if out != test.Expected {
		t.Errorf("[%s %s]: Expected %q, got %q", file, test.Name, test.Expected, out)
		return
	}

	t.Logf("[%s %s]: Passed", file, test.Name)
}
