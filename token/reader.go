package token

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

const (
	DefaultLeftDelim  = "{{"
	DefaultRightDelim = "}}"
)

// state represents the possible lexing states
type state int

const (
	stateText state = iota
	stateTag
	stateCommentTag
	stateSetDelimTag
)

const eof rune = -1

type Reader interface {
	// Next returns the next
	Next() (Token, error)
	// The following methods all refer to the most recent token returned by Next.
	// Text returns the original string representation of the
	Text() string
	// Pos reports the position or the most recent token.
	Pos() Position
}

type reader struct {
	src        *bufio.Reader // the input src
	text       bytes.Buffer  // contains the current token literal
	file       string        // name of the file being lexed, if any
	line       int           // line number
	col        int           // column number in bytes
	lastCol    int           // width of the last column in bytes
	width      int           // width of the last read rune in bytes
	lDelim     []byte        // left delim
	rDelim     []byte        // right delim
	nextLDelim []byte        // next left delim
	nextRDelim []byte        // next right delim
	state      state         // current lexing state
}

func NewReader(filename string, src io.Reader, lDelim, rDelim string) Reader {
	return &reader{
		src:        bufio.NewReader(src),
		file:       filename,
		line:       1,
		lDelim:     []byte(lDelim),
		rDelim:     []byte(rDelim),
		nextLDelim: []byte(lDelim),
		nextRDelim: []byte(rDelim),
		state:      stateText,
	}
}

// Text returns the literal string of the most recent returned token
func (r *reader) Text() string {
	return r.text.String()
}

// Pos reports the position or the most recent token.
func (r *reader) Pos() Position {
	return Position{
		File: r.file,
		Line: r.line,
		Col:  r.col,
	}
}

// read returns the next rune in the src stream
func (r *reader) read() rune {
	ch, w, err := r.src.ReadRune()
	if err == io.EOF {
		return eof
	}
	if err != nil {
		panic(err)
	}

	r.advancePosition(w, (ch == '\n'))

	// add byte to current selection
	r.text.WriteRune(ch)

	// store the width of last rune
	r.width = w

	return ch
}

// unread backs-up one rune in the input stream
func (r *reader) unread() {
	err := r.src.UnreadRune()
	if err != nil {
		panic(err)
	}
	r.text.Truncate(r.text.Len() - r.width)
	r.retreatPosition(r.width)
	r.width = 0
	return
}

// advancePosition moves moves the column and line number forward a
// specified number of bytes
func (r *reader) advancePosition(n int, eol bool) {
	r.col += n
	if eol {
		r.lastCol = r.col
		r.col = 0
		r.line++
	}
}

// retreatPosition moves moves the column and line number backwards a
// specified number of bytes
func (r *reader) retreatPosition(n int) {
	if r.col == 0 {
		r.col = r.lastCol
		r.lastCol = 0
		r.line--
	}
	r.col -= n
}

// hasNextBytes returns true if the src stream contains the next
// following bytes.
func (r *reader) hasNextBytes(expect []byte) bool {
	got, err := r.src.Peek(len(expect))
	if err == io.EOF {
		return false
	}
	if err != nil {
		panic(err)
	}
	return bytes.Equal(got, expect)
}

// atLeftDelim returns true when the next src bytes are a left delim
func (r *reader) atLeftDelim() bool {
	return r.hasNextBytes(r.lDelim)
}

// consume left delim will attempt to read the left delim bytes if they
// are the next bytes in the src
func (r *reader) consumeLeftDelim() bool {
	if !r.atLeftDelim() {
		return false
	}
	dLen := len(r.lDelim)
	if _, err := r.src.Discard(dLen); err != nil {
		panic(err)
	}
	r.advancePosition(dLen, false)
	r.text.Write(r.lDelim)
	return true
}

// atLeftDelim returns true when the next src bytes are a right delim
func (r *reader) atRightDelim() bool {
	unescapedRightDelim := append([]byte("}"), r.rDelim...)
	atURD := r.hasNextBytes(unescapedRightDelim)
	atRD := r.hasNextBytes(r.rDelim)
	return !atURD && atRD
}

// consume right delim will attempt to read the right delim bytes if they
// are the next bytes in the src
func (r *reader) consumeRightDelim() bool {
	if !r.atRightDelim() {
		return false
	}
	dLen := len(r.rDelim)
	if _, err := r.src.Discard(dLen); err != nil {
		panic(err)
	}
	r.advancePosition(dLen, false)
	r.text.Write(r.rDelim)
	return true
}

// Next returns the next token in the src file. If an error is returned, the
// token should be disregarded.
func (r *reader) Next() (tok Token, err error) {
	// catches intentionally paniced io.reader errors
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	// reset the literal text buffer
	r.text.Reset()

	// obtain the next token
	switch r.state {
	case stateText:
		tok, r.state = r.lexText()
	case stateTag:
		tok, r.state = r.lexTag()
	case stateCommentTag:
		tok, r.state = r.lexCommentTag()
	case stateSetDelimTag:
		tok, r.state = r.lexSetDelimsTag()
	}
	return
}

