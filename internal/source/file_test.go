package source

import "testing"

func TestFileMapAssignsStableIDs(t *testing.T) {
	m := NewFileMap()
	first := m.Add("main.gogo", "text")
	second := m.Add("other.gogo", "text")
	repeated := m.Add("main.gogo", "updated")

	if first != 1 || second != 2 || repeated != first {
		t.Fatalf("unexpected IDs: %d %d %d", first, second, repeated)
	}
	if m.Count() != 2 {
		t.Fatalf("expected 2 files, got %d", m.Count())
	}
	file, ok := m.Get(first)
	if !ok || file.Text != "updated" {
		t.Fatalf("file was not updated: %+v", file)
	}
}

func TestPositionAtUnicode(t *testing.T) {
	text := "hello\nবাংলা\nहिन्दी"

	p := PositionAt(text, len("hello\n"))
	if p.Line != 2 || p.Column != 1 {
		t.Fatalf("unexpected second-line position: %+v", p)
	}

	bengaliOffset := len("hello\nব")
	p = PositionAt(text, bengaliOffset)
	if p.Line != 2 || p.Column != 2 {
		t.Fatalf("unexpected Bengali position: %+v", p)
	}
}

func TestLineStartOffsets(t *testing.T) {
	text := "a\nবাংলা\nxyz"
	offsets := LineStartOffsets(text)
	if len(offsets) != 3 || offsets[0] != 0 || offsets[1] != 2 {
		t.Fatalf("unexpected line offsets: %v", offsets)
	}
}
