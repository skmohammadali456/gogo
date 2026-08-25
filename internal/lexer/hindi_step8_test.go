package lexer

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

func TestStep8HindiIdentifiersPreserveTextBytesAndSpans(t *testing.T) {
	text := "उपयोगकर्ता नाम संदेश परिणाम संख्या userहिंदी हिंदीUser उपयोगकर्ता123"
	lx := New(source.File{ID: 1, Path: "hi.gogo", Text: text})
	tokens := lx.LexAll()
	wants := []string{"उपयोगकर्ता", "नाम", "संदेश", "परिणाम", "संख्या", "userहिंदी", "हिंदीUser", "उपयोगकर्ता123"}
	for i, want := range wants {
		got := tokens[i]
		if got.Kind != token.Identifier || got.Text != want {
			t.Fatalf("token %d = %s %q want Identifier %q", i, got.Kind, got.Text, want)
		}
		if text[got.Span.Start.Offset:got.Span.End.Offset] != want {
			t.Fatalf("token %d bytes changed: %q", i, text[got.Span.Start.Offset:got.Span.End.Offset])
		}
		if got.Span.Start != source.PositionAt(text, got.Span.Start.Offset) || got.Span.End != source.PositionAt(text, got.Span.End.Offset) {
			t.Fatalf("token %d span inconsistent: %#v", i, got.Span)
		}
	}
}

func TestStep8HindiAsAliasWithUnderscoresIsSingleIdentifierToken(t *testing.T) {
	text := "के_रूप_में"
	lx := New(source.File{ID: 1, Path: "hi.gogo", Text: text})
	tokens := lx.LexAll()
	if len(tokens) < 2 || tokens[0].Kind != token.Identifier || tokens[0].Text != text {
		t.Fatalf("first token = %s %q, want single Identifier %q", tokens[0].Kind, tokens[0].Text, text)
	}
	if tokens[0].Span.Start.Offset != 0 || tokens[0].Span.End.Offset != len(text) {
		t.Fatalf("single-token span = %#v, want byte range [0,%d)", tokens[0].Span, len(text))
	}
}