// lexText lexes text content that exists between mustache
// tags - e.g. ...}} <---> {{...
func (r *reader) lexText() (Token, state) {
	if r.consumeLeftDelim() {
		return LDELIM, stateTag
	}

	switch ch := r.read(); {
	case ch == eof:
		return EOF, stateText
	case isEndOfLine(ch):
		if ch == '\r' {
			ch2 := r.read()
			if ch2 != '\n' {
				r.unread()
			}
		}
		return EOL, stateText
	case isSpace(ch):
		return r.lexTextContent(PAD)
	default:
		return r.lexTextContent(TEXT)
	}
}

// lextTextContent is a helper function that reads a long continuous
// string of text
func (r *reader) lexTextContent(t Token) (Token, state) {
	s := stateText
	for {
		if r.atLeftDelim() {
			return t, s
		}
		switch ch := r.read(); {
		case ch == eof:
			return t, s
		case isEndOfLine(ch):
			r.unread()
			return t, s
		case !isSpace(ch):
			t = TEXT
		}
	}
}

// lexTag lexss the content inside mustache tags {{<---->}}
func (r *reader) lexTag() (Token, state) {
	if r.consumeRightDelim() {
		return RDELIM, stateText
	}

	switch ch := r.read(); {
	case ch == eof:
		return EOF, stateText
	case isKey(ch):
		return r.lexKey()
	case isSpace(ch):
		return r.lexSpaces()
	case ch == '!':
		return COMMENT, stateCommentTag
	case ch == '=':
		return SETDELIM, stateSetDelimTag
	case ch == '.':
		return DOT, stateTag
	case ch == '^':
		return ISECTION, stateTag
	case ch == '>':
		return PARTIAL, stateTag
	case ch == '#':
		return SECTION, stateTag
	case ch == '/':
		return SECTIONEND, stateTag
	case ch == '&':
		return UNESC, stateTag
	case ch == '{':
		return LUNESC, stateTag
	case ch == '}':
		return RUNESC, stateTag
	default:
		return ILLEGAL, stateTag
	}
}

// lexKey lexes a tag identifier, the 'myvar' in {{ myvar }}
func (r *reader) lexKey() (Token, state) {
	for {
		if ch := r.read(); !isKey(ch) {
			r.unread()
			return KEY, stateTag
		}
	}
}

// lexSpaces lexes continuous spaces inside a tag
func (r *reader) lexSpaces() (Token, state) {
	for {
		if ch := r.read(); !isSpace(ch) {
			r.unread()
			return WS, stateTag
		}
	}
}

// lexCommentTag lexes the content with a mustache comment tag
func (r *reader) lexCommentTag() (Token, state) {
	if r.consumeRightDelim() {
		return RDELIM, stateText
	}
	switch ch := r.read(); {
	case ch == eof:
		return EOF, stateText
	default:
		return r.lexCommentContent()
	}
}

// lexCommentConent is a helper that lexes a comment tag's text content
func (r *reader) lexCommentContent() (Token, state) {
	for {
		if r.atRightDelim() {
			return TEXT, stateCommentTag
		}
		ch := r.read()
		if ch == eof {
			return TEXT, stateCommentTag
		}
	}
}

// lexSetDelimsTag lexes the insides of a set delims tag
func (r *reader) lexSetDelimsTag() (Token, state) {
	if r.consumeRightDelim() {
		r.lDelim = r.nextLDelim
		r.rDelim = r.nextRDelim
		return RDELIM, stateText
	}
	switch ch := r.read(); {
	case ch == eof:
		return EOF, stateText
	case ch == '=':
		return SETDELIM, stateSetDelimTag
	default:
		return r.lexDelimsDescription()
	}
}

// lexDelimsDescription lexes the new delimeter identifier, as well as parsing
// and assigning the new delimeters.
func (r *reader) lexDelimsDescription() (Token, state) {
Loop:
	for {
		if r.atRightDelim() {
			break
		}
		switch ch := r.read(); {
		case ch == eof:
			break Loop
		case ch == '=':
			r.unread()
			break Loop
		}
	}

	rawDelims := make([]byte, r.text.Len())
	copy(rawDelims, r.text.Bytes())
	delims := bytes.SplitN(bytes.TrimSpace(rawDelims), []byte(" "), 2)
	if len(delims) != 2 {
		return ILLEGAL, stateSetDelimTag
	}
	r.nextLDelim = bytes.TrimSpace(delims[0])
	r.nextRDelim = bytes.TrimSpace(delims[1])
	return TEXT, stateSetDelimTag
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(ch rune) bool {
	return ch == '\r' || ch == '\n'
}

// isKey reports whether r is an alphabetic, digit, or underscore.
func isKey(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

// isSpace reports whether r is a space character.
func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}
