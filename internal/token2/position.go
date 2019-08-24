package token2

import "fmt"

type Position struct {
	Offset int // offset, starting at 0
	Line   int // line number, starting at 1
	Column int // column number, starting at 1 (byte count)
}

func (pos Position) String() string {
	return fmt.Sprintf("%d:%d", pos.Line, pos.Column)
}
