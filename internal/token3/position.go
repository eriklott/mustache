package token3

import "fmt"

// Position represents a cursor location in the source template string
type Position struct {
	Offset int // Number of bytes away from the beginning of the string.
	Line   int // Line number of the cursor position.
	Column int // Column number of the cursor position.
}

func (pos Position) String() string {
	return fmt.Sprintf("%d:%d", pos.Line, pos.Column)
}
