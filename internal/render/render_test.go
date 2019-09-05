package render_test

// func BenchmarkRender(b *testing.B) {
// 	tree, err := parser.Parse("Hello {{subject}}")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	context := map[string]string{"subject": "World!"}
// 	treeMap := map[string]*ast.Tree{}
// 	for n := 0; n < b.N; n++ {
// 		render.Render(tree, treeMap, []interface{}{context})
// 	}
// }

// func BenchmarkMustacheRender(b *testing.B) {

// 	t, err := mustache.ParseString("Hello {{subject}}")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	context := map[string]string{"subject": "World!"}

// 	for n := 0; n < b.N; n++ {
// 		t.Render(context)

// 	}
// }
