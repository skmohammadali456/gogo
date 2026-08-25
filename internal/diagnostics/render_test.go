package diagnostics

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestRendererTextSnippetCaretMultilineUTF8AndTranslations(t *testing.T) {
	files := source.NewFileMap()
	id := files.Add("main.gogo", "শুরু = 1\nদ্বিতীয় = ?\n")
	d := Diagnostic{Severity: Error, FileID: id, Code: "G1001", Message: "This block comment never closes.", Span: source.Span{Start: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 22), End: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 23)}, Notes: []string{"operators need right operands"}, Hints: []string{"Add */ to close the comment."}, Suggestions: []Suggestion{{Message: "insert a value", Edits: []FixIt{{Span: source.Span{Start: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 23), End: source.PositionAt("শুরু = 1\nদ্বিতীয় = ?\n", 23)}, Replacement: "0"}}}}}
	out := Renderer{Files: files, Locale: Bengali}.Text([]Diagnostic{d})
	for _, want := range []string{"error[G1001]", "এই ব্লক মন্তব্যটি শেষ হয়নি", "main.gogo:2:", "দ্বিতীয় = ?", "^", "note:", "hint:", "মন্তব্য বন্ধ করতে */ যোগ করুন", "suggestion:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRendererLocalizesEnglishBengaliHindiWithoutChangingCode(t *testing.T) {
	text := "create variable user as\n"
	files := source.NewFileMap()
	id := files.Add("main.gogo", text)
	d := Diagnostic{Severity: Error, FileID: id, Code: "G2004", Message: "I expected a value after as.", Hint: "Give the variable an initial value.", Span: source.Span{Start: source.PositionAt(text, len(text)), End: source.PositionAt(text, len(text))}}
	cases := []struct {
		locale Locale
		want   string
	}{{English, "I expected a value after as."}, {Bengali, "'as' এর পরে একটি মান প্রত্যাশিত।"}, {Hindi, "'as' के बाद एक मान अपेक्षित है।"}}
	for _, tc := range cases {
		out := Renderer{Files: files, Locale: tc.locale}.Text([]Diagnostic{d})
		if !strings.Contains(out, "G2004") || !strings.Contains(out, tc.want) {
			t.Fatalf("%s output missing stable code or localized text:\n%s", tc.locale, out)
		}
		if strings.Contains(out, "main.gogo") && strings.Contains(out, "G২০০৪") {
			t.Fatalf("code was localized in %s output: %s", tc.locale, out)
		}
	}
}

func TestBagOrdersAndDeduplicatesDiagnosticsWithoutSuppressingDistinctDiagnostics(t *testing.T) {
	var b Bag
	late := Diagnostic{Severity: Error, Code: "G2", Message: "late", Span: source.Span{Start: source.Position{Offset: 10, Line: 2, Column: 1}, End: source.Position{Offset: 11, Line: 2, Column: 2}}}
	early := Diagnostic{Severity: Error, Code: "G1", Message: "early", Span: source.Span{Start: source.Position{Offset: 1, Line: 1, Column: 2}, End: source.Position{Offset: 2, Line: 1, Column: 3}}}
	distinct := early
	distinct.Message = "still distinct"
	b.Add(late)
	b.Add(early)
	b.Add(early)
	b.Add(distinct)
	all := b.All()
	if len(all) != 3 || all[0].Code != "G1" || all[1].Code != "G1" || all[2].Code != "G2" {
		t.Fatalf("unexpected order/dedup: %#v", all)
	}
}

func TestRendererJSONUsesStableSchema(t *testing.T) {
	files := source.NewFileMap()
	id := files.Add("main.gogo", "x")
	d := Diagnostic{Severity: Warning, FileID: id, Code: "G9000", Message: "careful", Span: source.Span{Start: source.Position{Offset: 0, Line: 1, Column: 1}, End: source.Position{Offset: 1, Line: 1, Column: 2}}}
	data, err := (Renderer{Files: files, Locale: Hindi}).JSON([]Diagnostic{d})
	if err != nil {
		t.Fatal(err)
	}
	var got []JSONDiagnostic
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
	if len(got) != 1 || got[0].Severity != "warning" || got[0].Code != "G9000" || got[0].Language != Hindi || got[0].File != "main.gogo" || got[0].Span.Start.Offset != 0 || got[0].Span.End.Column != 2 || len(got[0].Labels) != 1 {
		t.Fatalf("unexpected JSON schema: %#v", got)
	}
}
