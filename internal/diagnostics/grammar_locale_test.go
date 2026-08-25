package diagnostics

import (
	"strings"
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestGrammarDiagnosticLocalizesAllStep5Locales(t *testing.T) {
	files := source.NewFileMap()
	id := files.Add("mixed.gogo", "create চলক user হিসেবে 1")
	d := Diagnostic{Severity: Error, FileID: id, Code: "G2035", Message: "I found extra tokens after this expression statement.", Hint: "Separate statements with semicolons or use the selected grammar vocabulary.", Span: source.Span{Start: source.PositionAt("create চলক user হিসেবে 1", len("create ")), End: source.PositionAt("create চলক user হিসেবে 1", len("create চলক"))}}
	cases := []struct {
		locale Locale
		want   string
	}{
		{English, "extra tokens"},
		{Bengali, "অতিরিক্ত token"},
		{Hindi, "अतिरिक्त token"},
	}
	for _, tc := range cases {
		out := Renderer{Files: files, Locale: tc.locale}.Text([]Diagnostic{d})
		if !strings.Contains(out, "G2035") || !strings.Contains(out, tc.want) {
			t.Fatalf("%s grammar diagnostic not localized or code missing:\n%s", tc.locale, out)
		}
	}
}
