package token

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

type Token int

const (
	TEXT Token = iota
	NEWLINE
	TAG
)

type state int

const (
	stateText state = iota
	stateNewline
	stateCarriageReturn
	stateTag
)

var (
	eol = errors.New("end of line")
	eof = errors.New("end of file")
)

type Scanner struct {
	name   string
	src    string
	ldelim string
	rdelim string

	nextLDelim string
	nextRDelim string

	startPos int
	startCol int
	startLn  int

	pos int
	col int
	ln  int

	lastCol int

	state state
}

func NewScanner(name, src, ldelim, rdelim string) *Scanner {
	return &Scanner{
		name:       name,
		src:        src,
		ldelim:     ldelim,
		rdelim:     rdelim,
		nextLDelim: ldelim,
		nextRDelim: rdelim,
		startPos:   0,
		startCol:   1,
		startLn:    1,
		pos:        0,
		col:        1,
		ln:         1,
		lastCol:    1,
		state:      stateText,
	}
}

func (s *Scanner) Name() string {
	return s.name
}

type Item struct {
	Token    Token
	Text     string
	LDelim   string
	RDelim   string
	StartPos int
	StartCol int
	StartLn  int
	EndPos   int
	Indent   string
}

func (s *Scanner) Next() (Item, error) {
	s.startPos, s.startCol, s.startLn = s.pos, s.col, s.ln
	s.ldelim, s.rdelim = s.nextLDelim, s.nextRDelim

	var t Token
	var err error
	switch s.state {
	case stateText:
		t, err = s.scanText()
	case stateNewline:
		t, err = s.scanNewline()
	case stateCarriageReturn:
		t, err = s.scanCarriageReturn()
	case stateTag:
		t, err = s.scanTag()
	}
	if err != nil {
		return Item{}, err
	}

	i := Item{
		Token:    t,
		Text:     s.src[s.startPos:s.pos],
		LDelim:   s.ldelim,
		RDelim:   s.rdelim,
		StartPos: s.startPos,
		StartCol: s.startCol,
		StartLn:  s.startLn,
		EndPos:   s.pos,
	}

	return i, err
}

func (s *Scanner) readTo(pattern string, haltEOL bool) error {
	matchIdx := 0
	for {
		if s.pos >= len(s.src) {
			return eof
		}
		b := s.src[s.pos]
		s.pos++
		s.col++
		if b == '\n' {
			s.lastCol = s.col
			s.col = 1
			s.ln++
			if haltEOL {
				return eol
			}
		}

		if b != pattern[matchIdx] {
			matchIdx = 0
			continue
		}

		matchIdx++
		if matchIdx == len(pattern) {
			return nil
		}
	}
}

func (s *Scanner) scanText() (Token, error) {
	err := s.readTo(s.ldelim, true)

	if err == eof {
		if s.startPos == s.pos {
			return 0, io.EOF
		}
		return TEXT, nil
	}

	if err == eol {
		if s.pos > 1 && s.src[s.pos-2] == '\r' {
			s.pos -= 2
			s.col = s.lastCol - 1
			s.ln--
			s.state = stateCarriageReturn
			if s.startPos == s.pos {
				return s.scanCarriageReturn()
			}
			return TEXT, nil
		}

		s.pos--
		s.col = s.lastCol
		s.ln--
		s.state = stateNewline
		if s.startPos == s.pos {
			return s.scanNewline()
		}
		return TEXT, nil
	}

	s.pos -= len(s.ldelim)
	s.col -= len(s.ldelim)

	if s.startPos == s.pos {
		return s.scanTag()
	}

	s.state = stateTag
	return TEXT, nil
}
func (s *Scanner) scanNewline() (Token, error) {
	s.pos++
	s.col = 1
	s.ln++
	s.state = stateText
	return NEWLINE, nil
}

func (s *Scanner) scanCarriageReturn() (Token, error) {
	s.pos += 2
	s.col = 1
	s.ln++
	s.state = stateText
	return NEWLINE, nil
}

func (s *Scanner) scanTag() (Token, error) {
	s.pos += len(s.ldelim)
	s.col += len(s.ldelim)

	var err error
	if s.pos < len(s.src) && s.src[s.pos] == '{' {
		err = s.readTo("}"+s.rdelim, false)
	} else if s.pos < len(s.src) && s.src[s.pos] == '=' {
		err = s.readTo("="+s.rdelim, false)
		if err == nil {
			tagBody := s.src[s.startPos+len(s.ldelim)+1 : s.pos-len(s.rdelim)-1]
			tagBody = strings.TrimSpace(tagBody)
			newDelims := strings.Split(tagBody, " ")
			if len(newDelims) == 2 {
				s.nextLDelim = newDelims[0]
				s.nextRDelim = newDelims[1]
			}
		}

	} else {
		err = s.readTo(s.rdelim, false)
	}

	if err != nil {
		return 0, s.error("unclosed tag")
	}

	s.state = stateText
	return TAG, nil
}

func (s *Scanner) error(msg string) error {
	var b strings.Builder
	b.WriteString(s.name)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(s.startLn))
	b.WriteString(":")
	b.WriteString(strconv.Itoa(s.startCol))
	b.WriteString(":")
	b.WriteString(" ")
	b.WriteString(msg)
	return errors.New(b.String())
}
