package compiler

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestStep7MixedLanguageSessionIsolation(t *testing.T) {
	enSession := NewSession(WithGrammarLanguage(grammar.English))
	bnSession := NewSession(WithGrammarLanguage(grammar.Bengali))
	enID := enSession.AddFile("en.gogo", "create variable তৈরি as \"identifier\"")
	bnID := bnSession.AddFile("bn.gogo", "তৈরি চলক user হিসেবে \"Alex\"")
	enFile, _ := enSession.ParseFile(enID)
	bnFile, _ := bnSession.ParseFile(bnID)
	if enSession.Diagnostics.HasErrors() || bnSession.Diagnostics.HasErrors() {
		t.Fatalf("session diagnostics en=%v bn=%v", enSession.Diagnostics.All(), bnSession.Diagnostics.All())
	}
	if enFile.Statements[0].(ast.VariableDecl).Name.Name != "তৈরি" {
		t.Fatalf("Bengali keyword should be an English-session identifier")
	}
	if bnFile.Statements[0].(ast.VariableDecl).Name.Name != "user" {
		t.Fatalf("ASCII identifier should remain a Bengali-session identifier")
	}
	if enSession.GrammarVocabulary().Language != grammar.English || bnSession.GrammarVocabulary().Language != grammar.Bengali {
		t.Fatalf("sessions did not preserve independent vocabularies")
	}
}
