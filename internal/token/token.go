// Copyright 2019 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package token

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

// Type of a token
type Type int

// Types of tokens
const (
	TEXT Type = iota
	TEXT_EOL
	VARIABLE
	UNESCAPED_VARIABLE
	UNESCAPED_VARIABLE_SYM
	SECTION
	INVERTED_SECTION
	SECTION_END
	PARTIAL
	COMMENT
	SET_DELIMETERS
)

// Scanner transforms a mustache text template into a stream of tokens.
type Scanner struct {
	name      string
	src       string
	ldelim    string
	rdelim    string
	pos       int
	col       int
	ln        int
	isNewLine bool
	buf       Token
}

// NewScanner returns a new scanner instance
func NewScanner(name, src, ldelim, rdelim string) *Scanner {
	return &Scanner{
		name:      name,
		src:       src,
		ldelim:    ldelim,
		rdelim:    rdelim,
		pos:       0,
		col:       1,
		ln:        1,
		isNewLine: true,
	}
}

func (s *Scanner) LeftDelim() string  { return s.ldelim }
func (s *Scanner) RightDelim() string { return s.rdelim }

func (s *Scanner) readTo(pattern string, haltEOL bool) (bool, error) {
	matchIdx := 0
	for {
		if s.pos >= len(s.src) {
			return false, io.EOF
		}
		b := s.src[s.pos]
		s.pos++
		s.col++
		if b == '\n' {
			s.col = 1
			s.ln++
			if haltEOL {
				return true, nil
			}
		}

		if b != pattern[matchIdx] {
			matchIdx = 0
			continue
		}

		matchIdx++
		if matchIdx == len(pattern) {
			return false, nil
		}
	}
}

type Token struct {
	Type      Type
	Text      string
	Indent    string
	Offset    int
	EndOffset int
	Column    int
	Line      int
}

func (s *Scanner) Next() (Token, error) {
	// return the token in the buffer, if it is not empty
	if s.buf != (Token{}) {
		itm := s.buf
		s.buf = Token{}
		return itm, nil
	}

	// scan text
	startPos, startCol, startLn := s.pos, s.col, s.ln

	isEndOfLine, err := s.readTo(s.ldelim, true)
	if err == io.EOF {
		if startPos == s.pos {
			return Token{}, io.EOF
		}
		text := Token{
			Type:      TEXT,
			Text:      s.src[startPos:s.pos],
			Offset:    startPos,
			EndOffset: s.pos,
			Column:    startCol,
			Line:      startLn,
		}
		return text, nil
	}

	if isEndOfLine {
		s.isNewLine = true
		text := Token{
			Type:      TEXT_EOL,
			Text:      s.src[startPos:s.pos],
			Offset:    startPos,
			EndOffset: s.pos,
			Column:    startCol,
			Line:      startLn,
		}
		return text, nil
	}

	s.pos -= len(s.ldelim)
	s.col -= len(s.ldelim)

	text := Token{
		Type:      TEXT,
		Text:      s.src[startPos:s.pos],
		Offset:    startPos,
		EndOffset: s.pos,
		Column:    startCol,
		Line:      startLn,
	}

	// scan tag
	startPos, startCol, startLn = s.pos, s.col, s.ln
	isStandaloneTag := false

	s.pos += len(s.ldelim)
	s.col += len(s.ldelim)

	var tagSymbol byte
	if s.pos < len(s.src) {
		tagSymbol = s.src[s.pos]
	}

	var tagType Type
	var tagText string

	switch tagSymbol {
	case '{':
		_, err = s.readTo("}"+s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = UNESCAPED_VARIABLE
		key := s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)-1]
		key = strings.TrimSpace(key)
		err = s.validateDottedKey(startLn, startCol, key)
		if err != nil {
			return Token{}, err
		}
		tagText = key

	case '&':
		_, err = s.readTo(s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = UNESCAPED_VARIABLE_SYM
		key := s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)]
		key = strings.TrimSpace(key)
		err = s.validateDottedKey(startLn, startCol, key)
		if err != nil {
			return Token{}, err
		}
		tagText = key

	case '#':
		_, err = s.readTo(s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = SECTION
		key := s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)]
		key = strings.TrimSpace(key)
		err = s.validateDottedKey(startLn, startCol, key)
		if err != nil {
			return Token{}, err
		}
		tagText = key

	case '^':
		_, err = s.readTo(s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = INVERTED_SECTION
		key := s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)]
		key = strings.TrimSpace(key)
		err = s.validateDottedKey(startLn, startCol, key)
		if err != nil {
			return Token{}, err
		}
		tagText = key

	case '/':
		_, err = s.readTo(s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = SECTION_END
		key := s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)]
		key = strings.TrimSpace(key)
		err = s.validateDottedKey(startLn, startCol, key)
		if err != nil {
			return Token{}, err
		}
		tagText = key

	case '>':
		_, err = s.readTo(s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = PARTIAL
		key := s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)]
		key = strings.TrimSpace(key)
		err = s.validatePartialKey(startLn, startCol, key)
		if err != nil {
			return Token{}, err
		}
		tagText = key

	case '=':
		_, err = s.readTo("="+s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		delims := s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)-1]
		delims = strings.TrimSpace(delims)
		parts := strings.Split(delims, " ")
		if len(parts) == 2 {
			s.ldelim = parts[0]
			s.rdelim = parts[1]
		}
		tagType = SET_DELIMETERS
		tagText = delims

	case '!':
		_, err = s.readTo(s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = COMMENT
		tagText = s.src[startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)]
		tagText = strings.TrimSpace(tagText)

	default:
		_, err = s.readTo(s.rdelim, false)
		if err != nil {
			return Token{}, s.error(startLn, startCol, "unclosed tag")
		}
		tagType = VARIABLE
		key := s.src[startPos+len(s.ldelim) : s.pos-len(s.rdelim)]
		key = strings.TrimSpace(key)
		err = s.validateDottedKey(startLn, startCol, key)
		if err != nil {
			return Token{}, err
		}
		tagText = key
	}

	tag := Token{
		Type:      tagType,
		Text:      tagText,
		Offset:    startPos,
		EndOffset: s.pos,
		Column:    startCol,
		Line:      startLn,
	}

	if s.isNewLine {
		s.isNewLine = false

		if isStandaloneTagSymbol(tagSymbol) && s.hasLeftPadding(startPos) {
			endOfLinePos, ok := s.hasRightPadding(s.pos)
			if ok {
				isStandaloneTag = true
				s.pos = endOfLinePos
				s.col = 1
				s.ln++
				s.isNewLine = true
			}
		}
	}

	if isStandaloneTag {
		tag.Indent = text.Text
		return tag, nil
	}

	if len(text.Text) == 0 {
		return tag, nil
	}

	s.buf = tag
	return text, nil
}

