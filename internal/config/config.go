package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
	"github.com/skmohammadali786/gogo/internal/lexer"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

const FileName = "gogo.json"

type Strictness string

const (
	StrictStandard   Strictness = "standard"
	StrictStrict     Strictness = "strict"
	StrictPermissive Strictness = "permissive"
)

type Encoding string

const EncodingUTF8 Encoding = "utf-8"

type Target string

const TargetAST Target = "ast"

type Compatibility string

const CompatibilityStep9 Compatibility = "step9"

type Alias struct {
	Surface string `json:"surface"`
	Keyword string `json:"keyword"`
}
type Raw struct {
	Language      string  `json:"language,omitempty"`
	Aliases       []Alias `json:"aliases,omitempty"`
	Strictness    string  `json:"strictness,omitempty"`
	Encoding      string  `json:"encoding,omitempty"`
	Target        string  `json:"target,omitempty"`
	Compatibility string  `json:"compatibility,omitempty"`
}
type Overrides struct {
	Language   string
	Strictness string
	Target     string
	ConfigPath string
}

type Resolved struct {
	Language      grammar.Language
	Vocabulary    grammar.Vocabulary
	Aliases       map[string]grammar.Keyword
	Strictness    Strictness
	Encoding      Encoding
	Target        Target
	Compatibility Compatibility
	SourcePath    string
}

func Defaults() Raw {
	return Raw{Language: string(grammar.English), Strictness: string(StrictStandard), Encoding: string(EncodingUTF8), Target: string(TargetAST), Compatibility: string(CompatibilityStep9)}
}
func DefaultResolved() Resolved { r, _ := Resolve(Raw{}, Overrides{}); return r }

func Discover(start string) (string, bool) {
	dir := start
	if dir == "" {
		dir = "."
	}
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	dir, _ = filepath.Abs(dir)
	for {
		p := filepath.Join(dir, FileName)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
func Load(path string) (Raw, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Raw{}, err
	}
	var raw Raw
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		return Raw{}, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Raw{}, fmt.Errorf("configuration must contain exactly one JSON value")
		}
		return Raw{}, err
	}
	return raw, nil
}
func LoadDiscover(start string) (Raw, string, bool, error) {
	p, ok := Discover(start)
	if !ok {
		return Raw{}, "", false, nil
	}
	r, err := Load(p)
	return r, p, true, err
}

