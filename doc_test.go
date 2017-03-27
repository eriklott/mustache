package mustache

import (
	"os"

	"github.com/eriklott/mustache"
)

func Example_basic() {
	t := mustache.NewTemplate()
	t.ParseString("main", `Hello {{word}}`)
	t.Render(os.Stdout, "main", map[string]string{"word": "World!"})
	// "Hello World!"
}

func Example_partial() {
	t := mustache.NewTemplate()
	t.ParseString("main", `Hello {{>mypartial}}`)
	t.ParseString("mypartial", `World!`)
	tmpl.Render(os.Stdout, "main", nil)
	// "Hello World!"
}

// A lambda rendered by a variable tag must be a arity 0 function that returns
// a string. Any function not conforming to this signature will be considered
// falsey.
func Example_lambda() {
	lambda := func() string {
		return "Hello World!"
	}
	t := mustache.NewTemplate()
	t.ParseString("main", `{{lambda}}`)
	t.Render(os.Stdout, "main", map[string]interface{}{"lambda": lambda})
	// "Hello World!"
}

// A lambda rendered by a section tag must be a arity 1 function that accepts
// a string argument, and returns a string. Any function not conforming to this
// signature will be considered falsey by the section tag.
func Example_lambdaSection() {
	lambda := func(txt string) string {
		return "Hello " + txt
	}
	t := mustache.NewTemplate()
	t.ParseString("main", `{{#lambda}}World!{{/lambda}}`)
	t.Render(os.Stdout, "main", map[string]interface{}{"lambda": lambda})
	// "Hello World!"
}

// Dotted names should be considered a form of shorthand for sections.
func Example_dottedNames() {
	person := map[string]string{"name": "Joe"}
	data := map[string]interface{}{"person": person}

	t := mustache.NewTemplate()
	t.ParseString("main", `{{person.name}}`)
	t.Render(os.Stdout, "main", data)
	// "Joe"
}

// Structs used as template data will have their exported fields and methods
// available in the template. Unexported fields and methods are not available
// to the template.
func Example_dataStruct() {
	data := struct {
		Name string
		age  int
	}{
		Name: "John",
		age:  31,
	}

	t := mustache.NewTemplate()
	t.ParseString("main", `My name is {{Name}}, and I am {{age}} years old.`)
	t.Render(os.Stdout, "main", data)
	// "My name is John, and I am years old"
}
