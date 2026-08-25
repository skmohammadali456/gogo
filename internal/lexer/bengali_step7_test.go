package lexer

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

func TestStep7BengaliUnicodeIdentifierTokenTextAndSpans(t *testing.T) {
	text := "ব্যবহারকারী নাম বার্তা ফলাফল সংখ্যা userবাংলা বাংলাUser ব্যবহারকারী123"
	lx := New(source.File{ID: 1, Path: "bn-identifiers.gogo", Text: text})
	tokens := lx.LexAll()
	want := []string{"ব্যবহারকারী", "নাম", "বার্তা", "ফলাফল", "সংখ্যা", "userবাংলা", "বাংলাUser", "ব্যবহারকারী123"}
	if len(tokens) < len(want)+1 {
		t.Fatalf("got too few tokens: %#v", tokens)
	}
	for i, w := range want {
		got := tokens[i]
		if got.Kind != token.Identifier || got.Text != w {
			t.Fatalf("token %d = %s %q want Identifier %q", i, got.Kind, got.Text, w)
		}
		start := source.PositionAt(text, got.Span.Start.Offset)
		end := source.PositionAt(text, got.Span.End.Offset)
		if got.Span.Start != start || got.Span.End != end {
			t.Fatalf("token %q span=%#v want %#v..%#v", w, got.Span, start, end)
		}
		if text[got.Span.Start.Offset:got.Span.End.Offset] != w {
			t.Fatalf("source bytes for %q were not preserved", w)
		}
	}
}