func (s *Scanner) error(ln, col int, msg string) error {
	var b strings.Builder
	b.WriteString(s.name)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(ln))
	b.WriteString(":")
	b.WriteString(strconv.Itoa(col))
	b.WriteString(":")
	b.WriteString(" ")
	b.WriteString(msg)
	return errors.New(b.String())
}

func isStandaloneTagSymbol(b byte) bool {
	switch b {
	case '#', '^', '/', '>', '=', '!':
		return true
	default:
		return false
	}
}

func (s *Scanner) hasLeftPadding(pos int) bool {
	var b byte
	for {
		if pos <= 0 {
			return true
		}
		pos--

		b = s.src[pos]
		if b != ' ' && b != '\t' {
			return b == '\n'
		}
	}
}

func (s *Scanner) hasRightPadding(pos int) (int, bool) {
	var b byte
	for {
		if pos >= len(s.src) {
			return pos, true
		}
		b = s.src[pos]
		pos++

		if b != ' ' && b != '\t' && b != '\r' {
			return pos, (b == '\n')
		}
	}
}

func (s *Scanner) validatePartialKey(ln, col int, raw string) error {
	if len(raw) == 0 {
		return s.error(ln, col, "missing key")
	}
	for i := range raw {
		switch raw[i] {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// good char, do nothing
		default:
			return s.error(ln, col, "invalid key: "+raw)
		}
	}
	return nil
}

func (s *Scanner) validateDottedKey(ln, col int, raw string) error {
	if len(raw) == 0 {
		return s.error(ln, col, "missing key")
	}
	if raw == "." {
		return nil
	}
	isValid := false
Loop:
	for i := range raw {
		switch raw[i] {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			isValid = true
		case '.':
			if i == 0 {
				break Loop
			}
			isValid = false
		default:
			isValid = false
			break Loop
		}
	}
	if !isValid {
		return s.error(ln, col, "invalid key: "+raw)
	}
	return nil
}
