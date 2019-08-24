// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mustache

// import (
// 	"bytes"
// 	"fmt"
// 	"testing"
// )

// type lambdaTest struct {
// 	Name        string
// 	Description string
// 	Data        interface{}
// 	Template    string
// 	Expected    string
// }

// var testLambdaNum int = 0

// var lambdaTests = []lambdaTest{
// 	{
// 		"Interpolation",
// 		"A lambda's return value should be interpolated.",
// 		map[string]interface{}{
// 			"lambda": func() string { return "world" },
// 		},
// 		"Hello, {{lambda}}!",
// 		"Hello, world!",
// 	}, {
// 		"Interpolation - Expansion",
// 		"A lambda's return value should be parsed.",
// 		map[string]interface{}{
// 			"planet": "world",
// 			"lambda": func() string { return "{{planet}}" },
// 		},
// 		"Hello, {{lambda}}!",
// 		"Hello, world!",
// 	}, {
// 		"Interpolation - Alternate Delimiters",
// 		"A lambda's return value should parse with the default delimiters.",
// 		map[string]interface{}{
// 			"planet": "world",
// 			"lambda": func() string { return "|planet| => {{planet}}" },
// 		},
// 		"{{= | | =}}\nHello, (|&lambda|)!",
// 		"Hello, (|planet| => world)!",
// 	}, {
// 		"Interpolation - Multiple Calls",
// 		"Interpolated lambdas should not be cached.",
// 		map[string]interface{}{
// 			"lambda": func() string { testLambdaNum++; return fmt.Sprintf("%d", testLambdaNum) },
// 		},
// 		"{{lambda}} == {{{lambda}}} == {{lambda}}",
// 		"1 == 2 == 3",
// 	}, {
// 		"Escaping",
// 		"Lambda results should be appropriately escaped.",
// 		map[string]interface{}{
// 			"lambda": func() string { return ">" },
// 		},
// 		"<{{lambda}}{{{lambda}}}",
// 		"<&gt;>",
// 	}, {
// 		"Section",
// 		"Lambdas used for sections should receive the raw section string.",
// 		map[string]interface{}{
// 			"x": "Error!",
// 			"lambda": func(text string) string {
// 				if text == "{{x}}" {
// 					return "yes"
// 				} else {
// 					return "no"
// 				}
// 			},
// 		},
// 		"<{{#lambda}}{{x}}{{/lambda}}>",
// 		"<yes>",
// 	}, {
// 		"Section - Expansion",
// 		"Lambdas used for sections should have their results parsed.",
// 		map[string]interface{}{
// 			"planet": "Earth",
// 			"lambda": func(text string) string { return fmt.Sprintf("%s{{planet}}%s", text, text) },
// 		},
// 		"<{{#lambda}}-{{/lambda}}>",
// 		"<-Earth->",
// 	}, {
// 		"Section - Alternate Delimiters",
// 		"Lambdas used for sections should parse with the current delimiters.",
// 		map[string]interface{}{
// 			"planet": "Earth",
// 			"lambda": func(text string) string { return fmt.Sprintf("%s{{planet}} => |planet|%s", text, text) },
// 		},
// 		"{{= | | =}}<|#lambda|-|/lambda|>",
// 		"<-{{planet}} => Earth->",
// 	}, {
// 		"Section - Multiple Calls",
// 		"Lambdas used for sections should not be cached.",
// 		map[string]interface{}{
// 			"planet": "Earth",
// 			"lambda": func(text string) string { return "__" + text + "__" },
// 		},
// 		"{{#lambda}}FILE{{/lambda}} != {{#lambda}}LINE{{/lambda}}",
// 		"__FILE__ != __LINE__",
// 	}, {
// 		"Inverted Section",
// 		"Lambdas used for inverted sections should be considered truthy.",
// 		map[string]interface{}{
// 			"static": "static",
// 			"lambda": func(text string) string { return "" },
// 		},
// 		"<{{^lambda}}{{static}}{{/lambda}}>",
// 		"<>",
// 	},
// }

// func TestLambda(t *testing.T) {
// 	for _, test := range lambdaTests {
// 		tmpl := NewTemplate()
// 		err := tmpl.Parse("main", test.Template)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		var b bytes.Buffer
// 		tmpl.Render(&b, "main", test.Data)
// 		out := b.String()
// 		if out != test.Expected {
// 			t.Errorf("[%s]: Expected %q, got %q", test.Name, test.Expected, out)
// 			return
// 		}
// 		t.Logf("[%s]: Passed", test.Name)
// 	}
// }
