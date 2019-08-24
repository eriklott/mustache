package token

import "fmt"

type Position struct {
	Offset int // offset, starting at 0
	Line   int // line number, starting at 1
	Column int // column number, starting at 1 (byte count)
}

// IsValid reports whether the position is valid.
func (pos Position) IsValid() bool {
	return pos.Line > 0
}

func (pos Position) String() string {
	var s string
	if pos.IsValid() {
		s = fmt.Sprintf("%d", pos.Line)
		if pos.Column != 0 {
			s += fmt.Sprintf(":%d", pos.Column)
		}
	}
	if s == "" {
		s = "-"
	}
	return s
}
