// Copyright 2017 Erik Lott. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package token

import (
	"reflect"
	"testing"
)

func TestStringSelectionReader_readRune(t *testing.T) {
	reader := newStringSelectionReader("hello")
	runes := []rune{}
	runes = append(runes, reader.readRune())
	runes = append(runes, reader.readRune())
	runesWant := []rune{rune('h'), rune('e')}
	if !reflect.DeepEqual(runes, runesWant) {
		t.Errorf("unexpected rune, got: %v, want: %v", runes, runesWant)
	}
}

func TestStringSelectionReader_peekRune(t *testing.T) {
	reader := newStringSelectionReader("hello")
	runes := []rune{}
	runes = append(runes, reader.peekRune())
	runes = append(runes, reader.peekRune())
	runesWant := []rune{rune('h'), rune('h')}
	if !reflect.DeepEqual(runes, runesWant) {
		t.Errorf("unexpected rune, got: %v, want: %v", runes, runesWant)
	}
}

func TestStringSelectionReader_hasPrefix(t *testing.T) {
	reader := newStringSelectionReader("hello")
	if !reader.hasPrefix("hel") {
		t.Error("expected prefix to exist")
	}
}

func TestStringSelectionReader_selection(t *testing.T) {
	reader := newStringSelectionReader("hello")
	reader.readRune()
	reader.readRune()
	reader.readRune()
	got := reader.selection()
	want := "hel"
	if got != want {
		t.Errorf("unexpected selection got: %s, want %s", got, want)
	}
}
