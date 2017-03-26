package writer

import (
	"fmt"
	"html"
	"io"
)

type Writer interface {
	Write(interface{})
	WriteHTMLEscaped(interface{})
	Indent() int
	IncreaseIndent(int)
	DecreaseIndent(int)
	IndentNext()
}

type writer struct {
	w           io.Writer
	indent      bool
	indentBytes []byte
}

func New(w io.Writer) *writer {
	return &writer{
		w:           w,
		indent:      false,
		indentBytes: []byte(""),
	}
}

func (w *writer) Write(v interface{}) {
	switch t := v.(type) {
	case []byte:
		w.writeBytes(t)
	default:
		w.writeBytes([]byte(w.toString(v)))
	}
}

func (w *writer) WriteHTMLEscaped(v interface{}) {
	w.writeBytes([]byte(html.EscapeString(w.toString(v))))
}

func (w *writer) toString(v interface{}) string {
	switch t := v.(type) {
	case []byte:
		return string(t)
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (w *writer) writeBytes(b []byte) {
	if w.indent {
		w.indent = false
		w.w.Write(w.indentBytes)
	}
	w.w.Write(b)
}

func (w *writer) IndentNext() {
	w.indent = true
}

func (w *writer) Indent() int {
	return len(w.indentBytes)
}

func (w *writer) IncreaseIndent(n int) {
	w.makeIndent(len(w.indentBytes) + n)
}

func (w *writer) DecreaseIndent(n int) {
	w.makeIndent(len(w.indentBytes) - n)
}

func (w *writer) makeIndent(n int) {
	if n < 0 {
		n = 0
	}
	w.indentBytes = []byte(fmt.Sprintf(fmt.Sprintf("%%%ds", n), ""))
}
