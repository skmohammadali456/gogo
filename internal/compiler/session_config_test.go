package compiler

import (
	"sync"
	"testing"

	"github.com/skmohammadali786/gogo/internal/config"
	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestSessionConsumesResolvedConfiguration(t *testing.T) {
	resolved, diags := config.Resolve(config.Raw{Language: "bn"}, config.Overrides{})
	if len(diags) != 0 {
		t.Fatalf("config diagnostics: %#v", diags)
	}
	s := NewSession(WithResolvedConfig(resolved))
	if s.GrammarVocabulary().Language != grammar.Bengali {
		t.Fatalf("wrong vocabulary: %#v", s.GrammarVocabulary())
	}
	id := s.AddFile("main.gogo", "তৈরি চলক নাম হিসেবে \"আলেক্স\"")
	if _, ok := s.ParseFile(id); !ok || s.HasErrors() {
		t.Fatalf("Bengali session failed: %#v", s.Diagnostics.All())
	}
}

func TestMultipleConfiguredSessionsAreIndependent(t *testing.T) {
	cases := []config.Raw{{Language: "en"}, {Language: "bn"}, {Language: "hi"}}
	var wg sync.WaitGroup
	for _, raw := range cases {
		raw := raw
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, diags := config.Resolve(raw, config.Overrides{})
			if len(diags) != 0 {
				t.Errorf("diagnostics: %#v", diags)
				return
			}
			s := NewSession(WithResolvedConfig(r))
			if s.GrammarVocabulary().Language != r.Language {
				t.Errorf("session language = %s want %s", s.GrammarVocabulary().Language, r.Language)
			}
		}()
	}
	wg.Wait()
}

func TestConfiguredAliasDoesNotMutateBuiltInVocabularyOrOtherSessions(t *testing.T) {
	withAlias, diags := config.Resolve(config.Raw{Language: "en", Aliases: []config.Alias{{Surface: "make", Keyword: "create"}, {Surface: "thing", Keyword: "variable"}}}, config.Overrides{})
	if len(diags) != 0 {
		t.Fatalf("config diagnostics: %#v", diags)
	}
	withoutAlias, diags := config.Resolve(config.Raw{Language: "en"}, config.Overrides{})
	if len(diags) != 0 {
		t.Fatalf("config diagnostics: %#v", diags)
	}
	if _, ok := grammar.DefaultVocabulary().Lookup("make"); ok {
		t.Fatal("project alias mutated the global English vocabulary")
	}
	if _, ok := withoutAlias.Vocabulary.Lookup("make"); ok {
		t.Fatal("project alias leaked into another resolved configuration")
	}
	a := NewSession(WithResolvedConfig(withAlias))
	b := NewSession(WithResolvedConfig(withoutAlias))
	idA := a.AddFile("a.gogo", "make thing x as 1")
	a.ParseFile(idA)
	if a.HasErrors() {
		t.Fatalf("alias session should compile: %#v", a.Diagnostics.All())
	}
	idB := b.AddFile("b.gogo", "make thing x as 1")
	b.ParseFile(idB)
	if !b.HasErrors() {
		t.Fatal("session without alias must not recognize project alias")
	}
}

func TestWithGrammarLanguageKeepsResolvedConfigBackwardCompatible(t *testing.T) {
	s := NewSession(WithGrammarLanguage(grammar.Hindi))
	if s.GrammarVocabulary().Language != grammar.Hindi || s.Config.Language != grammar.Hindi {
		t.Fatalf("WithGrammarLanguage did not keep session config in sync: %#v", s.Config)
	}
}
