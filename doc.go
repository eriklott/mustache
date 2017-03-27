// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package mustache is a full implementation of the mustache template language. For
more information on the mustache spec, refer to the mustache manual at
https://mustache.github.io/mustache.5.html

Mustache Spec Compliance
This mustache implementation complies with 100% of the mustache specs, including
lambda specs.

Documentation



Templates And Partials

Templates and partials are both considered 'partials' by this library. Both
must be named, and are parsed and rendered in the same fashion.

Basic Template
        tmpl := mustache.NewTemplate()
        tmpl.ParseString("main", `Hello {{word}}`)
        tmpl.Render(os.Stdout, "main", map[string]string{"word": "World!"})
        // prints: "Hello World!"

Template with Partial
        tmpl := mustache.NewTemplate()
        tmpl.ParseString("main", `Hello {{>mypartial}}`)
        tmpl.ParseString("mypartial", `World!`)
        tmpl.Render(os.Stdout, "main", nil)
        // prints: "Hello World!"

Lambdas

A lambda can be used as the context data for variable or section tags.

When used as the data for a variable tag, the lambda must be an arity 0 function
that returns a string.
        lambda := func() string {
                return "Hello World!"
        }

        tmpl := mustache.NewTemplate()
        tmpl.ParseString("main", `{{lambda}}`)
        tmpl.Render(os.Stdout, "main", map[string]interface{}{"lambda": lambda})
        // prints: "Hello World!"


When used as the data for a section tag, the lambda must be an arity 1 function
that accepts a string argument, and returns a string.
        lambda := func(txt string) string {
                return "Hello " + txt
        }

        tmpl := mustache.NewTemplate()
        tmpl.ParseString("main", `{{#lambda}}World!{{/lambda}}`)
        tmpl.Render(os.Stdout, "main", map[string]interface{}{"lambda": lambda})
        // prints: "Hello World!"

A function provided as data in any other form will be ignored and considered
a 'falsey' value.

*/
package mustache
