// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package token

import "fmt"

// Position represents the starting position of a node in the source text
type Position struct {
	File string
	Line int
	Col  int
}

func (p Position) String() string {
	file := p.File
	if file == "" {
		file = "<input>"
	}
	return fmt.Sprintf("%s:%d:%d", file, p.Line, p.Col)
}
