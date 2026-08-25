package compiler

import (
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

// Session contains state shared by compiler phases for one compilation request.
type Session struct {
	Files       *source.FileMap
	Diagnostics diagnostics.Bag
	vocabulary  grammar.Vocabulary
}

// Option configures a compiler session.
type Option func(*Session)

// WithGrammarLanguage selects the active grammar vocabulary for this session.
func WithGrammarLanguage(language grammar.Language) Option {
	return func(s *Session) {
		if v, err := grammar.ForLanguage(language); err == nil {
			s.vocabulary = v
		}
	}
}

// WithGrammarVocabulary selects an explicit active grammar vocabulary for this session.
func WithGrammarVocabulary(v grammar.Vocabulary) Option {
	return func(s *Session) { s.vocabulary = v }
}

func NewSession(opts ...Option) *Session {
	s := &Session{Files: source.NewFileMap(), vocabulary: grammar.DefaultVocabulary()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Session) GrammarVocabulary() grammar.Vocabulary { return s.vocabulary }

func (s *Session) AddFile(path, text string) uint32 {
	if !source.ValidatePath(path) {
		s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G0001", Message: "I need a source file path before I can compile this file.", Hint: "Give the source file a non-empty path, such as main.gogo.", FilePath: path, Span: source.Span{Start: source.NewPosition(), End: source.NewPosition()}})
		return 0
	}
	if !source.ValidateFile(source.File{ID: 1, Path: path, Text: text}) {
		s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G0002", Message: "This source file is not valid UTF-8 or has invalid source metadata.", Hint: "Save the file as UTF-8 and make sure its path is valid.", FilePath: path, Span: source.Span{Start: source.NewPosition(), End: source.NewPosition()}})
		return 0
	}
	return s.Files.Add(path, text)
}

func (s *Session) LexFile(id uint32) []token.Token {
	file, ok := s.Files.Get(id)
	if !ok {
		s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G0003", Message: "The compiler cannot lex a source file that is not in this session.", Hint: "Add the file to the compilation session before lexing it.", Span: source.Span{Start: source.NewPosition(), End: source.NewPosition()}})
		return nil
	}
	l := lexer.New(file)
	tokens := l.LexAll()
	for _, diagnostic := range l.Diagnostics() {
		diagnostic.FileID = id
		s.Diagnostics.Add(diagnostic)
	}
	return tokens
}

func (s *Session) HasErrors() bool { return s.Diagnostics.HasErrors() }
