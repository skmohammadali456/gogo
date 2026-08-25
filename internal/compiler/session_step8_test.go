package compiler

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestStep8EnglishBengaliHindiSessionIsolation(t *testing.T) {
	enSession := NewSession(WithGrammarLanguage(grammar.English))
	bnSession := NewSession(WithGrammarLanguage(grammar.Bengali))
	hiSession := NewSession(WithGrammarLanguage(grammar.Hindi))
	enID := enSession.AddFile("en.gogo", "create variable बनाओ as \"identifier\"")
	bnID := bnSession.AddFile("bn.gogo", "তৈরি চলক उपयोगकर्ता হিসেবে \"Alex\"")
	hiID := hiSession.AddFile("hi.gogo", "बनाओ चर user रूप \"Alex\"")
	enFile, _ := enSession.ParseFile(enID)
	bnFile, _ := bnSession.ParseFile(bnID)
	hiFile, _ := hiSession.ParseFile(hiID)
	if enSession.Diagnostics.HasErrors() || bnSession.Diagnostics.HasErrors() || hiSession.Diagnostics.HasErrors() {
		t.Fatalf("session diagnostics en=%v bn=%v hi=%v", enSession.Diagnostics.All(), bnSession.Diagnostics.All(), hiSession.Diagnostics.All())
	}
	if enFile.Statements[0].(ast.VariableDecl).Name.Name != "बनाओ" {
		t.Fatalf("Hindi keyword should be an English-session identifier")
	}
	if bnFile.Statements[0].(ast.VariableDecl).Name.Name != "उपयोगकर्ता" {
		t.Fatalf("Hindi identifier should remain a Bengali-session identifier")
	}
	if hiFile.Statements[0].(ast.VariableDecl).Name.Name != "user" {
		t.Fatalf("ASCII identifier should remain a Hindi-session identifier")
	}
	if enSession.GrammarVocabulary().Language != grammar.English || bnSession.GrammarVocabulary().Language != grammar.Bengali || hiSession.GrammarVocabulary().Language != grammar.Hindi {
		t.Fatalf("sessions did not preserve independent vocabularies")
	}
}
