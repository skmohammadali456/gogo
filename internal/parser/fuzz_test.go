package parser

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
)

func FuzzParseNeverPanics(f *testing.F) {
	for _, seed := range []string{
		"create variable user as \"Alex\"",
		"create variable x as @",
		"create function f(a, b) { return a + b }",
		"বাংলা নাম",
		"हिन्दी नाम",
		"[ { bad: , ok: 1 }",
		"/* unterminated",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		l := lexer.New(source.File{ID: 1, Path: "fuzz.gogo", Text: text})
		p := New(l.LexAll())
		_ = p.ParseFile()
		_ = p.Diagnostics()
	})
}
