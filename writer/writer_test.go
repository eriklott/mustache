package writer

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

type writerTest struct {
	name     string
	expected string
	input    []interface{}
}

func mkWriterTest(name string, expected string, input ...interface{}) writerTest {
	return writerTest{
		name:     name,
		expected: expected,
		input:    input,
	}
}

var writerTests = []writerTest{
	mkWriterTest("string", "hello", "hello"),
	mkWriterTest("bytes", "hello", []byte("hello")),
	mkWriterTest("other structure", "hello", errors.New("hello")),
	mkWriterTest("reflect value", "hello", reflect.ValueOf("hello")),
}

func TestWriter_Write(t *testing.T) {
	for _, test := range writerTests {
		var b bytes.Buffer
		writer := New(&b)
		for _, i := range test.input {
			writer.Write(i)
		}
		got := b.String()
		if got != test.expected {
			t.Errorf("got\n'%s'\nexpected\n'%s'\n", got, test.expected)
		}
	}
}
