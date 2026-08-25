package diagnostics

import (
	"strings"
	"testing"
	"time"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestSnippetHandlesSourceLocationEdgeCases(t *testing.T) {
	cases := []struct {
		name, text           string
		offset               int
		wantLine, wantColumn int
	}{
		{"ascii", "abc", 1, 1, 2},
		{"tab", "a\tb", 2, 1, 3},
		{"empty-line", "a\n\nb", 3, 3, 1},
		{"bengali", "নাম", len("না"), 1, 3},
		{"hindi", "नाम", len("ना"), 1, 3},
		{"mixed-bengali-hindi", "x নাম नाम", len("x নাম "), 1, 7},
		{"combining", "e\u0301x", len("e\u0301"), 1, 3},
		{"eof", "শেষ", len("শেষ"), 1, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pos := source.PositionAt(tc.text, tc.offset)
			if pos.Offset != tc.offset || pos.Line != tc.wantLine || pos.Column != tc.wantColumn {
				t.Fatalf("PositionAt(%q,%d)=%+v", tc.text, tc.offset, pos)
			}
			span := source.Span{Start: pos, End: pos}
			out := snippet(tc.text, Diagnostic{Severity: Error, Code: "G9000", Span: span})
			if tc.text != "" && !strings.Contains(out, "^") {
				t.Fatalf("missing caret in %q", out)
			}
		})
	}
}

func TestRendererDoesNotHangOnMalformedOrOutOfRangeSpans(t *testing.T) {
	files := source.NewFileMap()
	id := files.Add("bad.gogo", string([]byte{'a', 0xff, 'b', '\n'}))
	diags := []Diagnostic{{Severity: Error, FileID: id, Code: "G1000", Message: "bad", Span: source.Span{Start: source.Position{Offset: 99, Line: 99, Column: 99}, End: source.Position{Offset: 99, Line: 99, Column: 99}}}}
	done := make(chan string, 1)
	go func() { done <- Renderer{Files: files}.Text(diags) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("renderer hung")
	}
}
