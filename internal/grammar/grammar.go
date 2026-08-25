// Package grammar maps source-language vocabulary to canonical GOGO grammar symbols.
package grammar

import "fmt"

// Language identifies a data-driven surface vocabulary.
type Language string

const (
	English Language = "en"
	Bengali Language = "bn"
	Hindi   Language = "hi"
)

// Keyword is a canonical semantic grammar symbol, independent of surface wording.
type Keyword uint8

const (
	KeywordUnknown Keyword = iota
	KeywordCreate
	KeywordVariable
	KeywordFunction
	KeywordAs
	KeywordReturn
	KeywordIf
	KeywordElse
)

func (k Keyword) String() string {
	switch k {
	case KeywordCreate:
		return "create"
	case KeywordVariable:
		return "variable"
	case KeywordFunction:
		return "function"
	case KeywordAs:
		return "as"
	case KeywordReturn:
		return "return"
	case KeywordIf:
		return "if"
	case KeywordElse:
		return "else"
	default:
		return "unknown"
	}
}

// Entry defines one surface spelling for a canonical grammar keyword.
type Entry struct {
	Surface string
	Keyword Keyword
}

// Vocabulary maps one source language's surface words to canonical keywords.
type Vocabulary struct {
	Language Language
	Name     string
	entries  map[string]Keyword
}

func NewVocabulary(language Language, name string, entries []Entry) Vocabulary {
	v := Vocabulary{Language: language, Name: name, entries: make(map[string]Keyword, len(entries))}
	for _, e := range entries {
		if e.Surface == "" || e.Keyword == KeywordUnknown {
			continue
		}
		v.entries[e.Surface] = e.Keyword
	}
	return v
}

func (v Vocabulary) Lookup(surface string) (Keyword, bool) {
	kw, ok := v.entries[surface]
	return kw, ok
}

func (v Vocabulary) IsKeyword(surface string) bool {
	_, ok := v.Lookup(surface)
	return ok
}

func (v Vocabulary) Entries() map[string]Keyword {
	out := make(map[string]Keyword, len(v.entries))
	for s, k := range v.entries {
		out[s] = k
	}
	return out
}

var vocabularies = map[Language]Vocabulary{
	English: NewVocabulary(English, "English", []Entry{
		{"create", KeywordCreate}, {"variable", KeywordVariable}, {"function", KeywordFunction},
		{"as", KeywordAs}, {"return", KeywordReturn}, {"if", KeywordIf}, {"else", KeywordElse},
	}),
	Bengali: NewVocabulary(Bengali, "Bengali", []Entry{
		{"তৈরি", KeywordCreate}, {"চলক", KeywordVariable}, {"ফাংশন", KeywordFunction},
		{"হিসেবে", KeywordAs}, {"ফেরত", KeywordReturn}, {"যদি", KeywordIf}, {"নইলে", KeywordElse},
	}),
	Hindi: NewVocabulary(Hindi, "Hindi", []Entry{
		{"बनाओ", KeywordCreate}, {"चर", KeywordVariable}, {"फ़ंक्शन", KeywordFunction},
		{"के_रूप_में", KeywordAs}, {"लौटाओ", KeywordReturn}, {"अगर", KeywordIf}, {"वरना", KeywordElse},
	}),
}

func DefaultVocabulary() Vocabulary { return vocabularies[English] }

func ForLanguage(language Language) (Vocabulary, error) {
	v, ok := vocabularies[language]
	if !ok {
		return Vocabulary{}, fmt.Errorf("unknown grammar language %q", language)
	}
	return v, nil
}

func Must(language Language) Vocabulary {
	v, err := ForLanguage(language)
	if err != nil {
		panic(err)
	}
	return v
}
