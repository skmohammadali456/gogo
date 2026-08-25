package grammar

import "testing"

func TestStep6EnglishAliases(t *testing.T) {
	v := Must(English)
	for surface, want := range map[string]Keyword{
		"create": KeywordCreate, "variable": KeywordVariable, "let": KeywordVariable, "var": KeywordVariable,
		"constant": KeywordConstant, "const": KeywordConstant, "function": KeywordFunction, "fn": KeywordFunction,
		"component": KeywordComponent, "import": KeywordImport, "from": KeywordFrom, "as": KeywordAs,
	} {
		if got, ok := v.Lookup(surface); !ok || got != want {
			t.Fatalf("%q = %v,%v want %v,true", surface, got, ok, want)
		}
	}
}
