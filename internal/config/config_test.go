package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skmohammadali786/gogo/internal/grammar"
)

func TestResolveDefaultsToEnglishUTF8ASTStep9(t *testing.T) {
	got, diags := Resolve(Raw{}, Overrides{})
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if got.Language != grammar.English || got.Strictness != StrictStandard || got.Encoding != EncodingUTF8 || got.Target != TargetAST || got.Compatibility != CompatibilityStep9 {
		t.Fatalf("bad defaults: %#v", got)
	}
	if _, ok := got.Vocabulary.Lookup("create"); !ok {
		t.Fatal("default vocabulary must be English")
	}
}

func TestResolveSelectsBuiltInLanguages(t *testing.T) {
	cases := []struct {
		in      string
		surface string
		lang    grammar.Language
	}{{"english", "create", grammar.English}, {"bengali", "তৈরি", grammar.Bengali}, {"hindi", "बनाओ", grammar.Hindi}}
	for _, tc := range cases {
		got, diags := Resolve(Raw{Language: tc.in}, Overrides{})
		if len(diags) != 0 {
			t.Fatalf("%s diagnostics: %#v", tc.in, diags)
		}
		if got.Language != tc.lang {
			t.Fatalf("language = %s got %s", tc.in, got.Language)
		}
		if _, ok := got.Vocabulary.Lookup(tc.surface); !ok {
			t.Fatalf("%s did not select %q", tc.in, tc.surface)
		}
	}
}

func TestResolveProjectAliasesMapToCanonicalKeywords(t *testing.T) {
	got, diags := Resolve(Raw{Aliases: []Alias{{Surface: "make", Keyword: "create"}, {Surface: "thing", Keyword: "variable"}}}, Overrides{})
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if kw, ok := got.Vocabulary.Lookup("make"); !ok || kw != grammar.KeywordCreate {
		t.Fatalf("bad alias: %v %v", kw, ok)
	}
}

func TestResolveRejectsInvalidConfiguration(t *testing.T) {
	_, diags := Resolve(Raw{Language: "xx", Strictness: "loose", Encoding: "latin1", Target: "web", Compatibility: "v0", Aliases: []Alias{{Surface: "", Keyword: "create"}, {Surface: "123", Keyword: "create"}, {Surface: "x", Keyword: "missing"}, {Surface: "create", Keyword: "return"}}}, Overrides{})
	if len(diags) < 8 {
		t.Fatalf("expected validation diagnostics, got %#v", diags)
	}
}

func TestResolveReportsCLIProjectLanguageConflict(t *testing.T) {
	_, diags := Resolve(Raw{Language: "bn"}, Overrides{Language: "hi"})
	if len(diags) == 0 || diags[0].Code != "G3008" {
		t.Fatalf("expected conflict diagnostic, got %#v", diags)
	}
}

func TestDiscoverFindsNearestNestedProjectConfig(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, FileName), []byte(`{"language":"bn"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", FileName), []byte(`{"language":"hi"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, ok := Discover(filepath.Join(nested, "main.gogo"))
	if !ok || got != filepath.Join(root, "a", FileName) {
		t.Fatalf("got %q %v", got, ok)
	}
}

func TestResolveRejectsDuplicateAliasAndReservedWordCollision(t *testing.T) {
	_, diags := Resolve(Raw{Aliases: []Alias{{Surface: "make", Keyword: "create"}, {Surface: "make", Keyword: "create"}, {Surface: "create", Keyword: "create"}}}, Overrides{})
	codes := map[string]bool{}
	for _, d := range diags {
		codes[d.Code] = true
	}
	if !codes["G3010"] || !codes["G3009"] {
		t.Fatalf("expected duplicate and collision diagnostics, got %#v", diags)
	}
}

func TestResolveDoesNotReportEquivalentLanguageSpellingsAsConflict(t *testing.T) {
	_, diags := Resolve(Raw{Language: "english"}, Overrides{Language: "en"})
	if len(diags) != 0 {
		t.Fatalf("equivalent language spellings should not conflict: %#v", diags)
	}
}

func TestLoadRejectsMalformedConfigurationFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	if err := os.WriteFile(path, []byte(`{"language":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected malformed configuration to fail")
	}
}