func Resolve(raw Raw, over Overrides) (Resolved, []diagnostics.Diagnostic) {
	d := []diagnostics.Diagnostic{}
	def := Defaults()
	lang := first(over.Language, raw.Language, def.Language)
	strict := first(over.Strictness, raw.Strictness, def.Strictness)
	enc := first(raw.Encoding, def.Encoding)
	target := first(over.Target, raw.Target, def.Target)
	compat := first(raw.Compatibility, def.Compatibility)
	if over.Language != "" && raw.Language != "" {
		cliLanguage, cliOK := parseLanguage(over.Language)
		projectLanguage, projectOK := parseLanguage(raw.Language)
		if cliOK && projectOK && cliLanguage != projectLanguage {
			d = append(d, diag("G3008", "language", fmt.Sprintf("Conflicting language settings: CLI %q overrides project %q.", over.Language, raw.Language), "Use one language setting, or make the CLI and project configuration agree."))
		}
	}
	language, ok := parseLanguage(lang)
	if !ok {
		d = append(d, diag("G3001", "language", fmt.Sprintf("Unknown language %q in project configuration.", lang), "Use en, english, bn, bengali, hi, or hindi."))
	}
	st, ok := parseStrict(strict)
	if !ok {
		d = append(d, diag("G3002", "strictness", fmt.Sprintf("Unknown strictness mode %q.", strict), "Use standard, strict, or permissive."))
	}
	encoding := Encoding(strings.ToLower(enc))
	if encoding != "utf8" && encoding != EncodingUTF8 {
		d = append(d, diag("G3003", "encoding", fmt.Sprintf("Unsupported source encoding %q.", enc), "GOGO source files must be UTF-8."))
	}
	encoding = EncodingUTF8
	tg := Target(strings.ToLower(target))
	if tg != TargetAST {
		d = append(d, diag("G3004", "target", fmt.Sprintf("Unknown target %q.", target), "Step 9 supports the ast target only."))
	}
	cp := Compatibility(strings.ToLower(compat))
	if cp != CompatibilityStep9 {
		d = append(d, diag("G3005", "compatibility", fmt.Sprintf("Unsupported compatibility setting %q.", compat), "Use step9 for the current language-configuration contract."))
	}
	vocab := grammar.DefaultVocabulary()
	if language != "" {
		if v, err := grammar.ForLanguage(language); err == nil {
			vocab = v
		}
	}
	aliases := map[string]grammar.Keyword{}
	entries := vocab.Entries()
	for _, a := range raw.Aliases {
		surf := strings.TrimSpace(a.Surface)
		if surf == "" {
			d = append(d, diag("G3006", "aliases", "Alias surface cannot be empty.", "Give every alias a non-empty identifier spelling."))
			continue
		}
		if !isIdentifier(surf) {
			d = append(d, diag("G3006", "aliases", fmt.Sprintf("Alias %q is not a valid GOGO identifier.", surf), "Aliases must lex as exactly one identifier token."))
			continue
		}
		kw, ok := grammar.ParseKeyword(a.Keyword)
		if !ok {
			d = append(d, diag("G3007", "aliases", fmt.Sprintf("Alias %q maps to unknown keyword %q.", surf, a.Keyword), "Map aliases to an existing canonical keyword such as variable or return."))
			continue
		}
		if _, ok := aliases[surf]; ok {
			d = append(d, diag("G3010", "aliases", fmt.Sprintf("Alias %q is defined more than once.", surf), "Define each alias once."))
			continue
		}
		if old, ok := entries[surf]; ok {
			d = append(d, diag("G3009", "aliases", fmt.Sprintf("Alias %q collides with reserved keyword %s.", surf, old), "Choose a spelling that is not reserved by the selected language."))
			continue
		}
		aliases[surf] = kw
		entries[surf] = kw
	}
	if len(d) > 0 {
		return Resolved{}, d
	}
	merged := make([]grammar.Entry, 0, len(entries))
	for s, k := range entries {
		merged = append(merged, grammar.Entry{Surface: s, Keyword: k})
	}
	return Resolved{Language: language, Vocabulary: grammar.NewVocabulary(language, vocab.Name, merged), Aliases: aliases, Strictness: st, Encoding: encoding, Target: tg, Compatibility: cp, SourcePath: over.ConfigPath}, nil
}
func first(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
func parseLanguage(s string) (grammar.Language, bool) {
	switch strings.ToLower(s) {
	case "en", "english":
		return grammar.English, true
	case "bn", "bengali":
		return grammar.Bengali, true
	case "hi", "hindi":
		return grammar.Hindi, true
	default:
		return "", false
	}
}
func parseStrict(s string) (Strictness, bool) {
	switch Strictness(strings.ToLower(s)) {
	case StrictStandard, StrictStrict, StrictPermissive:
		return Strictness(strings.ToLower(s)), true
	default:
		return "", false
	}
}
func isIdentifier(s string) bool {
	f := source.File{ID: 1, Path: "alias.gogo", Text: s}
	lx := lexer.New(f)
	toks := lx.LexAll()
	return len(lx.Diagnostics()) == 0 && len(toks) == 2 && toks[0].Kind == token.Identifier && toks[0].Text == s
}
func diag(code, field, msg, hint string) diagnostics.Diagnostic {
	return diagnostics.Diagnostic{Severity: diagnostics.Error, Code: code, Message: msg, Hint: field + ": " + hint, Span: source.Span{Start: source.NewPosition(), End: source.NewPosition()}}
}
