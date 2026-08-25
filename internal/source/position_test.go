package source

import "testing"

func TestNewPosition(t *testing.T) {
	p := NewPosition()
	if p.Offset != 0 || p.Line != 1 || p.Column != 1 {
		t.Fatalf("unexpected initial position: %+v", p)
	}
}

func TestSpanValidity(t *testing.T) {
	valid := Span{
		Start: Position{Offset: 0, Line: 1, Column: 1},
		End:   Position{Offset: 3, Line: 1, Column: 4},
	}
	if !valid.IsValid() {
		t.Fatal("expected valid span")
	}

	invalid := Span{
		Start: Position{Offset: 4, Line: 1, Column: 5},
		End:   Position{Offset: 3, Line: 1, Column: 4},
	}
	if invalid.IsValid() {
		t.Fatal("expected invalid span")
	}
}
