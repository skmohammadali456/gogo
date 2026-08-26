package diagnostics

import (
	"strings"
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestRendererRendersZeroLengthEOFSpanAfterTrailingNewline(t *testing.T) {
	files := source.NewFileMap()
	id := files.Add("main.gogo", "create\n")
	span := source.Span{Start: source.PositionAt("create\n", len("create\n")), End: source.PositionAt("create\n", len("create\n"))}
	got := (Renderer{Files: files}).Text([]Diagnostic{{Severity: Error, FileID: id, Code: "G2999", Message: "unexpected end", Span: span}})
	if !strings.Contains(got, "main.gogo:2:1") || !strings.Contains(got, "   2 | \n") || !strings.Contains(got, "     | ^\n") {
		t.Fatalf("EOF diagnostic did not render final empty line and caret:\n%s", got)
	}
}
