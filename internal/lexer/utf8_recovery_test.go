package lexer

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

func TestMalformedUTF8InStringsAndCommentsProducesDiagnosticAndRecovers(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []token.Kind
	}{
		{"string", "before \"bad" + string([]byte{0xff}) + "\" after", []token.Kind{token.Identifier, token.Invalid, token.Identifier, token.EOF}},
		{"line comment", "before // bad" + string([]byte{0xff}) + "\nafter", []token.Kind{token.Identifier, token.Identifier, token.EOF}},
		{"block comment", "before /* bad" + string([]byte{0xff}) + " */ after", []token.Kind{token.Identifier, token.Identifier, token.EOF}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(source.File{ID: 1, Path: "bad-utf8.gogo", Text: tt.text})
			got := l.LexAll()
			if len(got) != len(tt.want) {
				t.Fatalf("tokens = %#v, want %d", got, len(tt.want))
			}
			for i, kind := range tt.want {
				if got[i].Kind != kind {
					t.Fatalf("token %d = %s, want %s", i, got[i].Kind, kind)
				}
			}
			diags := l.Diagnostics()
			if len(diags) != 1 || diags[0].Code != "G1000" {
				t.Fatalf("diagnostics = %#v, want one G1000", diags)
			}
		})
	}
}
