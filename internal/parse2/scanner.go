package parse2

// scanner lexes the input source and returns a tokenized representation.
type scanner interface {
	next() (position, token, string, error)
}

type tokenScanner struct {
	reader     reader
	leftDelim  string
	rightDelim string
}

func (t *tokenScanner) next() (position, token, string, error) {
	t.reader.clearText()
	if t.reader.acceptString(t.leftDelim) {

	}
}
