package token

import "fmt"

type Position struct {
	Offset int
	Column int
	Line   int
}

// String returns a string in one of several forms:
//
//	line:column    valid position with file name
//	line           valid position with file name but no column (column == 0)
//
func (pos Position) String() string {
	s := fmt.Sprintf("%d", pos.Line)
	if pos.Column != 0 {
		s += fmt.Sprintf(":%d", pos.Column)
	}
	return s
}
