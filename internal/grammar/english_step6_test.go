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

func TestStep8EnglishVocabularyEntriesAreExactlyIntended(t *testing.T) {
	want := map[string]Keyword{
		"create":   KeywordCreate,
		"variable": KeywordVariable, "let": KeywordVariable, "var": KeywordVariable,
		"constant": KeywordConstant, "const": KeywordConstant,
		"function": KeywordFunction, "fn": KeywordFunction,
		"as":        KeywordAs,
		"return":    KeywordReturn,
		"if":        KeywordIf,
		"else":      KeywordElse,
		"import":    KeywordImport,
		"from":      KeywordFrom,
		"component": KeywordComponent,
	}
	got := Must(English).Entries()
	if len(got) != len(want) {
		t.Fatalf("English entries count = %d want %d: %#v", len(got), len(want), got)
	}
	for surface, wantKeyword := range want {
		if gotKeyword, ok := got[surface]; !ok || gotKeyword != wantKeyword {
			t.Fatalf("English entry %q = %v,%v want %v,true", surface, gotKeyword, ok, wantKeyword)
		}
	}
}
