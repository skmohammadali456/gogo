package diagnostics

import (
	"strings"
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestRendererTextSnippetCaretMultilineUTF8AndTranslations(t *testing.T) {
	files := source.NewFileMap()
	files.Add("main.gogo", "শুরু = 1\nদ্বিতীয় = ?\n")
	d := Diagnostic{Severity: Error, Code: "G1001", Message: "This block comment never closes.", Span: source.Span{Start: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 22), End: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 23)}, Notes: []string{"operators need right operands"}, Hints: []string{"Add */ to close the comment."}, Suggestions: []Suggestion{{Message: "insert a value", Edits: []FixIt{{Span: source.Span{Start: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 23), End: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 23)}, Replacement: "0"}}}}}
	out := Renderer{Files: files, Locale: Bengali}.Text([]Diagnostic{d})
	for _, want := range []string{"error[G1001]", "এই ব্লক মন্তব্যটি শেষ হয়নি", "main.gogo:2:", "দ্বিতীয় = ?", "^", "note:", "hint:", "মন্তব্য বন্ধ করতে */ যোগ করুন", "suggestion:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestBagOrdersAndDeduplicatesDiagnostics(t *testing.T) {
	var b Bag
	late := Diagnostic{Severity: Error, Code: "G2", Message: "late", Span: source.Span{Start: source.Position{Offset: 10, Line: 2, Column: 1}, End: source.Position{Offset: 11, Line: 2, Column: 2}}}
	early := Diagnostic{Severity: Error, Code: "G1", Message: "early", Span: source.Span{Start: source.Position{Offset: 1, Line: 1, Column: 2}, End: source.Position{Offset: 2, Line: 1, Column: 3}}}
	b.Add(late)
	b.Add(early)
	b.Add(early)
	all := b.All()
	if len(all) != 2 || all[0].Code != "G1" || all[1].Code != "G2" {
		t.Fatalf("unexpected order/dedup: %#v", all)
	}
}

func TestRendererJSONUsesStableShape(t *testing.T) {
	d := Diagnostic{Severity: Warning, FileID: 7, Code: "G9000", Message: "careful", Span: source.Span{Start: source.Position{Line: 1, Column: 1}, End: source.Position{Line: 1, Column: 2}}}
	data, err := (Renderer{}).JSON([]Diagnostic{d})
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	for _, want := range []string{`"severity": "warning"`, `"file_id": 7`, `"code": "G9000"`, `"labels"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
}
