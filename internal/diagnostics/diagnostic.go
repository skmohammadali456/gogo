package diagnostics

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/skmohammadali786/gogo/internal/source"
)

type Severity uint8

const (
	Info Severity = iota
	Warning
	Error
	Hint
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Error:
		return "error"
	case Hint:
		return "hint"
	default:
		return "unknown"
	}
}

type Locale string

const (
	English Locale = "en"
	Bengali Locale = "bn"
	Hindi   Locale = "hi"
)

type LabelStyle string

const (
	Primary   LabelStyle = "primary"
	Secondary LabelStyle = "secondary"
)

type Label struct {
	Style    LabelStyle  `json:"style"`
	FileID   uint32      `json:"file_id,omitempty"`
	FilePath string      `json:"file_path,omitempty"`
	Span     source.Span `json:"span"`
	Message  string      `json:"message,omitempty"`
}

type FixIt struct {
	FileID      uint32      `json:"file_id,omitempty"`
	FilePath    string      `json:"file_path,omitempty"`
	Span        source.Span `json:"span"`
	Replacement string      `json:"replacement"`
	Message     string      `json:"message,omitempty"`
}

type Suggestion struct {
	Message string  `json:"message"`
	Edits   []FixIt `json:"edits,omitempty"`
}

type Diagnostic struct {
	Severity    Severity     `json:"severity"`
	FileID      uint32       `json:"file_id,omitempty"`
	FilePath    string       `json:"file_path,omitempty"`
	Code        string       `json:"code"`
	Message     string       `json:"message"`
	Hint        string       `json:"hint,omitempty"`
	Span        source.Span  `json:"span"`
	Labels      []Label      `json:"labels,omitempty"`
	Notes       []string     `json:"notes,omitempty"`
	Hints       []string     `json:"hints,omitempty"`
	Suggestions []Suggestion `json:"suggestions,omitempty"`
}

func (d Diagnostic) normalized() Diagnostic {
	if len(d.Labels) == 0 && d.Span.IsValid() {
		d.Labels = []Label{{Style: Primary, FileID: d.FileID, FilePath: d.FilePath, Span: d.Span}}
	}
	for i := range d.Labels {
		if d.Labels[i].FileID == 0 {
			d.Labels[i].FileID = d.FileID
		}
		if d.Labels[i].FilePath == "" {
			d.Labels[i].FilePath = d.FilePath
		}
	}
	for i := range d.Suggestions {
		for j := range d.Suggestions[i].Edits {
			if d.Suggestions[i].Edits[j].FileID == 0 {
				d.Suggestions[i].Edits[j].FileID = d.FileID
			}
			if d.Suggestions[i].Edits[j].FilePath == "" {
				d.Suggestions[i].Edits[j].FilePath = d.FilePath
			}
		}
	}
	if d.Hint != "" && len(d.Hints) == 0 {
		d.Hints = []string{d.Hint}
	}
	return d
}

type Bag struct{ items []Diagnostic }

func (b *Bag) Add(d Diagnostic) { b.items = append(b.items, d.normalized()) }

func (b *Bag) HasErrors() bool {
	for _, d := range b.items {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

func (b *Bag) All() []Diagnostic { return orderedDeduped(b.items) }

func orderedDeduped(in []Diagnostic) []Diagnostic {
	out := make([]Diagnostic, 0, len(in))
	seen := map[string]bool{}
	for _, d := range in {
		d = d.normalized()
		key := fmt.Sprintf("%d|%s|%s|%d|%d|%d|%s", d.FileID, d.FilePath, d.Code, d.Span.Start.Offset, d.Span.End.Offset, d.Severity, d.Message)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, d)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.FileID != b.FileID {
			return a.FileID < b.FileID
		}
		if a.FilePath != b.FilePath {
			return a.FilePath < b.FilePath
		}
		if a.Span.Start.Offset != b.Span.Start.Offset {
			return a.Span.Start.Offset < b.Span.Start.Offset
		}
		if a.Span.End.Offset != b.Span.End.Offset {
			return a.Span.End.Offset < b.Span.End.Offset
		}
		return a.Code < b.Code
	})
	return out
}

type Renderer struct {
	Files  *source.FileMap
	Locale Locale
	Color  bool
}

func (r Renderer) Text(diags []Diagnostic) string {
	var b strings.Builder
	for i, d := range orderedDeduped(diags) {
		if i > 0 {
			b.WriteByte('\n')
		}
		d = translate(d.normalized(), r.Locale)
		fmt.Fprintf(&b, "%s[%s]: %s\n", d.Severity.String(), d.Code, d.Message)
		path, text := r.fileText(d.FileID, d.FilePath)
		fmt.Fprintf(&b, " --> %s:%d:%d\n", path, d.Span.Start.Line, d.Span.Start.Column)
		b.WriteString(snippet(text, d))
		for _, label := range d.Labels {
			if label.Style == Primary && label.Message != "" {
				fmt.Fprintf(&b, "primary: %s at %d:%d\n", label.Message, label.Span.Start.Line, label.Span.Start.Column)
			}
		}
		for _, label := range d.Labels {
			if label.Style == Secondary && label.Message != "" {
				fmt.Fprintf(&b, "secondary: %s at %d:%d\n", label.Message, label.Span.Start.Line, label.Span.Start.Column)
			}
		}
		for _, n := range d.Notes {
			fmt.Fprintf(&b, "note: %s\n", n)
		}
		for _, h := range d.Hints {
			fmt.Fprintf(&b, "hint: %s\n", h)
		}
		for _, s := range d.Suggestions {
			fmt.Fprintf(&b, "suggestion: %s\n", s.Message)
		}
	}
	return b.String()
}

type JSONPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type JSONSpan struct {
	Start JSONPosition `json:"start"`
	End   JSONPosition `json:"end"`
}

type JSONLabel struct {
	Style   LabelStyle `json:"style"`
	File    string     `json:"file,omitempty"`
	Span    JSONSpan   `json:"span"`
	Message string     `json:"message,omitempty"`
}

type JSONFixIt struct {
	File        string   `json:"file,omitempty"`
	Span        JSONSpan `json:"span"`
	Replacement string   `json:"replacement"`
	Message     string   `json:"message,omitempty"`
}

type JSONSuggestion struct {
	Message string      `json:"message"`
	Edits   []JSONFixIt `json:"edits,omitempty"`
}

type JSONDiagnostic struct {
	Code        string           `json:"code"`
	Severity    string           `json:"severity"`
	Language    Locale           `json:"language"`
	Message     string           `json:"message"`
	File        string           `json:"file,omitempty"`
	Span        JSONSpan         `json:"span"`
	Labels      []JSONLabel      `json:"labels,omitempty"`
	Notes       []string         `json:"notes,omitempty"`
	Hints       []string         `json:"hints,omitempty"`
	Suggestions []JSONSuggestion `json:"suggestions,omitempty"`
}

func (r Renderer) JSON(diags []Diagnostic) ([]byte, error) {
	out := orderedDeduped(diags)
	schema := make([]JSONDiagnostic, 0, len(out))
	for _, d := range out {
		d = translate(d.normalized(), r.Locale)
		file, _ := r.fileText(d.FileID, d.FilePath)
		schema = append(schema, r.jsonDiagnostic(d, file))
	}
	return json.MarshalIndent(schema, "", "  ")
}

func (r Renderer) jsonDiagnostic(d Diagnostic, file string) JSONDiagnostic {
	labels := make([]JSONLabel, 0, len(d.Labels))
	for _, label := range d.Labels {
		labelFile, _ := r.fileText(label.FileID, label.FilePath)
		labels = append(labels, JSONLabel{Style: label.Style, File: labelFile, Span: jsonSpan(label.Span), Message: label.Message})
	}
	suggestions := make([]JSONSuggestion, 0, len(d.Suggestions))
	for _, suggestion := range d.Suggestions {
		edits := make([]JSONFixIt, 0, len(suggestion.Edits))
		for _, edit := range suggestion.Edits {
			editFile, _ := r.fileText(edit.FileID, edit.FilePath)
			edits = append(edits, JSONFixIt{File: editFile, Span: jsonSpan(edit.Span), Replacement: edit.Replacement, Message: edit.Message})
		}
		suggestions = append(suggestions, JSONSuggestion{Message: suggestion.Message, Edits: edits})
	}
	return JSONDiagnostic{Code: d.Code, Severity: d.Severity.String(), Language: localeOrEnglish(r.Locale), Message: d.Message, File: file, Span: jsonSpan(d.Span), Labels: labels, Notes: d.Notes, Hints: d.Hints, Suggestions: suggestions}
}

func jsonSpan(span source.Span) JSONSpan {
	return JSONSpan{Start: jsonPosition(span.Start), End: jsonPosition(span.End)}
}

func jsonPosition(pos source.Position) JSONPosition {
	return JSONPosition{Offset: pos.Offset, Line: pos.Line, Column: pos.Column}
}

func localeOrEnglish(locale Locale) Locale {
	if locale == "" {
		return English
	}
	return locale
}

func (r Renderer) fileText(id uint32, fallbackPath string) (string, string) {
	if r.Files == nil {
		if fallbackPath != "" {
			return fallbackPath, ""
		}
		return "<unknown>", ""
	}
	if id != 0 {
		if f, ok := r.Files.Get(id); ok {
			return f.Path, f.Text
		}
	}
	if fallbackPath != "" {
		return fallbackPath, ""
	}
	if f, ok := r.Files.Get(1); ok {
		return f.Path, f.Text
	}
	return "<unknown>", ""
}

func snippet(text string, d Diagnostic) string {
	if text == "" || !d.Span.IsValid() {
		return ""
	}
	lines := strings.SplitAfter(text, "\n")
	start, end := d.Span.Start.Line, d.Span.End.Line
	if end < start {
		end = start
	}
	var b strings.Builder
	for lineNo := start; lineNo <= end && lineNo <= len(lines); lineNo++ {
		line := strings.TrimSuffix(strings.TrimSuffix(lines[lineNo-1], "\n"), "\r")
		fmt.Fprintf(&b, "%4d | %s\n", lineNo, line)
		from, to := 1, utf8.RuneCountInString(line)+1
		if lineNo == start {
			from = d.Span.Start.Column
		}
		if lineNo == end {
			to = d.Span.End.Column
		}
		if to <= from {
			to = from + 1
		}
		fmt.Fprintf(&b, "     | %s%s\n", strings.Repeat(" ", max(0, from-1)), strings.Repeat("^", max(1, to-from)))
	}
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func translate(d Diagnostic, l Locale) Diagnostic {
	if l == English || l == "" {
		return d
	}
	prefix := map[Locale]string{Bengali: "বাংলা: ", Hindi: "हिन्दी: "}[l]
	if prefix == "" {
		return d
	}
	d.Message = lookupText(d.Code, d.Message, l, false)
	for i := range d.Hints {
		d.Hints[i] = lookupText(d.Code, d.Hints[i], l, true)
	}
	if d.Hint != "" {
		d.Hint = lookupText(d.Code, d.Hint, l, true)
	}
	return d
}

func (s Severity) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }
