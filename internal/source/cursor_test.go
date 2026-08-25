package source

import "testing"

func TestCursorWalksUnicode(t *testing.T) {
	c := NewCursor("Go বাংলা हिन्दी")
	start := c.Offset
	r, size := c.Advance()
	if r != 'G' || size != 1 || c.Offset != start+1 {
		t.Fatalf("unexpected first rune: %q size=%d offset=%d", r, size, c.Offset)
	}

	for !c.Done() {
		_, size = c.Advance()
		if size <= 0 {
			t.Fatal("cursor failed to advance")
		}
	}
	if c.Offset != len(c.Text) {
		t.Fatalf("cursor ended at %d, want %d", c.Offset, len(c.Text))
	}
}

func TestCursorMatchAndSlice(t *testing.T) {
	c := NewCursor("create variable")
	if !c.Match("create") {
		t.Fatal("expected match")
	}
	if c.Match("missing") {
		t.Fatal("unexpected match")
	}
	if got := c.Slice(0); got != "create" {
		t.Fatalf("unexpected slice: %q", got)
	}
}
