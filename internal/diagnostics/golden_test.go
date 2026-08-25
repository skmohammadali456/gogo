package diagnostics

import (
	"os"
	"strings"
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestGoldenMultilineDiagnosticRendering(t *testing.T) {
	text := "create function greet() {\n  return \"hi\""
	files := source.NewFileMap()
	id := files.Add("main.gogo", text)
	d := Diagnostic{
		Severity: Error,
		FileID:   id,
		Code:     "G2006",
		Message:  "This block is missing its closing brace.",
		Span:     source.Span{Start: source.PositionAt(text, 0), End: source.PositionAt(text, len(text))},
		Labels: []Label{
			{Style: Primary, Span: source.Span{Start: source.PositionAt(text, 0), End: source.PositionAt(text, 24)}, Message: "open block starts here"},
			{Style: Secondary, Span: source.Span{Start: source.PositionAt(text, len(text)), End: source.PositionAt(text, len(text))}, Message: "parser reached end of file"},
		},
		Hint:        "Add } to close the block.",
		Suggestions: []Suggestion{{Message: "insert closing brace", Edits: []FixIt{{Span: source.Span{Start: source.PositionAt(text, len(text)), End: source.PositionAt(text, len(text))}, Replacement: "\n}"}}}},
	}
	got := Renderer{Files: files, Locale: English}.Text([]Diagnostic{d})
	want, err := os.ReadFile("testdata/multiline.golden")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("golden mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestCatalogCoversCurrentCompilerLexerParserCodesAndLocales(t *testing.T) {
	codes := []Code{"G0001", "G0002", "G0003", "G0004", "G0005", "G1000", "G1001", "G1002", "G1004", "G2000", "G2001", "G2002", "G2003", "G2004", "G2005", "G2006", "G2007", "G2008", "G2009", "G2010", "G2011", "G2012", "G2013", "G2014", "G2015", "G2016", "G2017", "G2018", "G2020", "G2021", "G2022", "G2023", "G2024", "G2025", "G2026", "G2027", "G2028", "G2029", "G2030", "G2031", "G2032", "G2033", "G2034"}
	for _, code := range codes {
		entry, ok := Catalog[code]
		if !ok {
			t.Fatalf("missing catalog entry for %s", code)
		}
		if entry.Severity != Error {
			t.Fatalf("%s severity = %s, want error", code, entry.Severity)
		}
		for _, locale := range []Locale{English, Bengali, Hindi} {
			if strings.TrimSpace(entry.Messages[locale]) == "" {
				t.Fatalf("%s missing %s message", code, locale)
			}
			if strings.TrimSpace(entry.Hints[locale]) == "" {
				t.Fatalf("%s missing %s hint", code, locale)
			}
		}
	}
}
