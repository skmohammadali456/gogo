package grammar

import "testing"

func TestStep7BengaliVocabularyMappings(t *testing.T) {
	v := Must(Bengali)
	for surface, want := range map[string]Keyword{
		"তৈরি": KeywordCreate, "ঘোষণা": KeywordCreate,
		"চলক": KeywordVariable, "ধরি": KeywordVariable,
		"ধ্রুবক": KeywordConstant, "অপরিবর্তনীয়": KeywordConstant,
		"ফাংশন": KeywordFunction, "কাজ": KeywordFunction,
		"হিসেবে": KeywordAs, "রূপে": KeywordAs,
		"ফেরত": KeywordReturn, "ফেরাও": KeywordReturn,
		"যদি":  KeywordIf,
		"নইলে": KeywordElse, "অন্যথায়": KeywordElse,
		"আমদানি": KeywordImport, "ইমপোর্ট": KeywordImport,
		"থেকে":       KeywordFrom,
		"কম্পোনেন্ট": KeywordComponent, "উপাদান": KeywordComponent,
	} {
		if got, ok := v.Lookup(surface); !ok || got != want {
			t.Fatalf("%q = %v,%v want %v,true", surface, got, ok, want)
		}
	}
}

func TestStep7BengaliKeywordsAreVocabularyScoped(t *testing.T) {
	bn := Must(Bengali)
	en := Must(English)
	if !bn.IsKeyword("তৈরি") || !bn.IsKeyword("চলক") {
		t.Fatalf("Bengali vocabulary did not reserve Bengali keywords")
	}
	if en.IsKeyword("তৈরি") || en.IsKeyword("চলক") {
		t.Fatalf("English vocabulary must not reserve Bengali keywords")
	}
	if bn.IsKeyword("ব্যবহারকারী") || bn.IsKeyword("ফলাফল") {
		t.Fatalf("unknown Bengali words must remain identifiers")
	}
}
