package grammar

import "testing"

func TestVocabularyLookup(t *testing.T) {
	tests := []struct {
		lang    Language
		surface string
		want    Keyword
	}{
		{English, "create", KeywordCreate}, {English, "variable", KeywordVariable},
		{Bengali, "তৈরি", KeywordCreate}, {Bengali, "চলক", KeywordVariable},
		{Hindi, "बनाओ", KeywordCreate}, {Hindi, "चर", KeywordVariable},
	}
	for _, tt := range tests {
		v := Must(tt.lang)
		got, ok := v.Lookup(tt.surface)
		if !ok || got != tt.want {
			t.Fatalf("%s lookup %q = %v, %v; want %v, true", tt.lang, tt.surface, got, ok, tt.want)
		}
	}
}

func TestUnknownWordsRemainIdentifiers(t *testing.T) {
	for _, lang := range []Language{English, Bengali, Hindi} {
		if Must(lang).IsKeyword("ব্যবহারকারী") || Must(lang).IsKeyword("उपयोगकर्ता") || Must(lang).IsKeyword("user") && lang != English {
			t.Fatalf("unexpected identifier reserved in %s", lang)
		}
	}
}
