package grammar

import "testing"

func TestStep8HindiVocabularyMappings(t *testing.T) {
	v := Must(Hindi)
	for surface, want := range map[string]Keyword{
		"बनाओ": KeywordCreate, "निर्माण": KeywordCreate,
		"चर": KeywordVariable, "मान": KeywordVariable,
		"स्थिर": KeywordConstant, "अचर": KeywordConstant,
		"फलन": KeywordFunction, "कार्य": KeywordFunction, "फ़ंक्शन": KeywordFunction,
		"रूप": KeywordAs, "जैसा": KeywordAs, "के_रूप_में": KeywordAs,
		"लौटाओ": KeywordReturn, "वापस": KeywordReturn,
		"अगर": KeywordIf, "यदि": KeywordIf,
		"वरना": KeywordElse, "अन्यथा": KeywordElse,
		"आयात": KeywordImport, "लाओ": KeywordImport,
		"से":  KeywordFrom,
		"घटक": KeywordComponent, "अवयव": KeywordComponent,
	} {
		if got, ok := v.Lookup(surface); !ok || got != want {
			t.Fatalf("%q = %v,%v want %v,true", surface, got, ok, want)
		}
	}
}

func TestStep8HindiKeywordsAreVocabularyScoped(t *testing.T) {
	hi := Must(Hindi)
	en := Must(English)
	bn := Must(Bengali)
	if !hi.IsKeyword("बनाओ") || !hi.IsKeyword("चर") || !hi.IsKeyword("घटक") {
		t.Fatalf("Hindi vocabulary did not reserve Hindi keywords")
	}
	if en.IsKeyword("बनाओ") || en.IsKeyword("चर") || bn.IsKeyword("बनाओ") || bn.IsKeyword("चर") {
		t.Fatalf("Hindi keywords must not be reserved outside Hindi vocabulary")
	}
	if hi.IsKeyword("उपयोगकर्ता") || hi.IsKeyword("परिणाम") {
		t.Fatalf("unknown Hindi words must remain identifiers")
	}
}

func TestStep8AllCanonicalKeywordsHaveHindiMappings(t *testing.T) {
	want := map[Keyword]bool{}
	for k := KeywordCreate; k <= KeywordComponent; k++ {
		want[k] = false
	}
	for _, k := range Must(Hindi).Entries() {
		want[k] = true
	}
	for k, ok := range want {
		if !ok {
			t.Fatalf("missing Hindi mapping for %s", k)
		}
	}
}

func TestStep8HindiVocabularyEntriesAreExactlyIntended(t *testing.T) {
	want := map[string]Keyword{
		"बनाओ": KeywordCreate, "निर्माण": KeywordCreate,
		"चर": KeywordVariable, "मान": KeywordVariable,
		"स्थिर": KeywordConstant, "अचर": KeywordConstant,
		"फलन": KeywordFunction, "कार्य": KeywordFunction, "फ़ंक्शन": KeywordFunction,
		"रूप": KeywordAs, "जैसा": KeywordAs, "के_रूप_में": KeywordAs,
		"लौटाओ": KeywordReturn, "वापस": KeywordReturn,
		"अगर": KeywordIf, "यदि": KeywordIf,
		"वरना": KeywordElse, "अन्यथा": KeywordElse,
		"आयात": KeywordImport, "लाओ": KeywordImport,
		"से":  KeywordFrom,
		"घटक": KeywordComponent, "अवयव": KeywordComponent,
	}
	got := Must(Hindi).Entries()
	if len(got) != len(want) {
		t.Fatalf("Hindi entries count = %d want %d: %#v", len(got), len(want), got)
	}
	for surface, wantKeyword := range want {
		if gotKeyword, ok := got[surface]; !ok || gotKeyword != wantKeyword {
			t.Fatalf("Hindi entry %q = %v,%v want %v,true", surface, gotKeyword, ok, wantKeyword)
		}
	}
}
