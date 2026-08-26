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
	KeywordConstant
	KeywordImport
	KeywordFrom
	KeywordComponent
	KeywordType
	KeywordReadonly
	KeywordEnum
	KeywordInterface
	KeywordExtends
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
	case KeywordConstant:
		return "constant"
	case KeywordImport:
		return "import"
	case KeywordFrom:
		return "from"
	case KeywordComponent:
		return "component"
	case KeywordType:
		return "type"
	case KeywordReadonly:
		return "readonly"
	case KeywordEnum:
		return "enum"
	case KeywordInterface:
		return "interface"
	case KeywordExtends:
		return "extends"
	default:
		return "unknown"
	}
}

// ParseKeyword resolves a canonical keyword name used by project alias configuration.
func ParseKeyword(name string) (Keyword, bool) {
	switch name {
	case "create":
		return KeywordCreate, true
	case "variable":
		return KeywordVariable, true
	case "function":
		return KeywordFunction, true
	case "as":
		return KeywordAs, true
	case "return":
		return KeywordReturn, true
	case "if":
		return KeywordIf, true
	case "else":
		return KeywordElse, true
	case "constant":
		return KeywordConstant, true
	case "import":
		return KeywordImport, true
	case "from":
		return KeywordFrom, true
	case "component":
		return KeywordComponent, true
	case "type":
		return KeywordType, true
	case "readonly":
		return KeywordReadonly, true
	case "enum":
		return KeywordEnum, true
	case "interface":
		return KeywordInterface, true
	case "extends":
		return KeywordExtends, true
	default:
		return KeywordUnknown, false
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
		{"constant", KeywordConstant}, {"import", KeywordImport}, {"from", KeywordFrom}, {"component", KeywordComponent}, {"type", KeywordType}, {"readonly", KeywordReadonly}, {"enum", KeywordEnum}, {"interface", KeywordInterface}, {"extends", KeywordExtends},
		{"let", KeywordVariable}, {"var", KeywordVariable}, {"const", KeywordConstant}, {"fn", KeywordFunction},
	}),
	Bengali: NewVocabulary(Bengali, "Bengali", []Entry{
		{"তৈরি", KeywordCreate}, {"ঘোষণা", KeywordCreate},
		{"চলক", KeywordVariable}, {"ধরি", KeywordVariable},
		{"ধ্রুবক", KeywordConstant}, {"অপরিবর্তনীয়", KeywordConstant},
		{"ফাংশন", KeywordFunction}, {"কাজ", KeywordFunction},
		{"হিসেবে", KeywordAs}, {"রূপে", KeywordAs},
		{"ফেরত", KeywordReturn}, {"ফেরাও", KeywordReturn},
		{"যদি", KeywordIf},
		{"নইলে", KeywordElse}, {"অন্যথায়", KeywordElse},
		{"আমদানি", KeywordImport}, {"ইমপোর্ট", KeywordImport},
		{"থেকে", KeywordFrom},
		{"কম্পোনেন্ট", KeywordComponent}, {"উপাদান", KeywordComponent}, {"ধরন", KeywordType}, {"শুধু_পঠন", KeywordReadonly}, {"এনাম", KeywordEnum}, {"ইন্টারফেস", KeywordInterface}, {"প্রসারিত", KeywordExtends},
	}),
	Hindi: NewVocabulary(Hindi, "Hindi", []Entry{
		{"बनाओ", KeywordCreate}, {"निर्माण", KeywordCreate},
		{"चर", KeywordVariable}, {"मान", KeywordVariable},
		{"स्थिर", KeywordConstant}, {"अचर", KeywordConstant},
		{"फलन", KeywordFunction}, {"कार्य", KeywordFunction}, {"फ़ंक्शन", KeywordFunction},
		{"रूप", KeywordAs}, {"जैसा", KeywordAs}, {"के_रूप_में", KeywordAs},
		{"लौटाओ", KeywordReturn}, {"वापस", KeywordReturn},
		{"अगर", KeywordIf}, {"यदि", KeywordIf},
		{"वरना", KeywordElse}, {"अन्यथा", KeywordElse},
		{"आयात", KeywordImport}, {"लाओ", KeywordImport},
		{"से", KeywordFrom},
		{"घटक", KeywordComponent}, {"अवयव", KeywordComponent}, {"प्रकार", KeywordType}, {"केवल_पढ़ने", KeywordReadonly}, {"एनम", KeywordEnum}, {"इंटरफ़ेस", KeywordInterface}, {"विस्तार", KeywordExtends},
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
